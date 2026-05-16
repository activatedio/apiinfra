package main

import (
	"log"

	"github.com/activatedio/apiinfra/genlib"
	"github.com/activatedio/apiinfra/genlib/grpc"
	"github.com/activatedio/apiinfra/genlib/grpc/buf"
	"github.com/activatedio/protogen/proto"
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

	b := grpc.NewCrudBuilder(grpc.CrudBuilderParams{
		MessagesTarget:      stf,
		ServiceTarget:       st,
		ServiceImportTarget: stf,
		APIBasePath:         "/v1",
	})

	ms := []grpc.CrudMessageParams{
		{
			Message: proto.NewMessage("Product").SetPackageName("example.api"),
		},
		{
			Message:    proto.NewMessage("ProductReview").SetPackageName("example.api"),
			ParentPath: "products/*",
		},
	}

	for _, m := range ms {
		mt.AddMessages(m.Message)
		b.BuildCrud(m)
	}

	bf := buf.NewFile(buf.FileParams{
		WellKnownVersion: "v33.0",
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
