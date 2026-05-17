package grpc

import (
	"fmt"
	"strings"

	"github.com/activatedio/protogen/proto"
	"github.com/activatedio/protogen/tfl"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

var (
	pl = pluralize.NewClient()
)

const (
	// EmptyMessageName is the fully qualified name of the well-known Empty message.
	EmptyMessageName = "google.protobuf.Empty"
)

// GetAPIMessageName returns the name of the message in snake_case format derived from the Message.GetName() result.
func (c CrudParams) GetAPIMessageName() string {
	return strcase.ToSnake(c.Message.GetName())
}

// GetPluralAPIMessageName returns the pluralized form of the API message name for the current CrudParams instance.
func (c CrudParams) GetPluralAPIMessageName() string {
	return pl.Plural(c.GetAPIMessageName())
}

// GetQualifiedMessageName returns the fully qualified name of the proto.Message by combining its package and message name.
func (c CrudParams) GetQualifiedMessageName() string {
	return GetQualifiedName(c.Message)
}

// GetNormalizedParentPath returns the ParentPath string without leading or trailing slashes.
func (c CrudParams) GetNormalizedParentPath() string {
	return strings.TrimSuffix(strings.TrimPrefix(c.ParentPath, "/"), "/")
}

// BuildCrudHeaders adds the imports required by the generated crud service definitions.
func BuildCrudHeaders(params CrudParams) {
	params.ServiceImportTarget.AddImports(proto.NewImport("google/api/annotations.proto"))
}

// BuildCrud add the methods needed ot a service for a crud message
func BuildCrud(params CrudParams) {

	BuildCrudHeaders(params)
	BuildGet(params)
	BuildList(params)
	BuildCreate(params)
	BuildUpdate(params)
	BuildPatch(params)
	BuildDelete(params)
}

// BuildGet builds support for a get method
func BuildGet(params CrudParams) {

	name := params.Message.GetName()

	params.ServiceTarget.AddMethods(
		proto.NewMethod(fmt.Sprintf("Get%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Get%sRequest", name),
			ResponseName: GetQualifiedName(params.Message),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(newGetOptions(params))),
		),
	)

	params.MessagesTarget.AddMessages(
		NewGetRequest(params),
	)

	params.MessagesTarget.AddImports(
		proto.NewImport("google/protobuf/field_mask.proto"),
	)
}

// BuildList creates objects for a list operation
func BuildList(params CrudParams) {

	name := params.Message.GetName()
	params.ServiceTarget.AddMethods(
		proto.NewMethod(fmt.Sprintf("List%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("List%sRequest", name),
			ResponseName: fmt.Sprintf("List%sResponse", name),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(newListOptions(params))),
		),
	)

	params.MessagesTarget.AddMessages(
		NewListRequestResponse(params).Messages()...,
	)

	params.MessagesTarget.AddImports(
		proto.NewImport("google/protobuf/field_mask.proto"),
	)
}

// BuildCreate builds operations for a create
func BuildCreate(params CrudParams) {

	name := params.Message.GetName()

	params.ServiceTarget.AddMethods(
		proto.NewMethod(fmt.Sprintf("Create%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Create%sRequest", name),
			ResponseName: GetQualifiedName(params.Message),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(newCreateOptions(params))),
		),
	)

	params.MessagesTarget.AddMessages(
		NewCreateRequest(params),
	)
}

// BuildUpdate builds operations for a create
func BuildUpdate(params CrudParams) {

	name := params.Message.GetName()

	params.ServiceTarget.AddMethods(
		proto.NewMethod(fmt.Sprintf("Update%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Update%sRequest", name),
			ResponseName: GetQualifiedName(params.Message),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(newUpdateOptions(params))),
		),
	)

	params.MessagesTarget.AddMessages(
		NewUpdateRequest(params),
	)
}

// BuildPatch builds operations for a create
func BuildPatch(params CrudParams) {

	name := params.Message.GetName()

	params.ServiceTarget.AddMethods(

		proto.NewMethod(fmt.Sprintf("Patch%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Patch%sRequest", name),
			ResponseName: GetQualifiedName(params.Message),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(newPatchOptions(params))),
		),
	)

	params.MessagesTarget.AddMessages(
		NewPatchRequest(params),
	)

	params.MessagesTarget.AddImports(
		proto.NewImport("google/protobuf/field_mask.proto"),
	)
}

// BuildDelete builds operations for a create
func BuildDelete(params CrudParams) {

	name := params.Message.GetName()

	params.ServiceImportTarget.AddImports(proto.NewImport("google/protobuf/empty.proto"))

	params.ServiceTarget.AddMethods(
		proto.NewMethod(fmt.Sprintf("Delete%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Delete%sRequest", name),
			ResponseName: EmptyMessageName,
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(newDeleteOptions(params))),
		),
	)

	params.MessagesTarget.AddMessages(
		NewDeleteRequest(params),
	)

}

func newGetOptions(params CrudParams) tfl.MessageValue {

	return tfl.NewMessageValue().AddFields(
		tfl.NewStringField("get", namePath(params)),
	)
}

func newDeleteOptions(params CrudParams) tfl.MessageValue {

	return tfl.NewMessageValue().AddFields(
		tfl.NewStringField("delete", namePath(params)),
	)
}

func newListOptions(params CrudParams) tfl.MessageValue {

	return tfl.NewMessageValue().AddFields(
		tfl.NewStringField("get", parentPath(params)+"/*"),
	)
}

func newCreateOptions(params CrudParams) tfl.MessageValue {

	return tfl.NewMessageValue().AddFields(
		tfl.NewStringField("post", parentPath(params)+"/*"),
		tfl.NewStringField("body", "*"),
	)
}

func newUpdateOptions(params CrudParams) tfl.MessageValue {

	return tfl.NewMessageValue().AddFields(
		tfl.NewStringField("put", namePath(params)),
		tfl.NewStringField("body", "*"),
	)
}

func newPatchOptions(params CrudParams) tfl.MessageValue {

	return tfl.NewMessageValue().AddFields(
		tfl.NewStringField("patch", namePath(params)),
		tfl.NewStringField("body", "*"),
	)
}

func namePath(params CrudParams) string {

	sb := strings.Builder{}

	sb.WriteString(params.APIBasePath)

	npp := params.GetNormalizedParentPath()
	if npp != "" {
		npp += "/"
	}

	fmt.Fprintf(&sb, "/{name=%s%s/*}", npp, params.GetPluralAPIMessageName())

	return sb.String()

}

func parentPath(params CrudParams) string {

	sb := strings.Builder{}

	sb.WriteString(params.APIBasePath)

	npp := params.GetNormalizedParentPath()

	if npp == "" {
		fmt.Fprintf(&sb, "/%s", params.GetPluralAPIMessageName())
	} else {
		fmt.Fprintf(&sb, "/{parent=%s}/%s", npp, params.GetPluralAPIMessageName())
	}

	return sb.String()

}

// NewGetRequest returns a RequestResponse with Request and Response messages derived from the provided proto message.
func NewGetRequest(params CrudParams) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Get%sRequest", params.Message.GetName())).AddFields(
		proto.NewField("name", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField("fields", proto.FieldParams{
			FieldType: "string",
			Repeated:  true,
			Number:    2,
		}),
	)
}

// NewListRequestResponse creates a RequestResponse with a List request and response message derived from the given proto.Message.
func NewListRequestResponse(params CrudParams) RequestResponse {
	return RequestResponse{
		Request: proto.NewMessage(fmt.Sprintf("List%sRequest", params.Message.GetName())).AddFields(
			proto.NewField("parent", proto.FieldParams{
				FieldType: "string",
				Number:    1,
			}),
			proto.NewField("fields", proto.FieldParams{
				FieldType: "string",
				Repeated:  true,
				Number:    2,
			}),
			proto.NewField("page_size", proto.FieldParams{
				FieldType: "int32",
				Number:    3,
			}),
			proto.NewField("page_token", proto.FieldParams{
				FieldType: "string",
				Number:    4,
			}),
			proto.NewField("selection", proto.FieldParams{
				FieldType: "string",
				Number:    5,
			}),
		),
		Response: proto.NewMessage(fmt.Sprintf("List%sResponse", params.Message.GetName())).AddFields(
			proto.NewField("list", proto.FieldParams{
				FieldType: params.GetQualifiedMessageName(),
				Repeated:  true,
				Number:    1,
			}),
			proto.NewField("next_page_token", proto.FieldParams{
				FieldType: "string",
				Number:    2,
			}),
		),
	}
}

// NewCreateRequest initializes a RequestResponse with formatted create request and response messages.
func NewCreateRequest(params CrudParams) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Create%sRequest", params.Message.GetName())).AddFields(
		proto.NewField("parent", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField(params.GetAPIMessageName(), proto.FieldParams{
			FieldType: params.GetQualifiedMessageName(),
			Number:    2,
		}),
	)
}

// NewUpdateRequest creates a new RequestResponse with an Update request and response based on the given message name.
func NewUpdateRequest(params CrudParams) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Update%sRequest", params.Message.GetName())).AddFields(
		proto.NewField("name", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField(params.GetAPIMessageName(), proto.FieldParams{
			FieldType: params.GetQualifiedMessageName(),
			Number:    2,
		}),
	)
}

// NewPatchRequest creates a RequestResponse with a "Patch" request and response message based on the given proto message.
func NewPatchRequest(params CrudParams) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Patch%sRequest", params.Message.GetName())).AddFields(
		proto.NewField("name", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField(params.GetAPIMessageName(), proto.FieldParams{
			FieldType: params.GetQualifiedMessageName(),
			Number:    2,
		}),
		proto.NewField("field_mask", proto.FieldParams{
			FieldType: "google.protobuf.FieldMask",
			Number:    3,
		}),
	)
}

// NewDeleteRequest creates a RequestResponse struct for delete operations based on the provided proto.Message.
func NewDeleteRequest(params CrudParams) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Delete%sRequest", params.Message.GetName())).AddFields(
		proto.NewField("name", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
	)
}
