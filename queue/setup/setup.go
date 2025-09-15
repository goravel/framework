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
				Find(match.Providers()).Modify(modify.Register("&queue.ServiceProvider{}")),
			modify.File(path.Config("queue.go")).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade("Queue", modify.File(path.Facades("queue.go")).Overwrite(stubs.QueueFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Queue"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&queue.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(path.Config("queue.go")).Remove(),
			),
			modify.WhenFacade("Queue", modify.File(path.Facades("queue.go")).Remove()),
		).
		Execute()
}
