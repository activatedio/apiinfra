package gateway

import (
	"context"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// GrpcParams holds fx-injected dependencies for the gRPC-only
// listener. GatewayFunc and gateway-only Config are intentionally
// absent; the runner that picks this mode does not need them.
type GrpcParams struct {
	fx.In
	Server       *ServerConfig
	Registration RegistrationFunc
}

// ProvideGrpcServer returns an fx.Option that wires a gRPC-only
// listener (no HTTP gateway, no aux routes). mTLS mode is bound
// here at fx-wiring time via WithMTLS.
func ProvideGrpcServer(opts ...ServerOpt) fx.Option {
	o := serverOpts{}
	for _, fn := range opts {
		fn(&o)
	}
	return fx.Provide(func(lc fx.Lifecycle, params GrpcParams) *RunningServer {
		return newGrpcServer(lc, params, o)
	})
}

func newGrpcServer(lc fx.Lifecycle, params GrpcParams, o serverOpts) *RunningServer {
	wantMTLS := resolveMTLS(o.mtls, params.Server)

	listener, err := net.Listen("tcp", params.Server.Addr())
	if err != nil {
		panic(fmt.Errorf("gateway: bind %s: %w", params.Server.Addr(), err))
	}
	port := listener.Addr().(*net.TCPAddr).Port

	var grpcOpts []grpc.ServerOption
	if params.Server.TLS {
		creds, err := serverCreds(params.Server, wantMTLS)
		if err != nil {
			panic(fmt.Errorf("gateway: build grpc creds: %w", err))
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(grpcOpts...)
	grpc_health_v1.RegisterHealthServer(grpcServer, grpchealth.NewServer())
	params.Registration(grpcServer)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				log.Info().Str("addr", listener.Addr().String()).Bool("tls", params.Server.TLS).Bool("mtls", wantMTLS).Msg("starting gRPC listener")
				if err := grpcServer.Serve(listener); err != nil {
					log.Error().Err(err).Msg("gRPC listener exited")
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			grpcServer.GracefulStop()
			return nil
		},
	})

	return &RunningServer{Port: port}
}
