package grpc

import "github.com/activatedio/protogen/proto"

type crudBuilder struct {
	// MessagesTarget is the file to add messages to
	MessagesTarget proto.File
	ServiceTarget  proto.Service
	// ServiceImportTarget is the file to add any imports required as a result of the service definitions
	ServiceImportTarget proto.File
	// APIBasePath is the path to prepend to all http paths
	APIBasePath string
}

func (c *crudBuilder) buildCrudParams(params CrudMessageParams) CrudParams {
	return CrudParams{
		Message:             params.Message,
		MessagesTarget:      c.MessagesTarget,
		ServiceTarget:       c.ServiceTarget,
		ServiceImportTarget: c.ServiceImportTarget,
		ParentPath:          params.ParentPath,
		APIBasePath:         c.APIBasePath,
	}
}

func (c *crudBuilder) BuildCrud(params CrudMessageParams) {
	BuildCrud(c.buildCrudParams(params))
}

func (c *crudBuilder) BuildGet(params CrudMessageParams) {
	BuildGet(c.buildCrudParams(params))
}

func (c *crudBuilder) BuildList(params CrudMessageParams) {
	BuildList(c.buildCrudParams(params))
}

func (c *crudBuilder) BuildCreate(params CrudMessageParams) {
	BuildCreate(c.buildCrudParams(params))
}

func (c *crudBuilder) BuildUpdate(params CrudMessageParams) {
	BuildUpdate(c.buildCrudParams(params))
}

func (c *crudBuilder) BuildPatch(params CrudMessageParams) {
	BuildPatch(c.buildCrudParams(params))
}

func (c *crudBuilder) BuildDelete(params CrudMessageParams) {
	BuildDelete(c.buildCrudParams(params))
}

// NewCrudBuilder returns a CrudBuilder configured with the supplied parameters.
func NewCrudBuilder(params CrudBuilderParams) CrudBuilder {
	return &crudBuilder{
		MessagesTarget:      params.MessagesTarget,
		ServiceTarget:       params.ServiceTarget,
		ServiceImportTarget: params.ServiceImportTarget,
		APIBasePath:         params.APIBasePath,
	}
}
