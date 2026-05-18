# apiinfra

Two halves for building Go gRPC + grpc-gateway services:

1. **Generation** (`./genlib`) — Build-time code that emits `.proto`, buf config, and `.pb.go` from Go declarations.
2. **Runtime** (`./pkg`) — Minimal runtime helpers consumed by the running service (config loading, HTTP/gRPC listeners).

Consumers (e.g. `gitlab.authwise.io/authwise/authwise-portal`) keep domain logic, gen entry points, fx Index modules, and `main` packages in their own tree; everything cross-cutting and non-domain-specific belongs here.

## Repo layout

```
genlib/             # build-time: code that generates other code
  grpc/
    operation.go        # core abstraction: Target, Operation, OperationFunc
    service_builder.go  # ServiceBuilder applies Operations to a Target
    util.go             # small protogen helpers
    crud/               # CRUD op family (Get, List, Create, Update, Patch, Delete)
    buf/                # renders buf.yaml, buf.gen.yaml, gen.go
pkg/                # runtime: linked into the consumer's service binary
  config/             # cs-based config loader (YAML/JSON + env)
  gateway/            # gRPC + grpc-gateway and gRPC-only fx listeners
  service/            # placeholder for future cross-service runtime helpers
examples/app/       # end-to-end example of the generator (does not wire pkg/gateway)
  gen/api/main.go       # generator entry point (go:generate target)
  proto/                # generated .proto + buf config + go:generate hook
  pb/                   # buf-produced .pb.go files
  server/               # hand-written service implementation stubs
```

### `genlib/` vs `pkg/` — what goes where

- `genlib/` is **build-time-only**. It runs from `gen/` packages in the consumer (or `examples/app/gen/api/main.go` here). Generator code can `panic` via `genlib.Check`/`CheckClose` — failure means a broken codegen run, not a runtime fault.
- `pkg/` is **linked into the running service binary**. It must not panic for caller mistakes — return errors and let the caller decide. The two exceptions in `pkg/gateway` are bind failures and `MTLSAlways` misconfiguration, which are startup-time invariants that should fail fast.

When adding a new feature, decide which half it belongs in by asking: does this run during `go generate`, or during `./service run`?

## Two generation stages

1. **Go → `.proto` + buf config.** Run `go run .` in `examples/app/gen/api`. The Go file `main.go` constructs a `grpc.ServiceBuilder`, adds CRUD operations and any custom ones, then writes `.proto` and buf files into `examples/app/proto/`.
2. **`.proto` → `.pb.go`.** Run `go generate ./...` in `examples/app/proto/`. That triggers buf to fetch deps and emit Go bindings into `examples/app/pb/`.

Both stages are exercised together by `go generate ./...` from `examples/app/`.

## Core abstraction: `Operation`

Everything that contributes to a service is an `Operation`:

```go
type Target struct {
    Messages       proto.File
    Service        proto.Service
    ServiceImports proto.File
    APIBasePath    string
}

type Operation interface { Apply(t Target) }
```

`ServiceBuilder` is just a runner that applies a sequence of `Operation`s to a `Target`. The CRUD methods live in `genlib/grpc/crud` and are one family of operations; future families (search, batch, etc.) live as sibling subpackages and return the same `Operation` type.

### CRUD usage

```go
sb := grpc.NewServiceBuilder(grpc.Target{
    Messages: msgFile, Service: svc, ServiceImports: svcFile, APIBasePath: "/v1",
})
sb.Add(
    crud.All(crud.Resource{Message: productMsg}),                // all six methods
    crud.All(crud.Resource{                                       // bitmask subset
        Message:    reviewMsg,
        ParentPath: "products/*",
        Ops:        crud.OpGet | crud.OpList | crud.OpDelete,
    }),
)
```

Zero `Ops` means `OpAll`. Per-op constructors (`crud.Get`, `crud.List`, …) are available when only one method is wanted.

### Escape hatch: arbitrary methods via protogen

Any caller can build methods directly with `protogen/proto` and pass them through `grpc.OperationFunc`. No new abstraction needed — protogen is the root.

```go
sb.Add(grpc.OperationFunc(func(t grpc.Target) {
    t.Service.AddMethods(
        proto.NewMethod("ArchiveProduct", proto.MethodParams{
            RequestName:  "ArchiveProductRequest",
            ResponseName: "example.api.Product",
        }).AddOptions(
            proto.NewOption("google.api.http", proto.NewMessageValueConstant(
                tfl.NewMessageValue().AddFields(
                    tfl.NewStringField("post", "/v1/{name=products/*}:archive"),
                    tfl.NewStringField("body", "*"),
                ),
            )),
        ),
    )
    t.Messages.AddMessages(/* request message */)
}))
```

Imports are deduped at the protogen `File` level, so it is safe for each operation (CRUD or custom) to register the imports it needs (`google/api/annotations.proto`, `google/protobuf/field_mask.proto`, …) without coordinating with siblings.

## API conventions (AIP-aligned)

The generated CRUD output targets Google AIP-132/134 conventions:

- **Resource names** are caller-provided `string name` in path bindings: `/{name=products/*}`, `/{name=tenants/*/products/*}`.
- **Collection URLs** are bare: `/products`, `/{parent=tenants/*}/products`. No trailing `/*` on list/create.
- **PatchRequest** carries `google.protobuf.FieldMask update_mask` (AIP-134), not `field_mask`.
- **ListResponse** uses the plural resource name as the list field: `repeated Product products`.
- **Get/List requests** use `google.protobuf.FieldMask read_mask`, not a `repeated string fields` workaround.
- **ListRequest** carries `page_size`, `page_token`, `filter`, `order_by`, `read_mask`.
- **DeleteRequest** returns `google.protobuf.Empty`.
- **Custom actions** follow `:verb` suffix style: `/v1/{name=products/*}:archive`.

When adding new op families or custom methods, follow these conventions unless there's a specific reason not to.

## Runtime packages (`pkg/`)

### `pkg/config`

`NewConfig(paths ...string) cs.Config` builds a `cs.Config` tree:
- file paths from `CONFIG_PATHS` env (comma-separated) come first
- explicit args next
- YAML (`.yaml`/`.yml`) and JSON (`.json`) are parsed; other extensions are ignored with a debug log
- a late-binding environment source is added last, so env vars override file values

`PrefixServer = "server"` is the conventional cs root for `gateway.ServerConfig`. Consumers can mount additional typed config at their own prefixes.

### `pkg/gateway`

Two fx provider factories — the runner picks one:

- `ProvideServer(opts ...ServerOpt)` — gRPC + grpc-gateway JSON shim on a single port. Uses an internal loopback gRPC listener that the in-process gateway dials.
- `ProvideGrpcServer(opts ...ServerOpt)` — gRPC only; no HTTP gateway, no aux routes.

Both consume `*ServerConfig` (cs-loaded at `"server"`) plus a `RegistrationFunc`; `ProvideServer` additionally needs `Config` (Title + OpenAPIJSON) and `GatewayFunc`. Both produce `*RunningServer` so downstream `fx.Invoke` can depend on the listener being bound.

mTLS is bound at fx-wiring time via `WithMTLS(MTLSMode)`:
- `MTLSDisabled` (default) — no client-cert verification.
- `MTLSFromConfig` — runtime `ServerConfig.MTLS` decides.
- `MTLSAlways` — always required; panics at startup if `TLS` or `TLSCAPath` is missing.

When `ServerConfig.TLS` is true, `ProvideServer` serves HTTPS (HTTP/2 ALPN auto-negotiated); when false it uses h2c (HTTP/2 cleartext) so gRPC still works.

### `pkg/service`

Currently empty — reserved for future runtime helpers shared across services.

## Available libraries for codegen

- `github.com/activatedio/protogen` — primary; produces `.proto`. Surface: `proto.NewFile`, `proto.NewMessage`, `proto.NewField`, `proto.NewService`, `proto.NewMethod`, `proto.NewImport`, `proto.NewOption`, plus `tfl` for HTTP option message-values.
- `github.com/dave/jennifer` (jen) — already imported via `genlib/util.go`; available when we need to generate Go source (e.g. server skeletons). Not currently used for `.proto`.
- `github.com/gertd/go-pluralize` — pluralizes resource names for AIP shapes.
- `github.com/iancoleman/strcase` — snake/camel conversions.

## Tests

- **Generator tests** are golden-string: build a `proto.Service`/`proto.File`, render to a buffer, compare to an expected string literal (see `genlib/grpc/crud/crud_test.go`). When changing generated output, update the golden strings deliberately — they ARE the spec for downstream consumers.
- The escape-hatch test in `crud_test.go` uses `assert.Contains` rather than full golden strings so the test stays robust against unrelated formatting drift.
- **Runtime tests** in `pkg/config` and `pkg/gateway` are conventional unit + smoke tests:
  - `pkg/config/config_test.go` round-trips YAML/JSON files and env overrides through `cs`.
  - `pkg/gateway/gateway_internal_test.go` covers `Addr()`, `resolveMTLS` modes (including panic paths for `MTLSAlways`), and `buildServerTLS` (with self-signed PEMs generated in `t.TempDir()`).
  - `pkg/gateway/gateway_test.go` spins up `ProvideServer` / `ProvideGrpcServer` via `fxtest.New` on `127.0.0.1:0` and hits `/health` and the gRPC health service. The HTTP smoke tests retry briefly to absorb the goroutine-startup race between `OnStart` returning and the listener calling `Accept`.
- Run `go test ./...` from the repo root.

## Deterministic regeneration

Both `buf.yaml` deps and `buf.gen.yaml` plugins are version-pinned through
`FileParams`. The example passes commit hashes for buf modules (e.g.
`GoogleAPIsVersion: "72c8614f3bd0466ea67931ef2c43d608"`) and tagged versions
for plugins (`ProtocolBuffersGoVersion: "v1.36.10"`, etc.). buf accepts either
form in `buf.yaml` deps.

Consequences:

- `go generate ./...` in `examples/app/proto/` is fully idempotent: identical
  `buf.lock` and untouched `go.mod`/`go.sum` across reruns.
- `gen.go` deliberately omits `go get -tool` — buf is already a `tool` in the
  root `go.mod`, so reruns don't drift its version.
- When bumping a buf-module pin, update both the value in
  `examples/app/gen/api/main.go` and let `go generate` produce the new
  `buf.lock` in the same commit.

The template does **not** declare `buf.build/grpc-ecosystem/grpc-gateway` as a
proto module dep. The CRUD output imports only `google/api/annotations.proto`
and well-known types. The grpc-gateway *plugin* (different artifact, lives in
`buf.gen.yaml`) is still required to produce `.pb.gw.go`.

## Working in this repo

- After changing anything under `genlib/grpc/`, re-run the example generator and verify the diff in `examples/app/proto/`:
  ```sh
  ( cd examples/app/gen/api && go run . )
  go build ./...
  go test ./...
  ```
- To regenerate `.pb.go` after a `.proto` change: `cd examples/app/proto && go generate ./...` (requires network — pulls buf modules).
- `examples/app/server/server.go` is hand-written. Adding a new method to the example generator means adding a stub here too, or `go build ./...` will fail with `does not implement … (missing method X)`.
- `make fmt` runs `gofmt` + `goimports` + `gci`. Use it before commits.
- `make clean` clears the Go test cache.

## Style

- Generator code uses panic via `genlib.Check`/`CheckClose` — fine for build-time tooling, not for runtime packages under `pkg/`.
- Public types have godoc comments; package-level docs live in `doc.go` files.
- The library is consumed downstream (notably `gitlab.authwise.io/authwise/kit`). API changes that affect generated output need to be coordinated with consumer regeneration.
