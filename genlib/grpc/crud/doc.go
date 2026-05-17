// Package crud emits AIP-compliant CRUD methods (Get, List, Create, Update,
// Patch, Delete) onto a grpc.ServiceBuilder. Each function returns a
// grpc.Operation so callers can mix CRUD methods with other op families and
// arbitrary custom methods.
package crud
