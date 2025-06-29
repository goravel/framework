package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&event.ServiceProvider{}")),
		).
		Uninstall(
			modify.GoFile(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.Unregister("&event.ServiceProvider{}")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
		).
		Execute()
}
