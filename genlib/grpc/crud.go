package grpc

import (
	"fmt"

	"github.com/activatedio/protogen/proto"
)

// CrudParams contains the options to build the crud service
type CrudParams struct {
	Message        proto.Message
	MessagesTarget proto.File
	ServiceTarget  proto.Service
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
		}),
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

	mFile.AddMessages(
		NewGetRequestResponse(msg).Messages()...,
	)
	mFile.AddMessages(
		NewListRequestResponse(msg).Messages()...,
	)
	mFile.AddMessages(
		NewCreateRequestResponse(msg).Messages()...,
	)
	mFile.AddMessages(
		NewUpdateRequestResponse(msg).Messages()...,
	)
	mFile.AddMessages(
		NewPatchRequestResponse(msg).Messages()...,
	)
	mFile.AddMessages(
		NewDeleteRequestResponse(msg).Messages()...,
	)
}

// NewGetRequestResponse returns a RequestResponse with Request and Response messages derived from the provided proto message.
func NewGetRequestResponse(msg proto.Message) RequestResponse {
	return RequestResponse{
		Request:  proto.NewMessage(fmt.Sprintf("Get%sRequest", msg.GetName())),
		Response: proto.NewMessage(fmt.Sprintf("Get%sResponse", msg.GetName())),
	}
}

// NewListRequestResponse creates a RequestResponse with a List request and response message derived from the given proto.Message.
func NewListRequestResponse(msg proto.Message) RequestResponse {
	return RequestResponse{
		Request:  proto.NewMessage(fmt.Sprintf("List%sRequest", msg.GetName())),
		Response: proto.NewMessage(fmt.Sprintf("List%sResponse", msg.GetName())),
	}
}

// NewCreateRequestResponse initializes a RequestResponse with formatted create request and response messages.
func NewCreateRequestResponse(msg proto.Message) RequestResponse {
	return RequestResponse{
		Request:  proto.NewMessage(fmt.Sprintf("Create%sRequest", msg.GetName())),
		Response: proto.NewMessage(fmt.Sprintf("Create%sResponse", msg.GetName())),
	}
}

// NewUpdateRequestResponse creates a new RequestResponse with an Update request and response based on the given message name.
func NewUpdateRequestResponse(msg proto.Message) RequestResponse {
	return RequestResponse{
		Request:  proto.NewMessage(fmt.Sprintf("Update%sRequest", msg.GetName())),
		Response: proto.NewMessage(fmt.Sprintf("Update%sResponse", msg.GetName())),
	}
}

// NewPatchRequestResponse creates a RequestResponse with a "Patch" request and response message based on the given proto message.
func NewPatchRequestResponse(msg proto.Message) RequestResponse {
	return RequestResponse{
		Request:  proto.NewMessage(fmt.Sprintf("Patch%sRequest", msg.GetName())),
		Response: proto.NewMessage(fmt.Sprintf("Patch%sResponse", msg.GetName())),
	}
}

// NewDeleteRequestResponse creates a RequestResponse struct for delete operations based on the provided proto.Message.
func NewDeleteRequestResponse(msg proto.Message) RequestResponse {
	return RequestResponse{
		Request:  proto.NewMessage(fmt.Sprintf("Delete%sRequest", msg.GetName())),
		Response: proto.NewMessage(fmt.Sprintf("Delete%sResponse", msg.GetName())),
	}
}
