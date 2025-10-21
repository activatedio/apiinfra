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
  rpc GetUnit (GetUnitRequest) returns (GetUnitResponse) {
    option (google.api.http) = {
      get: "/{name=units/*}"
    };
  }
  rpc ListUnit (ListUnitRequest) returns (ListUnitResponse) {
    option (google.api.http) = {
      get: "/units/*"
    };
  }
  rpc CreateUnit (CreateUnitRequest) returns (CreateUnitResponse) {
    option (google.api.http) = {
      post: "/units/*"
      body: "*"
    };
  }
  rpc UpdateUnit (UpdateUnitRequest) returns (UpdateUnitResponse) {
    option (google.api.http) = {
      put: "/{name=units/*}"
      body: "*"
    };
  }
  rpc PatchUnit (PatchUnitRequest) returns (PatchUnitResponse) {
    option (google.api.http) = {
      patch: "/{name=units/*}"
      body: "*"
    };
  }
  rpc DeleteUnit (DeleteUnitRequest) returns (DeleteUnitResponse) {
    option (google.api.http) = {
      delete: "/{name=units/*}"
    };
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
		{
			name: "full",
			params: grpc.CrudParams{
				Message:        proto.NewMessage("ModifiedUnit").SetPackageName("modified_unit.api"),
				MessagesTarget: proto.NewFile("modified_unit"),
				ServiceTarget:  proto.NewService("ModifiedUnitService"),
				ParentPath:     "/tenants/*/",
				APIBasePath:    "/api/v2",
			},
			expectedService: `service ModifiedUnitService {
  rpc GetModifiedUnit (GetModifiedUnitRequest) returns (GetModifiedUnitResponse) {
    option (google.api.http) = {
      get: "/api/v2/{name=tenants/*/modified_units/*}"
    };
  }
  rpc ListModifiedUnit (ListModifiedUnitRequest) returns (ListModifiedUnitResponse) {
    option (google.api.http) = {
      get: "/api/v2/{parent=tenants/*}/modified_units/*"
    };
  }
  rpc CreateModifiedUnit (CreateModifiedUnitRequest) returns (CreateModifiedUnitResponse) {
    option (google.api.http) = {
      post: "/api/v2/{parent=tenants/*}/modified_units/*"
      body: "*"
    };
  }
  rpc UpdateModifiedUnit (UpdateModifiedUnitRequest) returns (UpdateModifiedUnitResponse) {
    option (google.api.http) = {
      put: "/api/v2/{name=tenants/*/modified_units/*}"
      body: "*"
    };
  }
  rpc PatchModifiedUnit (PatchModifiedUnitRequest) returns (PatchModifiedUnitResponse) {
    option (google.api.http) = {
      patch: "/api/v2/{name=tenants/*/modified_units/*}"
      body: "*"
    };
  }
  rpc DeleteModifiedUnit (DeleteModifiedUnitRequest) returns (DeleteModifiedUnitResponse) {
    option (google.api.http) = {
      delete: "/api/v2/{name=tenants/*/modified_units/*}"
    };
  }
}

`,
			expectedMessages: `syntax = "proto3";

package modified_unit;

import "google/protobuf/field_mask.proto";
import "google/protobuf/empty.proto";

message GetModifiedUnitRequest {
  string name = 1;
  repeated string fields = 2;
}

message ListModifiedUnitRequest {
  string parent = 1;
  repeated string fields = 2;
  int32 page_size = 3;
  string page_token = 4;
  string selection = 5;
}

message ListModifiedUnitResponse {
  repeated modified_unit.api.ModifiedUnit list = 1;
  string next_page_token = 2;
}

message CreateModifiedUnitRequest {
  string parent = 1;
  modified_unit.api.ModifiedUnit modified_unit = 2;
}

message UpdateModifiedUnitRequest {
  name parent = 1;
  modified_unit.api.ModifiedUnit modified_unit = 2;
}

message PatchModifiedUnitRequest {
  name parent = 1;
  modified_unit.api.ModifiedUnit modified_unit = 2;
  google.protobuf.FieldMask field_mask = 3;
}

message DeleteModifiedUnitRequest {
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
