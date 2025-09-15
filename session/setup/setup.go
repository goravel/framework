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
				Find(match.Providers()).Modify(modify.Register("&session.ServiceProvider{}")),
			modify.File(path.Config("session.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("Session", modify.File(path.Facades("session.go")).Overwrite(stubs.SessionFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Session"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&session.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("session.go")).Remove(),
			),
			modify.WhenFacade("Session", modify.File(path.Facades("session.go")).Remove()),
		).
		Execute()
}
