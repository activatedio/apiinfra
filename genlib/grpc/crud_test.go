package grpc_test

import (
	"bytes"
	"github.com/activatedio/apiinfra/genlib/grpc"
	"github.com/activatedio/protogen"
	"github.com/activatedio/protogen/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuildCrud(t *testing.T) {

	a := assert.New(t)
	r := require.New(t)

	cases := []struct {
		name             string
		params           grpc.CrudParams
		expectedService  string
		expectedMessages string
	}{
		{
			name: "simple",
			params: grpc.CrudParams{
				Message:        proto.NewMessage("Unit"),
				MessagesTarget: proto.NewFile("unit"),
				ServiceTarget:  proto.NewService("UnitService"),
			},
			expectedService: `service UnitService {
  rpc GetUnitFoo (GetUnitRequest) returns (GetUnitResponse) {
  }
  rpc ListUnit (ListUnitRequest) returns (ListUnitResponse) {
  }
  rpc CreateUnit (CreateUnitRequest) returns (CreateUnitResponse) {
  }
  rpc UpdateUnit (UpdateUnitRequest) returns (UpdateUnitResponse) {
  }
  rpc PatchUnit (PatchUnitRequest) returns (PatchUnitResponse) {
  }
  rpc DeleteUnit (DeleteUnitRequest) returns (DeleteUnitResponse) {
  }
}

`,
			expectedMessages: `syntax = "proto3";

package unit;

message GetUnitRequest {
}

message GetUnitResponse {
}

message ListUnitRequest {
}

message ListUnitResponse {
}

message CreateUnitRequest {
}

message CreateUnitResponse {
}

message UpdateUnitRequest {
}

message UpdateUnitResponse {
}

message PatchUnitRequest {
}

message PatchUnitResponse {
}

message DeleteUnitRequest {
}

message DeleteUnitResponse {
}

`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			grpc.BuildCrud(tt.params)

			bufs := &bytes.Buffer{}
			r.NoError(tt.params.ServiceTarget.Render(protogen.NewWriterOutput(bufs)))

			bufm := &bytes.Buffer{}
			r.NoError(tt.params.MessagesTarget.Write(bufm))

			a.Equal(tt.expectedService, bufs.String(), "services are equal")
			a.Equal(tt.expectedMessages, bufm.String(), "messages are equal")

		})
	}
}
