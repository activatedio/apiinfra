package crud_test

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/activatedio/apiinfra/pkg/service/crud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// A shelf holds books: `shelves/{shelf}/books/{book}`. The store keeps
// books by id and scopes every call to the shelf on the context.

type book struct {
	ID, Shelf, Title, Author string
	Touched                  bool
}

type bookMsg struct {
	Name, Title, Author string
}

type shelfKey struct{}

type memStore struct {
	books   map[string]*book
	failGet error
}

func shelfOf(ctx context.Context) string { s, _ := ctx.Value(shelfKey{}).(string); return s }

func (m *memStore) Find(ctx context.Context, key string) (*book, bool, error) {
	if m.failGet != nil {
		return nil, false, m.failGet
	}
	b, ok := m.books[key]
	if !ok || b.Shelf != shelfOf(ctx) {
		return nil, false, nil
	}
	return b, true, nil
}

func (m *memStore) List(ctx context.Context, page crud.PageRequest) (crud.Page[*book], error) {
	var ids []string
	for id, b := range m.books {
		if b.Shelf == shelfOf(ctx) && id > page.Token {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	out := crud.Page[*book]{}
	for i, id := range ids {
		if i == page.Size {
			out.NextPageToken = ids[i-1]
			break
		}
		out.Items = append(out.Items, m.books[id])
	}
	return out, nil
}

func (m *memStore) Create(_ context.Context, b *book) error { m.books[b.ID] = b; return nil }
func (m *memStore) Update(_ context.Context, b *book) error { m.books[b.ID] = b; return nil }
func (m *memStore) Delete(_ context.Context, id string) error {
	delete(m.books, id)
	return nil
}

func newBooks(store *memStore) *crud.Resource[*book, *bookMsg, string] {
	next := 0
	return &crud.Resource[*book, *bookMsg, string]{
		Kind:     "book",
		Store:    store,
		Writable: []string{"title", "author"},
		Paging:   crud.Paging{Default: 2, Max: 3},
		ResolveName: func(ctx context.Context, name string) (context.Context, string, error) {
			parts := strings.Split(name, "/")
			if len(parts) != 4 || parts[0] != "shelves" || parts[2] != "books" {
				return ctx, "", status.Errorf(codes.InvalidArgument, "bad name %q", name)
			}
			return context.WithValue(ctx, shelfKey{}, parts[1]), parts[3], nil
		},
		ResolveParent: func(ctx context.Context, parent string) (context.Context, error) {
			shelf, ok := strings.CutPrefix(parent, "shelves/")
			if !ok || shelf == "" {
				return ctx, status.Errorf(codes.InvalidArgument, "bad parent %q", parent)
			}
			return context.WithValue(ctx, shelfKey{}, shelf), nil
		},
		New: func(ctx context.Context) (*book, error) {
			next++
			return &book{ID: "b" + strconv.Itoa(next), Shelf: shelfOf(ctx)}, nil
		},
		Apply: func(_ context.Context, b *book, in *bookMsg, paths crud.Paths, _ bool) error {
			if paths["title"] {
				if in.Title == "" {
					return status.Error(codes.InvalidArgument, "title is required")
				}
				b.Title = in.Title
			}
			if paths["author"] {
				b.Author = in.Author
			}
			return nil
		},
		Touch: func(_ context.Context, b *book) { b.Touched = true },
		ToProto: func(_ context.Context, b *book) (*bookMsg, error) {
			return &bookMsg{Name: "shelves/" + b.Shelf + "/books/" + b.ID, Title: b.Title, Author: b.Author}, nil
		},
		BeforeCreate: func(ctx context.Context, b *book) error {
			for _, other := range store.books {
				if other.Shelf == shelfOf(ctx) && other.Title == b.Title {
					return status.Errorf(codes.AlreadyExists, "%q is already on the shelf", b.Title)
				}
			}
			return nil
		},
		BeforeDelete: func(_ context.Context, b *book) error {
			if b.Title == "keep" {
				return status.Error(codes.FailedPrecondition, "kept")
			}
			return nil
		},
	}
}

func seeded() *memStore {
	return &memStore{books: map[string]*book{
		"x1": {ID: "x1", Shelf: "s1", Title: "One"},
		"x2": {ID: "x2", Shelf: "s1", Title: "Two"},
		"x3": {ID: "x3", Shelf: "s1", Title: "keep"},
		"y1": {ID: "y1", Shelf: "s2", Title: "Elsewhere"},
	}}
}

func TestResource(t *testing.T) {
	ctx := context.Background()
	cases := map[string]struct {
		act func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], store *memStore)
	}{
		"get converts the entity": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], _ *memStore) {
			got, err := r.Get(ctx, "shelves/s1/books/x1")
			require.NoError(t, err)
			assert.Equal(t, &bookMsg{Name: "shelves/s1/books/x1", Title: "One"}, got)
		}},
		"get under another parent is NotFound": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], _ *memStore) {
			_, err := r.Get(ctx, "shelves/s2/books/x1")
			assert.Equal(t, codes.NotFound, status.Code(err))
		}},
		"a malformed name is the resolver's error, never a lookup": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], store *memStore) {
			store.failGet = errors.New("must not be called")
			_, err := r.Get(ctx, "books/x1")
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		}},
		"a store error passes through untouched": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], store *memStore) {
			boom := errors.New("boom")
			store.failGet = boom
			_, err := r.Get(ctx, "shelves/s1/books/x1")
			assert.ErrorIs(t, err, boom)
		}},
		"list pages with the default size and follows the token": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], _ *memStore) {
			page1, next, err := r.List(ctx, "shelves/s1", 0, "")
			require.NoError(t, err)
			assert.Len(t, page1, 2)
			require.NotEmpty(t, next)
			page2, next, err := r.List(ctx, "shelves/s1", 0, next)
			require.NoError(t, err)
			assert.Len(t, page2, 1)
			assert.Empty(t, next)
		}},
		"list clamps above the cap and refuses a negative size": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], _ *memStore) {
			all, _, err := r.List(ctx, "shelves/s1", 500, "")
			require.NoError(t, err)
			assert.Len(t, all, 3)
			_, _, err = r.List(ctx, "shelves/s1", -1, "")
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		}},
		"create writes every writable field under the parent": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], store *memStore) {
			got, err := r.Create(ctx, "shelves/s2", &bookMsg{Name: "ignored", Title: "New", Author: "A"})
			require.NoError(t, err)
			assert.Equal(t, "shelves/s2/books/b1", got.Name, "the name is the resource's, never the caller's")
			assert.Equal(t, "s2", store.books["b1"].Shelf)
			assert.Equal(t, "A", store.books["b1"].Author)
		}},
		"create refuses what Apply refuses, and stores nothing": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], store *memStore) {
			_, err := r.Create(ctx, "shelves/s2", &bookMsg{})
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.NotContains(t, store.books, "b1")
		}},
		"BeforeCreate refuses after Apply, and stores nothing": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], store *memStore) {
			_, err := r.Create(ctx, "shelves/s1", &bookMsg{Title: "One"})
			assert.Equal(t, codes.AlreadyExists, status.Code(err))
			assert.NotContains(t, store.books, "b1")
		}},
		"patch writes only the masked fields and touches": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], store *memStore) {
			got, err := r.Patch(ctx, "shelves/s1/books/x1", &bookMsg{Title: "", Author: "B"},
				&fieldmaskpb.FieldMask{Paths: []string{"author"}})
			require.NoError(t, err)
			assert.Equal(t, "One", got.Title, "an unmasked field is left alone")
			assert.Equal(t, "B", got.Author)
			assert.True(t, store.books["x1"].Touched)
		}},
		"patch refuses an empty mask and an output-only field": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], _ *memStore) {
			_, err := r.Patch(ctx, "shelves/s1/books/x1", &bookMsg{}, nil)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			_, err = r.Patch(ctx, "shelves/s1/books/x1", &bookMsg{}, &fieldmaskpb.FieldMask{Paths: []string{"name"}})
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		}},
		"update replaces every writable field": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], _ *memStore) {
			got, err := r.Update(ctx, "shelves/s1/books/x1", &bookMsg{Title: "Replaced"})
			require.NoError(t, err)
			assert.Equal(t, "Replaced", got.Title)
			assert.Empty(t, got.Author)
		}},
		"delete removes, honours BeforeDelete, and is NotFound when missing": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], store *memStore) {
			require.NoError(t, r.Delete(ctx, "shelves/s1/books/x1"))
			assert.NotContains(t, store.books, "x1")
			assert.Equal(t, codes.FailedPrecondition, status.Code(r.Delete(ctx, "shelves/s1/books/x3")))
			assert.Contains(t, store.books, "x3")
			assert.Equal(t, codes.NotFound, status.Code(r.Delete(ctx, "shelves/s1/books/x1")))
		}},
		"a top-level resource refuses a parent": {act: func(t *testing.T, r *crud.Resource[*book, *bookMsg, string], _ *memStore) {
			r.ResolveParent = nil
			_, _, err := r.List(ctx, "shelves/s1", 0, "")
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		}},
	}
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			store := seeded()
			v.act(t, newBooks(store), store)
		})
	}
}

func TestPagingSize(t *testing.T) {
	p := crud.Paging{Default: 100, Max: 1000}
	cases := map[string]struct {
		in   int32
		want int
		code codes.Code
	}{
		"zero takes the default":  {in: 0, want: 100},
		"within the cap":          {in: 7, want: 7},
		"over the cap is clamped": {in: 5000, want: 1000},
		"negative is refused":     {in: -1, code: codes.InvalidArgument},
	}
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			got, err := p.Size(v.in)
			if v.code != codes.OK {
				assert.Equal(t, v.code, status.Code(err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, v.want, got)
		})
	}
}
