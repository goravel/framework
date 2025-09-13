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
				Find(match.Providers()).Modify(modify.Register("&auth.ServiceProvider{}")),
			modify.File(path.Config("auth.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("Auth", modify.File(path.Facades("auth.go")).Overwrite(stubs.AuthFacade())),
			modify.WhenFacade("Gate", modify.File(path.Facades("gate.go")).Overwrite(stubs.GateFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Auth", "Gate"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&auth.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("auth.go")).Remove(),
			),
			modify.WhenFacade("Auth", modify.File(path.Facades("auth.go")).Remove()),
			modify.WhenFacade("Gate", modify.File(path.Facades("gate.go")).Remove()),
		).
		Execute()
}
