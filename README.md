> ## Apiinfra
>
> Code generation **and** minimal runtime support for Go gRPC + grpc-gateway APIs.
>

# Apiinfra

Two-stage support for building Go API services:

1. **Generation** (`./genlib`) — Declare resources and operations in Go; emit `.proto`, `buf` config, and `.pb.go` outputs.
2. **Runtime** (`./pkg`) — Minimal helpers consumed by the running service: config loading and HTTP/gRPC listeners.

## Structure

```
genlib/   # build-time code that generates other code (proto, buf, server skeletons)
pkg/      # runtime code linked into the consumer's service binary
examples/ # end-to-end example using both halves
```

### `genlib/` — build-time

Imported by a consumer's `gen/...` package and run via `go generate`. Produces `.proto` and `buf` config from Go declarations. Panics on error via `genlib.Check` are acceptable here — this code only runs at build time. See `genlib/grpc/` for the `ServiceBuilder` + `Operation` abstraction, and `genlib/grpc/crud/` for the CRUD op family.

### `pkg/` — runtime

Imported by the consumer's runtime binary. Kept intentionally small so consumers can mix-and-match. No `panic`-on-error for caller mistakes — return errors and let the caller decide.

- `pkg/config` — `NewConfig(paths ...string) cs.Config` loads YAML/JSON files (plus `CONFIG_PATHS` env) with environment as a late-binding override. Built on [`activatedio/cs`](https://github.com/activatedio/cs).
- `pkg/gateway` — gRPC and grpc-gateway listener primitives:
  - `ServerConfig` — cs-loaded listener config (host/port/TLS/mTLS/cert paths) at the conventional cs prefix `"server"`.
  - `ProvideServer(opts...) fx.Option` — gRPC + JSON gateway on a single port.
  - `ProvideGrpcServer(opts...) fx.Option` — gRPC only (no HTTP gateway).
  - `WithMTLS(MTLSMode)` — compile-time mTLS mode (Disabled, FromConfig, Always).
- `pkg/service` — placeholder for future runtime helpers shared across services.

A consumer project keeps domain logic, generation entry points, FX index modules, and `main` packages in its own tree; everything cross-cutting and non-domain-specific belongs here.

## End-to-end example

See `examples/app/`:

- `gen/api/main.go` — generator entry point (runs `genlib`)
- `proto/` — generated `.proto` + buf config
- `pb/` — buf-produced `.pb.go`
- `server/` — hand-written service stubs

The example doesn't currently wire `pkg/gateway` (it just exercises generation); consumer projects do.

## Tests

Run `go test ./...` from the repo root.
