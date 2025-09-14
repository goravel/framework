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
				Find(match.Providers()).Modify(modify.Register("&crypt.ServiceProvider{}")),
			modify.WhenFacade("Crypt", modify.File(path.Facades("crypt.go")).Overwrite(stubs.CryptFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Crypt"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister("&crypt.ServiceProvider{}")).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade("Crypt", modify.File(path.Facades("crypt.go")).Remove()),
		).
		Execute()
}
