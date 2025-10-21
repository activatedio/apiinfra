package grpc

import (
	"strings"

	"github.com/activatedio/protogen/proto"
)

// GetQualifiedName constructs and returns the fully qualified name of a proto.Message by combining package and message names.
func GetQualifiedName(msg proto.Message) string {
	sb := strings.Builder{}

	if msg.GetPackageName() != "" {
		sb.WriteString(msg.GetPackageName())
		sb.WriteString(".")
	}

	sb.WriteString(msg.GetName())

	return sb.String()
}
