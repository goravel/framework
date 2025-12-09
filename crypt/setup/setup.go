package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	cryptFacadePath := path.Facades("crypt.go")
	cryptServiceProvider := "&crypt.ServiceProvider{}"
	modulePath := setup.ModulePath()

	setup.Install(
		// Add the crypt service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(modulePath, cryptServiceProvider),

		// Add the Crypt facade
		modify.File(cryptFacadePath).Overwrite(stubs.CryptFacade()),
	).Uninstall(
		// Remove the crypt service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(modulePath, cryptServiceProvider),

		// Remove the Crypt facade
		modify.File(cryptFacadePath).Remove(),
	).Execute()
}
