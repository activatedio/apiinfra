package grpc

// ServiceBuilder applies a sequence of Operations to a single Target. Each
// Operation contributes methods, supporting messages, and imports. CRUD,
// custom actions, and future op families all reach the service through this
// builder.
type ServiceBuilder struct {
	target Target
}

// NewServiceBuilder returns a builder that writes to the supplied target.
func NewServiceBuilder(target Target) *ServiceBuilder {
	return &ServiceBuilder{target: target}
}

// Add applies the given operations to the underlying target, in order.
func (b *ServiceBuilder) Add(ops ...Operation) *ServiceBuilder {
	for _, op := range ops {
		op.Apply(b.target)
	}
	return b
}

// Target returns the underlying target. Useful when the caller needs direct
// protogen access alongside Add — for example, registering imports that span
// multiple operations.
func (b *ServiceBuilder) Target() Target { return b.target }
