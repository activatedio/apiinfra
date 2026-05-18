package crud

import (
	"fmt"
	"strings"

	"github.com/activatedio/apiinfra/genlib/grpc"
	"github.com/activatedio/protogen/proto"
	"github.com/activatedio/protogen/tfl"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

var pl = pluralize.NewClient()

const (
	emptyMessageName     = "google.protobuf.Empty"
	fieldMaskMessageName = "google.protobuf.FieldMask"

	annotationsImport = "google/api/annotations.proto"
	fieldMaskImport   = "google/protobuf/field_mask.proto"
	emptyImport       = "google/protobuf/empty.proto"
)

// Ops is a bitmask selecting which CRUD operations All emits.
type Ops uint

// Individual CRUD operation flags. OpAll combines every flag.
const (
	OpGet Ops = 1 << iota
	OpList
	OpCreate
	OpUpdate
	OpPatch
	OpDelete

	OpAll = OpGet | OpList | OpCreate | OpUpdate | OpPatch | OpDelete
)

// Resource describes a CRUD-managed resource: the proto message that
// represents it, any parent path it lives under, and an optional Ops bitmask.
// When Ops is zero, All treats it as OpAll.
type Resource struct {
	Message    proto.Message
	ParentPath string
	Ops        Ops
}

func (r Resource) name() string       { return r.Message.GetName() }
func (r Resource) apiName() string    { return strcase.ToSnake(r.name()) }
func (r Resource) pluralName() string { return pl.Plural(r.apiName()) }
func (r Resource) qualifiedName() string {
	if pkg := r.Message.GetPackageName(); pkg != "" {
		return pkg + "." + r.name()
	}
	return r.name()
}

func (r Resource) normalizedParentPath() string {
	return strings.TrimSuffix(strings.TrimPrefix(r.ParentPath, "/"), "/")
}

// nameURL builds the single-resource URL: /<base>/{name=<parent>/<plural>/*}.
func (r Resource) nameURL(apiBase string) string {
	npp := r.normalizedParentPath()
	if npp != "" {
		npp += "/"
	}
	return fmt.Sprintf("%s/{name=%s%s/*}", apiBase, npp, r.pluralName())
}

// collectionURL builds the collection URL: /<base>/<plural> with no parent,
// or /<base>/{parent=<parent>}/<plural> when ParentPath is set.
func (r Resource) collectionURL(apiBase string) string {
	npp := r.normalizedParentPath()
	if npp == "" {
		return fmt.Sprintf("%s/%s", apiBase, r.pluralName())
	}
	return fmt.Sprintf("%s/{parent=%s}/%s", apiBase, npp, r.pluralName())
}

// All returns an Operation that emits the CRUD methods selected by r.Ops.
// A zero r.Ops is treated as OpAll.
func All(r Resource) grpc.Operation {
	ops := r.Ops
	if ops == 0 {
		ops = OpAll
	}
	return grpc.OperationFunc(func(t grpc.Target) {
		if ops&OpGet != 0 {
			applyGet(r, t)
		}
		if ops&OpList != 0 {
			applyList(r, t)
		}
		if ops&OpCreate != 0 {
			applyCreate(r, t)
		}
		if ops&OpUpdate != 0 {
			applyUpdate(r, t)
		}
		if ops&OpPatch != 0 {
			applyPatch(r, t)
		}
		if ops&OpDelete != 0 {
			applyDelete(r, t)
		}
	})
}

// Get returns an Operation that emits only the Get method for r.
func Get(r Resource) grpc.Operation {
	return grpc.OperationFunc(func(t grpc.Target) { applyGet(r, t) })
}

// List returns an Operation that emits only the List method for r.
func List(r Resource) grpc.Operation {
	return grpc.OperationFunc(func(t grpc.Target) { applyList(r, t) })
}

// Create returns an Operation that emits only the Create method for r.
func Create(r Resource) grpc.Operation {
	return grpc.OperationFunc(func(t grpc.Target) { applyCreate(r, t) })
}

// Update returns an Operation that emits only the Update method for r.
func Update(r Resource) grpc.Operation {
	return grpc.OperationFunc(func(t grpc.Target) { applyUpdate(r, t) })
}

// Patch returns an Operation that emits only the Patch method for r.
func Patch(r Resource) grpc.Operation {
	return grpc.OperationFunc(func(t grpc.Target) { applyPatch(r, t) })
}

// Delete returns an Operation that emits only the Delete method for r.
func Delete(r Resource) grpc.Operation {
	return grpc.OperationFunc(func(t grpc.Target) { applyDelete(r, t) })
}

func applyGet(r Resource, t grpc.Target) {
	t.ServiceImports.AddImports(proto.NewImport(annotationsImport))
	t.Messages.AddImports(proto.NewImport(fieldMaskImport))

	t.Service.AddMethods(
		proto.NewMethod(fmt.Sprintf("Get%s", r.name()), proto.MethodParams{
			RequestName:  fmt.Sprintf("Get%sRequest", r.name()),
			ResponseName: r.qualifiedName(),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(
				tfl.NewMessageValue().AddFields(
					tfl.NewStringField("get", r.nameURL(t.APIBasePath)),
				),
			)),
		),
	)

	t.Messages.AddMessages(NewGetRequest(r))
}

func applyList(r Resource, t grpc.Target) {
	t.ServiceImports.AddImports(proto.NewImport(annotationsImport))
	t.Messages.AddImports(proto.NewImport(fieldMaskImport))

	t.Service.AddMethods(
		proto.NewMethod(fmt.Sprintf("List%s", r.name()), proto.MethodParams{
			RequestName:  fmt.Sprintf("List%sRequest", r.name()),
			ResponseName: fmt.Sprintf("List%sResponse", r.name()),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(
				tfl.NewMessageValue().AddFields(
					tfl.NewStringField("get", r.collectionURL(t.APIBasePath)),
				),
			)),
		),
	)

	req, resp := NewListRequestResponse(r)
	t.Messages.AddMessages(req, resp)
}

func applyCreate(r Resource, t grpc.Target) {
	t.ServiceImports.AddImports(proto.NewImport(annotationsImport))

	t.Service.AddMethods(
		proto.NewMethod(fmt.Sprintf("Create%s", r.name()), proto.MethodParams{
			RequestName:  fmt.Sprintf("Create%sRequest", r.name()),
			ResponseName: r.qualifiedName(),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(
				tfl.NewMessageValue().AddFields(
					tfl.NewStringField("post", r.collectionURL(t.APIBasePath)),
					tfl.NewStringField("body", "*"),
				),
			)),
		),
	)

	t.Messages.AddMessages(NewCreateRequest(r))
}

func applyUpdate(r Resource, t grpc.Target) {
	t.ServiceImports.AddImports(proto.NewImport(annotationsImport))

	t.Service.AddMethods(
		proto.NewMethod(fmt.Sprintf("Update%s", r.name()), proto.MethodParams{
			RequestName:  fmt.Sprintf("Update%sRequest", r.name()),
			ResponseName: r.qualifiedName(),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(
				tfl.NewMessageValue().AddFields(
					tfl.NewStringField("put", r.nameURL(t.APIBasePath)),
					tfl.NewStringField("body", "*"),
				),
			)),
		),
	)

	t.Messages.AddMessages(NewUpdateRequest(r))
}

func applyPatch(r Resource, t grpc.Target) {
	t.ServiceImports.AddImports(proto.NewImport(annotationsImport))
	t.Messages.AddImports(proto.NewImport(fieldMaskImport))

	t.Service.AddMethods(
		proto.NewMethod(fmt.Sprintf("Patch%s", r.name()), proto.MethodParams{
			RequestName:  fmt.Sprintf("Patch%sRequest", r.name()),
			ResponseName: r.qualifiedName(),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(
				tfl.NewMessageValue().AddFields(
					tfl.NewStringField("patch", r.nameURL(t.APIBasePath)),
					tfl.NewStringField("body", "*"),
				),
			)),
		),
	)

	t.Messages.AddMessages(NewPatchRequest(r))
}

func applyDelete(r Resource, t grpc.Target) {
	t.ServiceImports.AddImports(
		proto.NewImport(annotationsImport),
		proto.NewImport(emptyImport),
	)

	t.Service.AddMethods(
		proto.NewMethod(fmt.Sprintf("Delete%s", r.name()), proto.MethodParams{
			RequestName:  fmt.Sprintf("Delete%sRequest", r.name()),
			ResponseName: emptyMessageName,
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(
				tfl.NewMessageValue().AddFields(
					tfl.NewStringField("delete", r.nameURL(t.APIBasePath)),
				),
			)),
		),
	)

	t.Messages.AddMessages(NewDeleteRequest(r))
}

// NewGetRequest returns the Get<Resource>Request message: { name, read_mask }.
func NewGetRequest(r Resource) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Get%sRequest", r.name())).AddFields(
		proto.NewField("name", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField("read_mask", proto.FieldParams{
			FieldType: fieldMaskMessageName,
			Number:    2,
		}),
	)
}

// NewListRequestResponse returns the List<Resource>Request and List<Resource>Response messages.
// The response uses the plural resource name as its list field (AIP-132).
func NewListRequestResponse(r Resource) (proto.Message, proto.Message) {
	req := proto.NewMessage(fmt.Sprintf("List%sRequest", r.name())).AddFields(
		proto.NewField("parent", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField("page_size", proto.FieldParams{
			FieldType: "int32",
			Number:    2,
		}),
		proto.NewField("page_token", proto.FieldParams{
			FieldType: "string",
			Number:    3,
		}),
		proto.NewField("filter", proto.FieldParams{
			FieldType: "string",
			Number:    4,
		}),
		proto.NewField("order_by", proto.FieldParams{
			FieldType: "string",
			Number:    5,
		}),
		proto.NewField("read_mask", proto.FieldParams{
			FieldType: fieldMaskMessageName,
			Number:    6,
		}),
	)
	resp := proto.NewMessage(fmt.Sprintf("List%sResponse", r.name())).AddFields(
		proto.NewField(r.pluralName(), proto.FieldParams{
			FieldType: r.qualifiedName(),
			Repeated:  true,
			Number:    1,
		}),
		proto.NewField("next_page_token", proto.FieldParams{
			FieldType: "string",
			Number:    2,
		}),
	)
	return req, resp
}

// NewCreateRequest returns the Create<Resource>Request message: { parent, <resource> }.
func NewCreateRequest(r Resource) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Create%sRequest", r.name())).AddFields(
		proto.NewField("parent", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField(r.apiName(), proto.FieldParams{
			FieldType: r.qualifiedName(),
			Number:    2,
		}),
	)
}

// NewUpdateRequest returns the Update<Resource>Request message: { name, <resource> }.
func NewUpdateRequest(r Resource) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Update%sRequest", r.name())).AddFields(
		proto.NewField("name", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField(r.apiName(), proto.FieldParams{
			FieldType: r.qualifiedName(),
			Number:    2,
		}),
	)
}

// NewPatchRequest returns the Patch<Resource>Request message: { name, <resource>, update_mask }.
func NewPatchRequest(r Resource) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Patch%sRequest", r.name())).AddFields(
		proto.NewField("name", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField(r.apiName(), proto.FieldParams{
			FieldType: r.qualifiedName(),
			Number:    2,
		}),
		proto.NewField("update_mask", proto.FieldParams{
			FieldType: fieldMaskMessageName,
			Number:    3,
		}),
	)
}

// NewDeleteRequest returns the Delete<Resource>Request message: { name }.
func NewDeleteRequest(r Resource) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Delete%sRequest", r.name())).AddFields(
		proto.NewField("name", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
	)
}
