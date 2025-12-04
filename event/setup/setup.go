package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	eventFacade := "Event"
	eventFacadePath := path.Facades("event.go")
	eventServiceProvider := "&event.ServiceProvider{}"
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			// Add the event service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, eventServiceProvider),

			// Add the Event facade.
			modify.WhenFacade(eventFacade, modify.File(eventFacadePath).Overwrite(stubs.EventFacade())),
		).
		Uninstall(
			// Remove the Event facade and service provider.
			modify.WhenFacade(eventFacade, modify.File(eventFacadePath).Remove()),

			modify.WhenNoFacades([]string{eventFacade},
				// Remove the event service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, eventServiceProvider),
			),
		).
		Execute()
}
