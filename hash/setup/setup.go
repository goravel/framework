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
				Find(match.Providers()).Modify(modify.Register("&hash.ServiceProvider{}")),
			modify.File(path.Config("hashing.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("Hash", modify.File(path.Facades("hash.go")).Overwrite(stubs.HashFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Hash"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&hash.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("hashing.go")).Remove(),
			),
			modify.WhenFacade("Hash", modify.File(path.Facades("hash.go")).Remove()),
		).
		Execute()
}
