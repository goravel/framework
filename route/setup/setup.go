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
				Find(match.Providers()).Modify(modify.Register("&route.ServiceProvider{}")),
			modify.WhenFacade("Route", modify.File(path.Facades("route.go")).Overwrite(stubs.RouteFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Route"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&route.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade("Route", modify.File(path.Facades("route.go")).Remove()),
		).
		Execute()
}
