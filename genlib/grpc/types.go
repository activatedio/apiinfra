package grpc

import "github.com/activatedio/protogen/proto"

// RequestResponse represents a pair of proto messages for a request and its corresponding response.
type RequestResponse struct {
	Request  proto.Message
	Response proto.Message
}

// CrudParams contains the options to build the crud service
type CrudParams struct {
	// Message is the message definition for the crud endpoints
	Message proto.Message
	// MessagesTarget is the file to add messages to
	MessagesTarget proto.File
	ServiceTarget  proto.Service
	// ServiceImportTarget is the file to add any imports required as a result of the service definitions
	ServiceImportTarget proto.File
	// ParentPath is the api path to prefix in front of the paths for options
	ParentPath string
	// APIBasePath is the path to prepend to all http paths
	APIBasePath string
}

// CrudParams contains the options to build the crud service
type CrudBuilderParams struct {
	// MessagesTarget is the file to add messages to
	MessagesTarget proto.File
	ServiceTarget  proto.Service
	// ServiceImportTarget is the file to add any imports required as a result of the service definitions
	ServiceImportTarget proto.File
	// APIBasePath is the path to prepend to all http paths
	APIBasePath string
}

type CrudMessageParams struct {
	// Message is the message definition for the crud endpoints
	Message proto.Message
	// ParentPath is the api path to prefix in front of the paths for options
	ParentPath string
}

// Messages returns a slice containing the Request and Response proto messages from the RequestResponse struct.
func (rr RequestResponse) Messages() []proto.Message {
	return []proto.Message{rr.Request, rr.Response}
}

type CrudBuilder interface {
	BuildCrud(params CrudMessageParams)
	BuildGet(params CrudMessageParams)
	BuildList(params CrudMessageParams)
	BuildCreate(params CrudMessageParams)
	BuildUpdate(params CrudMessageParams)
	BuildPatch(params CrudMessageParams)
	BuildDelete(params CrudMessageParams)
}
