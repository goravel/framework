package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	// config, err := supportfile.GetFrameworkContent("hash/setup/config/hashing.go")
	// if err != nil {
	// 	panic(err)
	// }

	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&hash.ServiceProvider{}")),
			// modify.File(path.Config("hashing.go")).Overwrite(config),
		).
		Uninstall(
			modify.GoFile(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.Unregister("&hash.ServiceProvider{}")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			// modify.File(path.Config("hashing.go")).Remove(),
		).
		Execute()
}
