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
				Find(match.Providers()).Modify(modify.Register("&validation.ServiceProvider{}")),
			modify.WhenFacade("Validation", modify.File(path.Facades("validation.go")).Overwrite(stubs.ValidationFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Validation"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&validation.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade("Validation", modify.File(path.Facades("validation.go")).Remove()),
		).
		Execute()
}
