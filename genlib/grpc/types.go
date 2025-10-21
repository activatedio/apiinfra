package grpc

import "github.com/activatedio/protogen/proto"

// RequestResponse represents a pair of proto messages for a request and its corresponding response.
type RequestResponse struct {
	Request  proto.Message
	Response proto.Message
}

// Messages returns a slice containing the Request and Response proto messages from the RequestResponse struct.
func (rr RequestResponse) Messages() []proto.Message {
	return []proto.Message{rr.Request, rr.Response}
}
