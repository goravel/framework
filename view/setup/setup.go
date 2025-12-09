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
	modulePath := setup.ModulePath()
	viewFacadePath := path.Facades("view.go")

	setup.Install(
		// Add the view service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(modulePath, viewServiceProvider),

		// Add the View facade
		modify.File(viewFacadePath).Overwrite(stubs.ViewFacade()),
	).Uninstall(
		// Remove the view service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(modulePath, viewServiceProvider),

		// Remove the View facade
		modify.File(viewFacadePath).Remove(),
	).Execute()
}
