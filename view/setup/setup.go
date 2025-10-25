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
				Find(match.Providers()).Modify(modify.Register("&view.ServiceProvider{}")),
			modify.WhenFacade("View", modify.File(path.Facades("view.go")).Overwrite(stubs.ViewFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"View"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&view.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade("View", modify.File(path.Facades("view.go")).Remove()),
		).
		Execute()
}
