// Package gateway provides the HTTP+gRPC and gRPC-only listeners
// used by api services built on top of apiinfra.
//
// The consumer's runner picks the listener mode at fx-wiring time:
//
//   - ProvideServer     — gRPC + grpc-gateway JSON shim on one port
//   - ProvideGrpcServer — gRPC only (no JSON, no aux routes)
//
// Both accept ServerOpts; WithMTLS configures whether the listener
// verifies client certificates. mTLS has three compile-time modes:
//
//   - MTLSFromConfig — runtime ServerConfig.MTLS gates verification (default)
//   - MTLSDisabled   — never check client certs
//   - MTLSAlways     — always on; panics if not configured at start
//
// ServerConfig is cs-loaded by the consumer (see pkg/config).
package gateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// ServerConfig is the cs-loaded listener config. By convention the
// consumer roots it at config.PrefixServer.
type ServerConfig struct {
	// Host is the bind interface. Empty binds to all interfaces.
	Host string
	// Port is the TCP port. Zero binds to an ephemeral port.
	Port int
	// TLS enables HTTPS on the public listener. The loopback used
	// by the in-process gateway dial is always cleartext.
	TLS bool
	// MTLS is only consulted when the listener was built with
	// WithMTLS(MTLSFromConfig) (the default). Has no effect under
	// MTLSDisabled or MTLSAlways.
	MTLS bool
	// TLSCertPath is the server certificate (PEM).
	TLSCertPath string `key:"tlsCertPath"`
	// TLSKeyPath is the server private key (PEM).
	TLSKeyPath string `key:"tlsKeyPath"`
	// TLSCAPath is the CA bundle used to verify presented client
	// certificates when mTLS is active.
	TLSCAPath string `key:"tlsCaPath"`
}

// Addr returns the bind address in host:port form.
func (s *ServerConfig) Addr() string {
	p := s.Port
	if p < 0 {
		p = 0
	}
	return fmt.Sprintf("%s:%d", s.Host, p)
}

// Config carries the compile-time surface metadata used only by the
// gateway listener (title shown on the home page, OpenAPI spec).
type Config struct {
	// Title is shown on the home page.
	Title string
	// OpenAPIJSON is served at /openapi.json.
	OpenAPIJSON []byte
}

// RegistrationFunc registers a gRPC service implementation onto
// the process's grpc.Server. The fx module for each surface
// provides one.
type RegistrationFunc func(s *grpc.Server)

// Func registers a service's grpc-gateway handler. Only consulted
// by the gateway listener; the gRPC-only listener ignores it.
type Func func(ctx context.Context, mux *runtime.ServeMux, target string, opts []grpc.DialOption) error

// readHeaderTimeout protects the public HTTP listener from
// Slowloris-style attacks by capping how long clients can take to
// send their request headers.
const readHeaderTimeout = 30 * time.Second

// MTLSMode picks how (or whether) the listener verifies client
// certs. Bound at fx-wiring time via WithMTLS.
type MTLSMode int

const (
	// MTLSFromConfig consults ServerConfig.MTLS at startup; if
	// false, behaves like MTLSDisabled. Zero-value default, so a
	// runner that calls ProvideServer() with no options gets
	// runtime-config-driven mTLS.
	MTLSFromConfig MTLSMode = iota
	// MTLSDisabled never verifies client certs.
	MTLSDisabled
	// MTLSAlways always verifies client certs; panics at startup
	// if TLS or TLSCAPath isn't configured.
	MTLSAlways
)

// ServerOpt configures a listener at fx-wiring time.
type ServerOpt func(*serverOpts)

type serverOpts struct {
	mtls MTLSMode
}

// WithMTLS picks the mTLS mode for the listener.
func WithMTLS(mode MTLSMode) ServerOpt {
	return func(o *serverOpts) { o.mtls = mode }
}

// Params holds fx-injected dependencies for the gateway listener.
//
// UnaryInterceptors, StreamInterceptors, and HTTPMiddleware are
// optional: a consumer that wants interceptors provides a single
// fx.Provide returning the chain-ordered slice. Apiinfra is
// deliberately opinion-free here — no logging/error/transaction
// defaults are injected. Empty slices behave as if the field were
// absent.
//
// Interceptor order is "first is outermost" — i.e. the first
// element runs before the next on the way in and after it on the
// way out. Same convention for HTTPMiddleware.
type Params struct {
	fx.In
	Server             *ServerConfig
	Config             Config
	Registration       RegistrationFunc
	Gateway            Func
	UnaryInterceptors  []grpc.UnaryServerInterceptor     `optional:"true"`
	StreamInterceptors []grpc.StreamServerInterceptor    `optional:"true"`
	HTTPMiddleware     []func(http.Handler) http.Handler `optional:"true"`
}

// RunningServer is what downstream fx invokes depend on so they
// only run once the listener is bound.
type RunningServer struct {
	Port int
}

// ProvideServer returns an fx.Option that wires the combined gRPC
// + JSON gateway listener. Runner mode and mTLS mode are bound
// here.
func ProvideServer(opts ...ServerOpt) fx.Option {
	o := serverOpts{}
	for _, fn := range opts {
		fn(&o)
	}
	return fx.Provide(func(lc fx.Lifecycle, params Params) *RunningServer {
		return newServer(lc, params, o)
	})
}

func newServer(lc fx.Lifecycle, params Params, o serverOpts) *RunningServer {
	wantMTLS := resolveMTLS(o.mtls, params.Server)

	var listenCfg net.ListenConfig
	listener, err := listenCfg.Listen(context.Background(), "tcp", params.Server.Addr())
	if err != nil {
		panic(fmt.Errorf("gateway: bind %s: %w", params.Server.Addr(), err))
	}
	publicPort := listener.Addr().(*net.TCPAddr).Port

	loopback, err := listenCfg.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Errorf("gateway: bind loopback: %w", err))
	}
	loopbackPort := loopback.Addr().(*net.TCPAddr).Port

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(params.UnaryInterceptors...),
		grpc.ChainStreamInterceptor(params.StreamInterceptors...),
	)
	grpc_health_v1.RegisterHealthServer(grpcServer, &healthServer{})
	params.Registration(grpcServer)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mux := runtime.NewServeMux(
				runtime.WithMarshalerOption("*", &runtime.JSONPb{
					MarshalOptions:   protojson.MarshalOptions{UseProtoNames: true},
					UnmarshalOptions: protojson.UnmarshalOptions{DiscardUnknown: true},
				}),
			)
			target := fmt.Sprintf("127.0.0.1:%d", loopbackPort)
			dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
			if err := params.Gateway(ctx, mux, target, dialOpts); err != nil {
				return fmt.Errorf("gateway register: %w", err)
			}

			attachAuxRoutes(mux, params.Config)

			dispatch := composeMiddleware(grpcDispatch(grpcServer, mux), params.HTTPMiddleware)

			go func() {
				log.Info().Str("addr", loopback.Addr().String()).Msg("starting internal gRPC loopback")
				if err := grpcServer.Serve(loopback); err != nil {
					log.Error().Err(err).Msg("internal gRPC loopback exited")
				}
			}()
			go func() {
				log.Info().Str("addr", listener.Addr().String()).Bool("tls", params.Server.TLS).Bool("mtls", wantMTLS).Msg("starting public HTTP/gRPC listener")
				if err := servePublic(listener, dispatch, params.Server, wantMTLS); err != nil {
					log.Error().Err(err).Msg("public listener exited")
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			grpcServer.GracefulStop()
			return nil
		},
	})

	return &RunningServer{Port: publicPort}
}

// servePublic runs the public listener with optional TLS / mTLS.
// When TLS is off the listener uses h2c so HTTP/2 gRPC works
// without TLS.
func servePublic(listener net.Listener, dispatch http.Handler, cfg *ServerConfig, mtls bool) error {
	if !cfg.TLS {
		srv := &http.Server{
			Handler:           h2c.NewHandler(dispatch, &http2.Server{}),
			ReadHeaderTimeout: readHeaderTimeout,
		}
		return srv.Serve(listener)
	}
	tlsCfg, err := buildServerTLS(cfg, mtls)
	if err != nil {
		return fmt.Errorf("gateway: build tls: %w", err)
	}
	srv := &http.Server{
		Handler:           dispatch,
		TLSConfig:         tlsCfg,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	return srv.ServeTLS(listener, "", "")
}

// composeMiddleware folds the slice into a single http.Handler with
// middleware[0] outermost. An empty/nil slice is a no-op.
func composeMiddleware(h http.Handler, middleware []func(http.Handler) http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// resolveMTLS turns the compile-time mode + runtime config into a
// single boolean. Validates required config for MTLSAlways.
func resolveMTLS(mode MTLSMode, cfg *ServerConfig) bool {
	switch mode {
	case MTLSDisabled:
		return false
	case MTLSFromConfig:
		return cfg.MTLS
	case MTLSAlways:
		if !cfg.TLS {
			panic("gateway: MTLSAlways requires server.tls=true")
		}
		if cfg.TLSCAPath == "" {
			panic("gateway: MTLSAlways requires server.tlsCaPath")
		}
		return true
	default:
		panic(fmt.Errorf("gateway: unknown MTLSMode %d", mode))
	}
}

// buildServerTLS assembles a *tls.Config from ServerConfig. When
// mtls is true, TLSCAPath is loaded into ClientCAs and ClientAuth
// is set to RequireAndVerifyClientCert.
func buildServerTLS(cfg *ServerConfig, mtls bool) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.TLSCertPath, cfg.TLSKeyPath)
	if err != nil {
		return nil, err
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"h2", "http/1.1"},
	}
	if mtls {
		if cfg.TLSCAPath == "" {
			return nil, fmt.Errorf("mTLS requested but tlsCaPath is empty")
		}
		caPEM, err := os.ReadFile(cfg.TLSCAPath)
		if err != nil {
			return nil, fmt.Errorf("read tlsCaPath: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("tlsCaPath: no PEM blocks parsed")
		}
		tlsCfg.ClientCAs = pool
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsCfg, nil
}

// serverCreds wraps buildServerTLS for the gRPC-only listener.
func serverCreds(cfg *ServerConfig, mtls bool) (credentials.TransportCredentials, error) {
	tlsCfg, err := buildServerTLS(cfg, mtls)
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(tlsCfg), nil
}

// grpcDispatch routes HTTP/2 requests whose Content-Type indicates
// gRPC to the gRPC server, otherwise to the gateway/HTTP handler.
func grpcDispatch(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

// attachAuxRoutes mounts the utility routes the gateway always
// serves: home page, /openapi.json, and /health.
func attachAuxRoutes(mux *runtime.ServeMux, cfg Config) {
	_ = mux.HandlePath(http.MethodGet, "/openapi.json", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(cfg.OpenAPIJSON)
	})
	_ = mux.HandlePath(http.MethodGet, "/health", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("SERVING"))
	})
	_ = mux.HandlePath(http.MethodGet, "/", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = homePageTmpl.Execute(w, cfg)
	})
}

var homePageTmpl = template.Must(template.New("home").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{{ .Title }}</title>
<style>
  body { font: 16px/1.5 system-ui, sans-serif; max-width: 640px; margin: 4em auto; padding: 0 1em; color: #222; }
  h1 { font-size: 1.4em; margin-bottom: 0.25em; }
  h2 { font-size: 1.1em; margin-top: 1.5em; }
  ul { padding-left: 1.2em; }
  code { background: #f3f3f3; padding: 0.1em 0.3em; border-radius: 3px; }
</style>
</head>
<body>
<h1>{{ .Title }}</h1>
<p>gRPC + HTTP/JSON service. The OpenAPI spec describes every RPC the gateway proxies.</p>

<h2>OpenAPI spec</h2>
<ul>
  <li><a href="/openapi.json">/openapi.json</a></li>
</ul>

<h2>Health</h2>
<ul>
  <li><a href="/health">/health</a></li>
</ul>
</body>
</html>
`))

type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (h *healthServer) Check(_ context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}
