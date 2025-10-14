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
				Find(match.Providers()).Modify(modify.Register("&translation.ServiceProvider{}")),
			modify.WhenFacade("Lang", modify.File(path.Facades("lang.go")).Overwrite(Stubs{}.LangFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Lang"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&translation.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade("Lang", modify.File(path.Facades("lang.go")).Remove()),
		).
		Execute()
}
