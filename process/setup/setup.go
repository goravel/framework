package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	appConfigPath := path.Config("app.go")
	processFacadePath := path.Facades("process.go")
	modulePath := packages.GetModulePath()
	processServiceProvider := "&process.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			modify.GoFile(appConfigPath).
				Find(match.Imports()).Modify(modify.AddImport(modulePath)).
				Find(match.Providers()).Modify(modify.Register(processServiceProvider)),
			modify.WhenFacade(facades.Process, modify.File(processFacadePath).Overwrite(stubs.ProcessFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Process},
				modify.GoFile(appConfigPath).
					Find(match.Providers()).Modify(modify.Unregister(processServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(modulePath)),
			),
			modify.WhenFacade(facades.Process, modify.File(processFacadePath).Remove()),
		).
		Execute()
}
