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
				Find(match.Providers()).Modify(modify.Register("&schedule.ServiceProvider{}")),
			modify.WhenFacade("Schedule", modify.File(path.Facades("schedule.go")).Overwrite(stubs.ScheduleFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Schedule"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&schedule.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade("Schedule", modify.File(path.Facades("schedule.go")).Remove()),
		).
		Execute()
}
