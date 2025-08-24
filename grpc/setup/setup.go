package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	// config, err := supportfile.GetFrameworkContent("grpc/setup/config/grpc.go")
	// if err != nil {
	// 	panic(err)
	// }

	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&grpc.ServiceProvider{}")),
			// modify.File(path.Config("grpc.go")).Overwrite(config),
		).
		Uninstall(
			modify.GoFile(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.Unregister("&grpc.ServiceProvider{}")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			// modify.File(path.Config("grpc.go")).Remove(),
		).
		Execute()
}
