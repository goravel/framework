package main

import (
	"fmt"
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	eventFacade := "Event"
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	appServiceProviderPath := path.App("providers", "app_service_provider.go")
	registerSeeders := "facades.Event().Register(map[event.Event][]event.Listener{})"
	eventImport := "github.com/goravel/framework/contracts/event"
	facadesImport := fmt.Sprintf("%s/app/facades", moduleName)
	eventFacadePath := path.Facades("event.go")
	eventServiceProvider := "&event.ServiceProvider{}"
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			// Add the event service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, eventServiceProvider),

			// Add the Register method to the app service provider.
			modify.GoFile(appServiceProviderPath).
				Find(match.Imports()).Modify(modify.AddImport(eventImport)).
				Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
				Find(match.RegisterFunc()).Modify(modify.Add(registerSeeders)),

			// Add the Event facade.
			modify.WhenFacade(eventFacade, modify.File(eventFacadePath).Overwrite(stubs.EventFacade())),
		).
		Uninstall(
			// Remove the Event facade and service provider.
			modify.WhenFacade(eventFacade, modify.File(eventFacadePath).Remove()),

			modify.WhenNoFacades([]string{eventFacade},
				// Remove the Register method from the app service provider.
				modify.GoFile(appServiceProviderPath).
					Find(match.RegisterFunc()).Modify(modify.Remove(registerSeeders)).
					Find(match.Imports()).Modify(modify.RemoveImport(eventImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)),

				// Remove the event service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, eventServiceProvider),
			),
		).
		Execute()
}
