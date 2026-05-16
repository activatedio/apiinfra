package buf

import "io"

type Module struct {
	Name string
	Path string
}

type FileParams struct {
	WellKnownVersion         string
	Modules                  []Module
	ProtocolBuffersGoVersion string
	GrpcGoVersion            string
	GrpcGatewayVersion       string
	GoOutputPath             string
}

type File interface {
	WriteBufYAML(w io.Writer) error
	WriteBufGenYAML(w io.Writer) error
	WriteGenGo(w io.Writer) error
}
