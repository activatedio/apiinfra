package grpc

import (
	"fmt"

	"github.com/activatedio/protogen/proto"
	"github.com/activatedio/protogen/tfl"
	"github.com/iancoleman/strcase"
)

// CrudParams contains the options to build the crud service
type CrudParams struct {
	Message        proto.Message
	MessagesTarget proto.File
	ServiceTarget  proto.Service
	// ParentPath is the api path to prefix in front of the paths for options
	ParentPath string
}

// BuildCrud add the methods needed ot a service for a crud message
func BuildCrud(params CrudParams) {

	msg := params.Message
	name := msg.GetName()
	svc := params.ServiceTarget
	mFile := params.MessagesTarget

	svc.AddMethods(
		proto.NewMethod(fmt.Sprintf("Get%sFoo", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Get%sRequest", name),
			ResponseName: fmt.Sprintf("Get%sResponse", name),
		}).AddOptions(
			proto.NewOption("google.api.http", proto.NewMessageValueConstant(tfl.NewMessageValue())),
		),
		proto.NewMethod(fmt.Sprintf("List%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("List%sRequest", name),
			ResponseName: fmt.Sprintf("List%sResponse", name),
		}),
		proto.NewMethod(fmt.Sprintf("Create%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Create%sRequest", name),
			ResponseName: fmt.Sprintf("Create%sResponse", name),
		}),
		proto.NewMethod(fmt.Sprintf("Update%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Update%sRequest", name),
			ResponseName: fmt.Sprintf("Update%sResponse", name),
		}),
		proto.NewMethod(fmt.Sprintf("Patch%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Patch%sRequest", name),
			ResponseName: fmt.Sprintf("Patch%sResponse", name),
		}),
		proto.NewMethod(fmt.Sprintf("Delete%s", name), proto.MethodParams{
			RequestName:  fmt.Sprintf("Delete%sRequest", name),
			ResponseName: fmt.Sprintf("Delete%sResponse", name),
		}),
	)

	mFile.AddImports(
		proto.NewImport("google/protobuf/field_mask.proto"),
		proto.NewImport("google/protobuf/empty.proto"),
	)

	mFile.AddMessages(
		NewGetRequest(msg),
	)
	mFile.AddMessages(
		NewListRequestResponse(msg).Messages()...,
	)
	mFile.AddMessages(
		NewCreateRequest(msg),
	)
	mFile.AddMessages(
		NewUpdateRequest(msg),
	)
	mFile.AddMessages(
		NewPatchRequest(msg),
	)
	mFile.AddMessages(
		NewDeleteRequest(msg),
	)
}

// NewGetRequest returns a RequestResponse with Request and Response messages derived from the provided proto message.
func NewGetRequest(msg proto.Message) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Get%sRequest", msg.GetName())).AddFields(
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
func NewListRequestResponse(msg proto.Message) RequestResponse {
	return RequestResponse{
		Request: proto.NewMessage(fmt.Sprintf("List%sRequest", msg.GetName())).AddFields(
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
		Response: proto.NewMessage(fmt.Sprintf("List%sResponse", msg.GetName())).AddFields(
			proto.NewField("list", proto.FieldParams{
				FieldType: GetQualifiedName(msg),
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
func NewCreateRequest(msg proto.Message) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Create%sRequest", msg.GetName())).AddFields(
		proto.NewField("parent", proto.FieldParams{
			FieldType: "string",
			Number:    1,
		}),
		proto.NewField(strcase.ToLowerCamel(msg.GetName()), proto.FieldParams{
			FieldType: GetQualifiedName(msg),
			Number:    2,
		}),
	)
}

// NewUpdateRequest creates a new RequestResponse with an Update request and response based on the given message name.
func NewUpdateRequest(msg proto.Message) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Update%sRequest", msg.GetName())).AddFields(
		proto.NewField("parent", proto.FieldParams{
			FieldType: "name",
			Number:    1,
		}),
		proto.NewField(strcase.ToLowerCamel(msg.GetName()), proto.FieldParams{
			FieldType: GetQualifiedName(msg),
			Number:    2,
		}),
	)
}

// NewPatchRequest creates a RequestResponse with a "Patch" request and response message based on the given proto message.
func NewPatchRequest(msg proto.Message) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Patch%sRequest", msg.GetName())).AddFields(
		proto.NewField("parent", proto.FieldParams{
			FieldType: "name",
			Number:    1,
		}),
		proto.NewField(strcase.ToLowerCamel(msg.GetName()), proto.FieldParams{
			FieldType: GetQualifiedName(msg),
			Number:    2,
		}),
		proto.NewField("field_mask", proto.FieldParams{
			FieldType: "google.protobuf.FieldMask",
			Number:    3,
		}),
	)
}

// NewDeleteRequest creates a RequestResponse struct for delete operations based on the provided proto.Message.
func NewDeleteRequest(msg proto.Message) proto.Message {
	return proto.NewMessage(fmt.Sprintf("Delete%sRequest", msg.GetName())).AddFields(
		proto.NewField("parent", proto.FieldParams{
			FieldType: "name",
			Number:    1,
		}),
	)
}
