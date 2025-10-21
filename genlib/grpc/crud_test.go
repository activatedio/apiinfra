package grpc_test

import (
	"bytes"
	"testing"

	"github.com/activatedio/apiinfra/genlib/grpc"
	"github.com/activatedio/protogen"
	"github.com/activatedio/protogen/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				Message:        proto.NewMessage("Unit").SetPackageName("unit.api"),
				MessagesTarget: proto.NewFile("unit"),
				ServiceTarget:  proto.NewService("UnitService"),
			},
			expectedService: `service UnitService {
  rpc GetUnitFoo (GetUnitRequest) returns (GetUnitResponse) {
    option (google.api.http) = {
    };
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

import "google/protobuf/field_mask.proto";
import "google/protobuf/empty.proto";

message GetUnitRequest {
  string name = 1;
  repeated string fields = 2;
}

message ListUnitRequest {
  string parent = 1;
  repeated string fields = 2;
  int32 page_size = 3;
  string page_token = 4;
  string selection = 5;
}

message ListUnitResponse {
  repeated unit.api.Unit list = 1;
  string next_page_token = 2;
}

message CreateUnitRequest {
  string parent = 1;
  unit.api.Unit unit = 2;
}

message UpdateUnitRequest {
  name parent = 1;
  unit.api.Unit unit = 2;
}

message PatchUnitRequest {
  name parent = 1;
  unit.api.Unit unit = 2;
  google.protobuf.FieldMask field_mask = 3;
}

message DeleteUnitRequest {
  name parent = 1;
}

`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(_ *testing.T) {

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
