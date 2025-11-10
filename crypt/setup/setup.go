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
	providersBootstrapPath := path.Bootstrap("providers.go")
	cryptFacadePath := path.Facades("crypt.go")
	cryptServiceProvider := "&crypt.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			modify.GoFile(providersBootstrapPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(cryptServiceProvider)),
			modify.WhenFacade(facades.Crypt, modify.File(cryptFacadePath).Overwrite(stubs.CryptFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Crypt},
				modify.GoFile(providersBootstrapPath).
					Find(match.Providers()).Modify(modify.Unregister(cryptServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade(facades.Crypt, modify.File(cryptFacadePath).Remove()),
		).
		Execute()
}
