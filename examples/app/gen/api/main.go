// Package main runs the example proto generation.
package main

import (
	"log"

	"github.com/activatedio/apiinfra/genlib"
	"github.com/activatedio/apiinfra/genlib/grpc"
	"github.com/activatedio/apiinfra/genlib/grpc/buf"
	"github.com/activatedio/apiinfra/genlib/grpc/crud"
	"github.com/activatedio/protogen/proto"
	"github.com/activatedio/protogen/tfl"
)

//go:generate go run .

const (
	GoPackageOptionName = "go_package"
	GoPackage           = "github.com/activatedio/apiinfra/examples/app/pb"
)

func main() {

	pkgOption := proto.NewOption(GoPackageOptionName,
		proto.NewStringConstant(GoPackage))

	mt := proto.NewFile("example.api").AddOptions(pkgOption)
	st := proto.NewService("ExampleApiService")
	stf := proto.NewFile("example.api").AddOptions(pkgOption).
		AddImports(proto.NewImport("types.proto"))
	stf.AddServices(st)

	product := proto.NewMessage("Product").SetPackageName("example.api")
	productReview := proto.NewMessage("ProductReview").SetPackageName("example.api")

	mt.AddMessages(product, productReview)

	sb := grpc.NewServiceBuilder(grpc.Target{
		Messages:       stf,
		Service:        st,
		ServiceImports: stf,
		APIBasePath:    "/v1",
	})

	sb.Add(
		crud.All(crud.Resource{Message: product}),
		crud.All(crud.Resource{
			Message:    productReview,
			ParentPath: "products/*",
		}),
		// Escape hatch: a custom action on Product that doesn't fit the CRUD
		// shape. The caller reaches into protogen directly through the
		// supplied grpc.Target.
		grpc.OperationFunc(func(t grpc.Target) {
			t.Service.AddMethods(
				proto.NewMethod("ArchiveProduct", proto.MethodParams{
					RequestName:  "ArchiveProductRequest",
					ResponseName: "example.api.Product",
				}).AddOptions(
					proto.NewOption("google.api.http", proto.NewMessageValueConstant(
						tfl.NewMessageValue().AddFields(
							tfl.NewStringField("post", "/v1/{name=products/*}:archive"),
							tfl.NewStringField("body", "*"),
						),
					)),
				),
			)
			t.Messages.AddMessages(
				proto.NewMessage("ArchiveProductRequest").AddFields(
					proto.NewField("name", proto.FieldParams{
						FieldType: "string",
						Number:    1,
					}),
				),
			)
		}),
	)

	bf := buf.NewFile(buf.FileParams{
		WellKnownVersion:  "v33.0",
		GoogleAPIsVersion: "72c8614f3bd0466ea67931ef2c43d608",
		Modules: []buf.Module{
			{
				Path: "example",
			},
		},
		ProtocolBuffersGoVersion: "v1.36.10",
		GrpcGoVersion:            "v1.5.1",
		GrpcGatewayVersion:       "v2.27.3",
		GoOutputPath:             "../pb",
	})

	wfm := genlib.WritableFile("../../proto/example/types.proto")
	defer genlib.CheckClose(wfm)
	wfs := genlib.WritableFile("../../proto/example/grpc.proto")
	defer genlib.CheckClose(wfs)
	wfb := genlib.WritableFile("../../proto/buf.yaml")
	defer genlib.CheckClose(wfb)
	wfbg := genlib.WritableFile("../../proto/buf.gen.yaml")
	defer genlib.CheckClose(wfbg)
	wfgg := genlib.WritableFile("../../proto/gen.go")
	defer genlib.CheckClose(wfgg)

	check(mt.Write(wfm))
	check(stf.Write(wfs))
	check(bf.WriteBufYAML(wfb))
	check(bf.WriteBufGenYAML(wfbg))
	check(bf.WriteGenGo(wfgg))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
