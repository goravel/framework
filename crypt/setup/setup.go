package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	cryptFacadePath := path.Facades("crypt.go")
	cryptServiceProvider := "&crypt.ServiceProvider{}"
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			// Add the crypt service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, cryptServiceProvider),

			// Add the Crypt facade
			modify.WhenFacade(facades.Crypt, modify.File(cryptFacadePath).Overwrite(stubs.CryptFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Crypt},
				// Remove the crypt service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, cryptServiceProvider),
			),

			// Remove the Crypt facade
			modify.WhenFacade(facades.Crypt, modify.File(cryptFacadePath).Remove()),
		).
		Execute()
}
