package crud

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// PageRequest is one page of a list: a resolved size and the caller's token.
type PageRequest struct {
	Size  int
	Token string
}

// Page is one page of results and the token to the next ("" at the end).
type Page[M any] struct {
	Items         []M
	NextPageToken string
}

// Store is what a Resource reads and writes through. The context a Store
// method receives is the one ResolveName or ResolveParent returned, so a
// store scoped by context (a parent id on it) sees that scope.
type Store[M any, K comparable] interface {
	// Find returns the entity with key, or found false.
	Find(ctx context.Context, key K) (entity M, found bool, err error)
	// List returns one page.
	List(ctx context.Context, page PageRequest) (Page[M], error)
	Create(ctx context.Context, entity M) error
	Update(ctx context.Context, entity M) error
	Delete(ctx context.Context, key K) error
}

// Paging is a list's page-size policy.
type Paging struct {
	// Default is the size of a page the caller names none for.
	Default int
	// Max caps a requested size; a larger one is clamped, not refused.
	Max int
}

// Size resolves a requested page_size: 0 takes the default, above Max is
// clamped, below zero is InvalidArgument.
func (p Paging) Size(requested int32) (int, error) {
	n := int(requested)
	switch {
	case n < 0:
		return 0, status.Error(codes.InvalidArgument, "page_size must not be negative")
	case n == 0:
		n = p.Default
	case p.Max > 0 && n > p.Max:
		n = p.Max
	}
	return n, nil
}

// Paths is the set of field paths an operation writes.
type Paths map[string]bool

// AllPaths names every writable field: what an Update (PUT) or a Create
// writes.
func AllPaths(writable []string) Paths {
	out := Paths{}
	for _, w := range writable {
		out[w] = true
	}
	return out
}

// MaskPaths validates a PATCH's update_mask against the writable fields
// and returns the set. An empty mask is refused: a PATCH that replaced
// everything would be a PUT the caller did not ask for. Paths arrive as
// proto field names (protojson maps JSON camelCase on the way in).
func MaskPaths(mask *fieldmaskpb.FieldMask, writable ...string) (Paths, error) {
	if len(mask.GetPaths()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "update_mask is required and must name at least one field")
	}
	allowed := AllPaths(writable)
	out := Paths{}
	for _, p := range mask.GetPaths() {
		if !allowed[p] {
			return nil, status.Errorf(codes.InvalidArgument, "update_mask names no writable field: %q", p)
		}
		out[p] = true
	}
	return out, nil
}

// Resource is one resource's CRUD plumbing over a Store: M is the stored
// entity, P the API message, K the storage key.
type Resource[M any, P any, K comparable] struct {
	// Kind names the resource in error messages ("application").
	Kind string
	// Store reads and writes the entities.
	Store Store[M, K]
	// Writable are the proto field names a Create, Update or Patch writes.
	Writable []string
	// Paging is the list's page-size policy.
	Paging Paging

	// ResolveName parses a resource name into its key, returning the
	// context the Store is called with (a parent's scope on it). A
	// malformed name is the hook's InvalidArgument; a missing parent its
	// NotFound.
	ResolveName func(ctx context.Context, name string) (context.Context, K, error)
	// ResolveParent parses a collection's parent and returns the context
	// the Store is called with. Nil for a top-level resource, whose parent
	// is "".
	ResolveParent func(ctx context.Context, parent string) (context.Context, error)
	// New returns a fresh entity for Create: its key minted, its creation
	// stamps set. Apply then writes every writable field onto it.
	New func(ctx context.Context) (M, error)
	// Apply validates the named fields of in and copies them onto entity.
	// creating is true inside Create.
	Apply func(ctx context.Context, entity M, in P, paths Paths, creating bool) error
	// BeforeCreate runs after Apply and before the entity is stored: a
	// uniqueness check, or replacing what the new entity supersedes.
	// Optional.
	BeforeCreate func(ctx context.Context, entity M) error
	// Touch marks an entity changed before an Update or Patch is stored
	// (an update time, a modified-by). Optional.
	Touch func(ctx context.Context, entity M)
	// ToProto converts an entity to its message; output-only fields are
	// set here.
	ToProto func(ctx context.Context, entity M) (P, error)
	// BeforeDelete refuses a delete (FailedPrecondition, say) or does what
	// must go first. Optional.
	BeforeDelete func(ctx context.Context, entity M) error
}

// Find resolves name and returns the entity under the context its Store
// was called with, or NotFound: the building block of a custom verb.
func (r *Resource[M, P, K]) Find(ctx context.Context, name string) (context.Context, M, error) {
	var zero M
	ctx, key, err := r.ResolveName(ctx, name)
	if err != nil {
		return ctx, zero, err
	}
	e, found, err := r.Store.Find(ctx, key)
	if err != nil {
		return ctx, zero, err
	}
	if !found {
		return ctx, zero, status.Errorf(codes.NotFound, "%s %s was not found", r.Kind, name)
	}
	return ctx, e, nil
}

// Get is a Get method.
func (r *Resource[M, P, K]) Get(ctx context.Context, name string) (P, error) {
	ctx, e, err := r.Find(ctx, name)
	if err != nil {
		var zero P
		return zero, err
	}
	return r.ToProto(ctx, e)
}

// List is a List method: one page of messages and the next page's token.
func (r *Resource[M, P, K]) List(ctx context.Context, parent string, pageSize int32, pageToken string) ([]P, string, error) {
	ctx, err := r.parent(ctx, parent)
	if err != nil {
		return nil, "", err
	}
	size, err := r.Paging.Size(pageSize)
	if err != nil {
		return nil, "", err
	}
	page, err := r.Store.List(ctx, PageRequest{Size: size, Token: pageToken})
	if err != nil {
		return nil, "", err
	}
	out, err := r.Protos(ctx, page.Items)
	if err != nil {
		return nil, "", err
	}
	return out, page.NextPageToken, nil
}

// Protos converts entities to messages, for a list a consumer pages itself.
func (r *Resource[M, P, K]) Protos(ctx context.Context, entities []M) ([]P, error) {
	out := make([]P, 0, len(entities))
	for _, e := range entities {
		p, err := r.ToProto(ctx, e)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// Create is a Create method: New, Apply of every writable field,
// BeforeCreate, then the store.
func (r *Resource[M, P, K]) Create(ctx context.Context, parent string, in P) (P, error) {
	var zero P
	ctx, err := r.parent(ctx, parent)
	if err != nil {
		return zero, err
	}
	e, err := r.New(ctx)
	if err != nil {
		return zero, err
	}
	if err := r.Apply(ctx, e, in, AllPaths(r.Writable), true); err != nil {
		return zero, err
	}
	if r.BeforeCreate != nil {
		if err := r.BeforeCreate(ctx, e); err != nil {
			return zero, err
		}
	}
	if err := r.Store.Create(ctx, e); err != nil {
		return zero, err
	}
	return r.ToProto(ctx, e)
}

// Update is an Update (PUT) method: a Patch of every writable field.
func (r *Resource[M, P, K]) Update(ctx context.Context, name string, in P) (P, error) {
	return r.write(ctx, name, in, AllPaths(r.Writable))
}

// Patch is a Patch method: the fields update_mask names, writable only.
func (r *Resource[M, P, K]) Patch(ctx context.Context, name string, in P, mask *fieldmaskpb.FieldMask) (P, error) {
	paths, err := MaskPaths(mask, r.Writable...)
	if err != nil {
		var zero P
		return zero, err
	}
	return r.write(ctx, name, in, paths)
}

func (r *Resource[M, P, K]) write(ctx context.Context, name string, in P, paths Paths) (P, error) {
	var zero P
	ctx, e, err := r.Find(ctx, name)
	if err != nil {
		return zero, err
	}
	if err := r.Apply(ctx, e, in, paths, false); err != nil {
		return zero, err
	}
	if r.Touch != nil {
		r.Touch(ctx, e)
	}
	if err := r.Store.Update(ctx, e); err != nil {
		return zero, err
	}
	return r.ToProto(ctx, e)
}

// Delete is a Delete method.
func (r *Resource[M, P, K]) Delete(ctx context.Context, name string) error {
	ctx, key, err := r.ResolveName(ctx, name)
	if err != nil {
		return err
	}
	e, found, err := r.Store.Find(ctx, key)
	if err != nil {
		return err
	}
	if !found {
		return status.Errorf(codes.NotFound, "%s %s was not found", r.Kind, name)
	}
	if r.BeforeDelete != nil {
		if err := r.BeforeDelete(ctx, e); err != nil {
			return err
		}
	}
	return r.Store.Delete(ctx, key)
}

func (r *Resource[M, P, K]) parent(ctx context.Context, parent string) (context.Context, error) {
	if r.ResolveParent == nil {
		if parent != "" {
			return ctx, status.Errorf(codes.InvalidArgument, "a %s has no parent, got %q", r.Kind, parent)
		}
		return ctx, nil
	}
	return r.ResolveParent(ctx, parent)
}
