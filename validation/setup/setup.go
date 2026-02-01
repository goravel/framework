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
	validationFacadePath := path.Facade("validation.go")
	validationServiceProvider := "&validation.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	setup.Install(
		// Add the validation service provider to the providers array in bootstrap/providers.go
		modify.RegisterProvider(moduleImport, validationServiceProvider),

		// Add the Validation facade
		modify.File(validationFacadePath).Overwrite(stubs.ValidationFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		// Remove the validation service provider from the providers array in bootstrap/providers.go
		modify.UnregisterProvider(moduleImport, validationServiceProvider),

		// Remove the Validation facade
		modify.File(validationFacadePath).Remove(),
	).Execute()
}
