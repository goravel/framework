package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	validationFacadePath := path.Facades("validation.go")
	validationServiceProvider := "&validation.ServiceProvider{}"
	modulePath := setup.ModulePath()

	setup.Install(
		// Add the validation service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(modulePath, validationServiceProvider),

		// Add the Validation facade
		modify.File(validationFacadePath).Overwrite(stubs.ValidationFacade()),
	).Uninstall(
		modify.WhenNoFacades([]string{facades.Validation},
			// Remove the validation service provider from the providers array in bootstrap/providers.go
			modify.RemoveProviderApply(modulePath, validationServiceProvider),
		),

		// Remove the Validation facade
		modify.File(validationFacadePath).Remove(),
	).Execute()
}
