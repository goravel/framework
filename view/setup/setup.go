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
	viewServiceProvider := "&view.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	viewFacadePath := path.Facades("view.go")

	setup.Install(
		// Add the view service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, viewServiceProvider),

		// Add the View facade
		modify.File(viewFacadePath).Overwrite(stubs.ViewFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		// Remove the view service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, viewServiceProvider),

		// Remove the View facade
		modify.File(viewFacadePath).Remove(),
	).Execute()
}
