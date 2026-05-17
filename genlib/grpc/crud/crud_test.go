package crud_test

import (
	"bytes"
	"testing"

	"github.com/activatedio/apiinfra/genlib/grpc"
	"github.com/activatedio/apiinfra/genlib/grpc/crud"
	"github.com/activatedio/protogen"
	"github.com/activatedio/protogen/proto"
	"github.com/activatedio/protogen/tfl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {

	cases := []struct {
		name                  string
		resource              crud.Resource
		apiBasePath           string
		serviceName           string
		messagesPackage       string
		expectedService       string
		expectedServiceImport string
		expectedMessages      string
	}{
		{
			name:            "simple",
			serviceName:     "UnitService",
			messagesPackage: "unit",
			resource: crud.Resource{
				Message: proto.NewMessage("Unit").SetPackageName("unit.api"),
			},
			expectedService: `service UnitService {
  rpc GetUnit (GetUnitRequest) returns (unit.api.Unit) {
    option (google.api.http) = {
      get: "/{name=units/*}"
    };
  }
  rpc ListUnit (ListUnitRequest) returns (ListUnitResponse) {
    option (google.api.http) = {
      get: "/units"
    };
  }
  rpc CreateUnit (CreateUnitRequest) returns (unit.api.Unit) {
    option (google.api.http) = {
      post: "/units"
      body: "*"
    };
  }
  rpc UpdateUnit (UpdateUnitRequest) returns (unit.api.Unit) {
    option (google.api.http) = {
      put: "/{name=units/*}"
      body: "*"
    };
  }
  rpc PatchUnit (PatchUnitRequest) returns (unit.api.Unit) {
    option (google.api.http) = {
      patch: "/{name=units/*}"
      body: "*"
    };
  }
  rpc DeleteUnit (DeleteUnitRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/{name=units/*}"
    };
  }
}

`,
			expectedServiceImport: `syntax = "proto3";

package services.proto;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";

`,
			expectedMessages: `syntax = "proto3";

package unit;

import "google/protobuf/field_mask.proto";

message GetUnitRequest {
  string name = 1;
  google.protobuf.FieldMask read_mask = 2;
}

message ListUnitRequest {
  string parent = 1;
  int32 page_size = 2;
  string page_token = 3;
  string filter = 4;
  string order_by = 5;
  google.protobuf.FieldMask read_mask = 6;
}

message ListUnitResponse {
  repeated unit.api.Unit units = 1;
  string next_page_token = 2;
}

message CreateUnitRequest {
  string parent = 1;
  unit.api.Unit unit = 2;
}

message UpdateUnitRequest {
  string name = 1;
  unit.api.Unit unit = 2;
}

message PatchUnitRequest {
  string name = 1;
  unit.api.Unit unit = 2;
  google.protobuf.FieldMask update_mask = 3;
}

message DeleteUnitRequest {
  string name = 1;
}

`,
		},
		{
			name:            "full",
			serviceName:     "ModifiedUnitService",
			messagesPackage: "modified_unit",
			apiBasePath:     "/api/v2",
			resource: crud.Resource{
				Message:    proto.NewMessage("ModifiedUnit").SetPackageName("modified_unit.api"),
				ParentPath: "/tenants/*/",
			},
			expectedService: `service ModifiedUnitService {
  rpc GetModifiedUnit (GetModifiedUnitRequest) returns (modified_unit.api.ModifiedUnit) {
    option (google.api.http) = {
      get: "/api/v2/{name=tenants/*/modified_units/*}"
    };
  }
  rpc ListModifiedUnit (ListModifiedUnitRequest) returns (ListModifiedUnitResponse) {
    option (google.api.http) = {
      get: "/api/v2/{parent=tenants/*}/modified_units"
    };
  }
  rpc CreateModifiedUnit (CreateModifiedUnitRequest) returns (modified_unit.api.ModifiedUnit) {
    option (google.api.http) = {
      post: "/api/v2/{parent=tenants/*}/modified_units"
      body: "*"
    };
  }
  rpc UpdateModifiedUnit (UpdateModifiedUnitRequest) returns (modified_unit.api.ModifiedUnit) {
    option (google.api.http) = {
      put: "/api/v2/{name=tenants/*/modified_units/*}"
      body: "*"
    };
  }
  rpc PatchModifiedUnit (PatchModifiedUnitRequest) returns (modified_unit.api.ModifiedUnit) {
    option (google.api.http) = {
      patch: "/api/v2/{name=tenants/*/modified_units/*}"
      body: "*"
    };
  }
  rpc DeleteModifiedUnit (DeleteModifiedUnitRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/api/v2/{name=tenants/*/modified_units/*}"
    };
  }
}

`,
			expectedServiceImport: `syntax = "proto3";

package services.proto;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";

`,
			expectedMessages: `syntax = "proto3";

package modified_unit;

import "google/protobuf/field_mask.proto";

message GetModifiedUnitRequest {
  string name = 1;
  google.protobuf.FieldMask read_mask = 2;
}

message ListModifiedUnitRequest {
  string parent = 1;
  int32 page_size = 2;
  string page_token = 3;
  string filter = 4;
  string order_by = 5;
  google.protobuf.FieldMask read_mask = 6;
}

message ListModifiedUnitResponse {
  repeated modified_unit.api.ModifiedUnit modified_units = 1;
  string next_page_token = 2;
}

message CreateModifiedUnitRequest {
  string parent = 1;
  modified_unit.api.ModifiedUnit modified_unit = 2;
}

message UpdateModifiedUnitRequest {
  string name = 1;
  modified_unit.api.ModifiedUnit modified_unit = 2;
}

message PatchModifiedUnitRequest {
  string name = 1;
  modified_unit.api.ModifiedUnit modified_unit = 2;
  google.protobuf.FieldMask update_mask = 3;
}

message DeleteModifiedUnitRequest {
  string name = 1;
}

`,
		},
		{
			name:            "partial: get/list/delete only",
			serviceName:     "UnitService",
			messagesPackage: "unit",
			resource: crud.Resource{
				Message: proto.NewMessage("Unit").SetPackageName("unit.api"),
				Ops:     crud.OpGet | crud.OpList | crud.OpDelete,
			},
			expectedService: `service UnitService {
  rpc GetUnit (GetUnitRequest) returns (unit.api.Unit) {
    option (google.api.http) = {
      get: "/{name=units/*}"
    };
  }
  rpc ListUnit (ListUnitRequest) returns (ListUnitResponse) {
    option (google.api.http) = {
      get: "/units"
    };
  }
  rpc DeleteUnit (DeleteUnitRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/{name=units/*}"
    };
  }
}

`,
			expectedServiceImport: `syntax = "proto3";

package services.proto;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";

`,
			expectedMessages: `syntax = "proto3";

package unit;

import "google/protobuf/field_mask.proto";

message GetUnitRequest {
  string name = 1;
  google.protobuf.FieldMask read_mask = 2;
}

message ListUnitRequest {
  string parent = 1;
  int32 page_size = 2;
  string page_token = 3;
  string filter = 4;
  string order_by = 5;
  google.protobuf.FieldMask read_mask = 6;
}

message ListUnitResponse {
  repeated unit.api.Unit units = 1;
  string next_page_token = 2;
}

message DeleteUnitRequest {
  string name = 1;
}

`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			a := assert.New(t)
			r := require.New(t)

			messages := proto.NewFile(tt.messagesPackage)
			service := proto.NewService(tt.serviceName)
			serviceImports := proto.NewFile("services.proto")

			sb := grpc.NewServiceBuilder(grpc.Target{
				Messages:       messages,
				Service:        service,
				ServiceImports: serviceImports,
				APIBasePath:    tt.apiBasePath,
			})
			sb.Add(crud.All(tt.resource))

			bufs := &bytes.Buffer{}
			r.NoError(service.Render(protogen.NewWriterOutput(bufs)))

			bufsi := &bytes.Buffer{}
			r.NoError(serviceImports.Write(bufsi))

			bufm := &bytes.Buffer{}
			r.NoError(messages.Write(bufm))

			a.Equal(tt.expectedService, bufs.String(), "services are equal")
			a.Equal(tt.expectedServiceImport, bufsi.String(), "service imports are equal")
			a.Equal(tt.expectedMessages, bufm.String(), "messages are equal")
		})
	}
}

// TestEscapeHatch exercises OperationFunc — a caller using protogen directly
// to add an arbitrary method alongside CRUD operations.
func TestEscapeHatch(t *testing.T) {

	a := assert.New(t)
	r := require.New(t)

	resource := crud.Resource{
		Message: proto.NewMessage("Unit").SetPackageName("unit.api"),
		Ops:     crud.OpGet,
	}

	messages := proto.NewFile("unit")
	service := proto.NewService("UnitService")
	serviceImports := proto.NewFile("services.proto")

	sb := grpc.NewServiceBuilder(grpc.Target{
		Messages:       messages,
		Service:        service,
		ServiceImports: serviceImports,
	})

	// First a CRUD op, then a custom method through the escape hatch.
	sb.Add(
		crud.Get(resource),
		grpc.OperationFunc(func(t grpc.Target) {
			t.Service.AddMethods(
				proto.NewMethod("ArchiveUnit", proto.MethodParams{
					RequestName:  "ArchiveUnitRequest",
					ResponseName: "unit.api.Unit",
				}).AddOptions(
					proto.NewOption("google.api.http", proto.NewMessageValueConstant(
						tfl.NewMessageValue().AddFields(
							tfl.NewStringField("post", "/{name=units/*}:archive"),
							tfl.NewStringField("body", "*"),
						),
					)),
				),
			)
			t.Messages.AddMessages(
				proto.NewMessage("ArchiveUnitRequest").AddFields(
					proto.NewField("name", proto.FieldParams{
						FieldType: "string",
						Number:    1,
					}),
				),
			)
		}),
	)

	bufs := &bytes.Buffer{}
	r.NoError(service.Render(protogen.NewWriterOutput(bufs)))

	bufm := &bytes.Buffer{}
	r.NoError(messages.Write(bufm))

	a.Contains(bufs.String(), "rpc GetUnit (GetUnitRequest)")
	a.Contains(bufs.String(), "rpc ArchiveUnit (ArchiveUnitRequest)")
	a.Contains(bufs.String(), `post: "/{name=units/*}:archive"`)
	a.Contains(bufm.String(), "message ArchiveUnitRequest")
	a.Contains(bufm.String(), "message GetUnitRequest")
}
