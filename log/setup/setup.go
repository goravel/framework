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
				Find(match.Providers()).Modify(modify.Register("&log.ServiceProvider{}")),
			modify.File(path.Config("logging.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("Log", modify.File(path.Facades("log.go")).Overwrite(stubs.LogFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Log"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&log.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("logging.go")).Remove(),
			),
			modify.WhenFacade("Log", modify.File(path.Facades("log.go")).Remove()),
		).
		Execute()
}
