package buf

import "io"

// Module describes a buf module entry written to buf.yaml.
type Module struct {
	Name string
	Path string
}

// FileParams holds the configuration values used when rendering buf files.
//
// The *Version fields pin buf.build modules and plugins to specific releases
// (or commit hashes) so regeneration is deterministic. GrpcGatewayVersion
// pins the gateway plugin used in buf.gen.yaml — distinct from any
// grpc-gateway proto module, which the templates do not declare since the
// CRUD output does not import grpc-gateway-specific protos.
type FileParams struct {
	WellKnownVersion         string
	GoogleAPIsVersion        string
	Modules                  []Module
	ProtocolBuffersGoVersion string
	GrpcGoVersion            string
	GrpcGatewayVersion       string
	GoOutputPath             string
}

// File renders the set of buf configuration files (buf.yaml, buf.gen.yaml, gen.go).
type File interface {
	WriteBufYAML(w io.Writer) error
	WriteBufGenYAML(w io.Writer) error
	WriteGenGo(w io.Writer) error
}
