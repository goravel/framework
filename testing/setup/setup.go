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
				Find(match.Providers()).Modify(modify.Register("&testing.ServiceProvider{}")),
			modify.WhenFacade("Testing", modify.File(path.Facades("testing.go")).Overwrite(stubs.TestingFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Testing"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&testing.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade("Testing", modify.File(path.Facades("testing.go")).Remove()),
		).
		Execute()
}
