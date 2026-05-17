package grpc

import "github.com/activatedio/protogen/proto"

// Target groups the proto artifacts an Operation writes into.
type Target struct {
	// Messages is the file that receives request/response message definitions.
	Messages proto.File
	// Service is the service that receives generated methods.
	Service proto.Service
	// ServiceImports is the file containing Service; receives any imports the
	// service definition needs (e.g. google/api/annotations.proto).
	ServiceImports proto.File
	// APIBasePath is prepended to every HTTP path emitted by an Operation.
	APIBasePath string
}

// Operation is one unit of generation: a method (plus supporting messages and
// imports) that gets applied to a service Target. CRUD operations, custom
// actions, and other op families implement this interface.
type Operation interface {
	Apply(t Target)
}

// OperationFunc adapts a plain function into an Operation. Use this as an
// escape hatch when you need to emit arbitrary methods by calling protogen
// primitives directly.
type OperationFunc func(t Target)

// Apply implements Operation by invoking f.
func (f OperationFunc) Apply(t Target) { f(t) }
