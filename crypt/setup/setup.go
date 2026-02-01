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
	cryptFacadePath := path.Facade("crypt.go")
	cryptServiceProvider := "&crypt.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	setup.Install(
		// Add the crypt service provider to the providers array in bootstrap/providers.go
		modify.RegisterProvider(moduleImport, cryptServiceProvider),

		// Add the Crypt facade
		modify.File(cryptFacadePath).Overwrite(stubs.CryptFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		// Remove the crypt service provider from the providers array in bootstrap/providers.go
		modify.UnregisterProvider(moduleImport, cryptServiceProvider),

		// Remove the Crypt facade
		modify.File(cryptFacadePath).Remove(),
	).Execute()
}
