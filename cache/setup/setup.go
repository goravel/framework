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
				Find(match.Providers()).Modify(modify.Register("&cache.ServiceProvider{}")),
			modify.File(path.Config("cache.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("Cache", modify.File(path.Facades("cache.go")).Overwrite(stubs.CacheFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Cache"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&cache.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("cache.go")).Remove(),
			),
			modify.WhenFacade("Cache", modify.File(path.Facades("cache.go")).Remove()),
		).
		Execute()
}
