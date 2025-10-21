package grpc

import "github.com/activatedio/protogen/proto"

// RequestResponse represents a pair of proto messages for a request and its corresponding response.
type RequestResponse struct {
	Request  proto.Message
	Response proto.Message
}

func (rr RequestResponse) Messages() []proto.Message {
	return []proto.Message{rr.Request, rr.Response}
}
