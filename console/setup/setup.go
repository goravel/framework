package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	artisanFacadePath := path.Facade("artisan.go")
	consoleServiceProvider := "&console.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	setup.Install(
		// Add the console service provider to the providers array in bootstrap/providers.go
		modify.RegisterProvider(moduleImport, consoleServiceProvider),

		// Create the Artisan facade file.
		modify.File(artisanFacadePath).Overwrite(Stubs{}.ArtisanFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		// Remove the Artisan facade file.
		modify.File(artisanFacadePath).Remove(),

		// Remove the console service provider from the providers array in bootstrap/providers.go
		modify.UnregisterProvider(moduleImport, consoleServiceProvider),
	).Execute()
}
