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
				Find(match.Providers()).Modify(modify.Register("&filesystem.ServiceProvider{}")),
			modify.File(path.Config("filesystems.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("Storage", modify.File(path.Facades("storage.go")).Overwrite(stubs.StorageFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Storage"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&filesystem.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("filesystems.go")).Remove(),
			),
			modify.WhenFacade("Storage", modify.File(path.Facades("storage.go")).Remove()),
		).
		Execute()
}
