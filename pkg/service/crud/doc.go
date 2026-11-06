// Package crud is the runtime half of genlib/grpc/crud: the request
// plumbing every generated CRUD method repeats, so a service's handler for
// one is a line, and what is left by hand is what is particular to the
// resource — how its name resolves, which fields are writable and how they
// are validated, how it converts to its message, and any precondition on
// delete.
//
// It follows the same AIP conventions as the generator: names are the
// public identity and a malformed one is InvalidArgument, never a lookup;
// a missing resource is NotFound; page_size 0 takes the default, above the
// cap is clamped, below zero is refused (AIP-158); a PATCH needs a
// non-empty update_mask naming writable fields only, and an Update (PUT)
// is a PATCH of every writable field (AIP-134).
//
// The package is storage-agnostic: a Resource reads and writes through a
// Store, which a consumer adapts from its repositories once, generically.
// It owns no transactions, authorization or error mapping: those stay in
// the consumer's interceptor chain, and every error a hook or a Store
// returns is passed through untouched.
//
//	apps := crud.Resource[*model.App, *pb.App, string]{
//		Kind:          "application",
//		Store:         appStore,
//		Writable:      []string{"display_name", "url"},
//		Paging:        crud.Paging{Default: 100, Max: 1000},
//		ResolveName:   resolveAppName,   // name   -> (scoped ctx, key)
//		ResolveParent: resolveOrgParent, // parent -> scoped ctx
//		New:           newApp,
//		Apply:         applyApp,
//		ToProto:       toProtoApp,
//	}
//
//	func (s *Server) GetApp(ctx context.Context, r *pb.GetAppRequest) (*pb.App, error) {
//		return s.apps.Get(ctx, r.GetName())
//	}
package crud
