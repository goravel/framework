package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	viewServiceProvider := "&view.ServiceProvider{}"
	modulePath := packages.GetModulePath()
	viewFacade := "View"
	viewFacadePath := path.Facades("view.go")

	packages.Setup(os.Args).
		Install(
			// Add the view service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, viewServiceProvider),

			// Add the View facade
			modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Overwrite(stubs.ViewFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{viewFacade},
				// Remove the view service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, viewServiceProvider),
			),
			modify.WhenFacade(viewFacade, modify.File(viewFacadePath).Remove()),
		).
		Execute()
}
