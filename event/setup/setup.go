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
	eventFacadePath := path.Facades("event.go")
	eventServiceProvider := "&event.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	setup.Install(
		// Add the event service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, eventServiceProvider),

		// Add the Event facade.
		modify.File(eventFacadePath).Overwrite(stubs.EventFacade(setup.Paths().Facades().Import())),
	).Uninstall(
		// Remove the Event facade and service provider.
		modify.File(eventFacadePath).Remove(),

		// Remove the event service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, eventServiceProvider),
	).Execute()
}
