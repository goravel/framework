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
	appConfigPath := path.Config("app.go")
	eventFacadePath := path.Facades("event.go")
	eventServiceProvider := "&event.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			// Add the Event facade and service provider.
			modify.GoFile(appConfigPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(eventServiceProvider)),

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

				// Remove the Event service provider from the app config.
				modify.GoFile(appConfigPath).
					Find(match.Providers()).Modify(modify.Unregister(eventServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
		).
		Execute()
}
