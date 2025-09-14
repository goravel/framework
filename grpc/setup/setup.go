package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}

	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&grpc.ServiceProvider{}")),
			modify.File(path.Config("grpc.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("Grpc", modify.File(path.Facades("grpc.go")).Overwrite(stubs.GrpcFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Grpc"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&grpc.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("grpc.go")).Remove(),
			),
			modify.WhenFacade("Grpc", modify.File(path.Facades("grpc.go")).Remove()),
		).
		Execute()
}
