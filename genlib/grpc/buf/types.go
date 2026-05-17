package buf

import "io"

// Module describes a buf module entry written to buf.yaml.
type Module struct {
	Name string
	Path string
}

// FileParams holds the configuration values used when rendering buf files.
type FileParams struct {
	WellKnownVersion         string
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
