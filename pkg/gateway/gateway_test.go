package gateway_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/activatedio/apiinfra/pkg/gateway"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// providers returns the fx.Options that satisfy Params/GrpcParams
// with no-op registrations and the supplied ServerConfig.
func providers(cfg *gateway.ServerConfig) []fx.Option {
	return []fx.Option{
		fx.Supply(cfg, gateway.Config{Title: "test"}),
		fx.Provide(
			func() gateway.RegistrationFunc {
				return func(_ *grpc.Server) {}
			},
			func() gateway.GatewayFunc {
				return func(_ context.Context, _ *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
					return nil
				}
			},
		),
	}
}

func TestProvideServer_HealthRoute(t *testing.T) {
	cfg := &gateway.ServerConfig{Host: "127.0.0.1", Port: 0}

	var rs *gateway.RunningServer
	opts := append(providers(cfg),
		gateway.ProvideServer(),
		fx.Populate(&rs),
	)
	app := fxtest.New(t, opts...)
	app.RequireStart()
	defer app.RequireStop()

	require.NotNil(t, rs)
	url := fmt.Sprintf("http://127.0.0.1:%d/health", rs.Port)
	resp := getWithRetry(t, url)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "SERVING", string(body))
}

func TestProvideServer_HomePageRoute(t *testing.T) {
	cfg := &gateway.ServerConfig{Host: "127.0.0.1", Port: 0}

	var rs *gateway.RunningServer
	opts := append(providers(cfg),
		gateway.ProvideServer(),
		fx.Populate(&rs),
	)
	app := fxtest.New(t, opts...)
	app.RequireStart()
	defer app.RequireStop()

	resp := getWithRetry(t, fmt.Sprintf("http://127.0.0.1:%d/", rs.Port))
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "test")
}

func TestProvideGrpcServer_HealthCheck(t *testing.T) {
	cfg := &gateway.ServerConfig{Host: "127.0.0.1", Port: 0}

	var rs *gateway.RunningServer
	app := fxtest.New(t,
		fx.Supply(cfg),
		fx.Provide(func() gateway.RegistrationFunc { return func(_ *grpc.Server) {} }),
		gateway.ProvideGrpcServer(),
		fx.Populate(&rs),
	)
	app.RequireStart()
	defer app.RequireStop()

	addr := fmt.Sprintf("127.0.0.1:%d", rs.Port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	hc := healthpb.NewHealthClient(conn)
	resp, err := hc.Check(ctx, &healthpb.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.Status)
}

// getWithRetry retries the GET briefly to absorb the listener
// goroutine startup race between fx OnStart returning and the
// listener calling Accept.
func getWithRetry(t *testing.T, url string) *http.Response {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(url) //nolint:gosec // test localhost URL
		if err == nil {
			return resp
		}
		lastErr = err
		time.Sleep(20 * time.Millisecond)
	}
	require.NoError(t, lastErr)
	return nil
}
