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
	queueFacade := "Queue"
	databaseDriver := "database"
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	appServiceProviderPath := path.App("providers", "app_service_provider.go")
	registerJobs := "facades.Queue().Register([]queue.Job{})"
	queueImport := "github.com/goravel/framework/contracts/queue"
	facadesImport := fmt.Sprintf("%s/app/facades", moduleName)
	appConfigPath := path.Config("app.go")
	queueFacadePath := path.Facades("queue.go")
	queueConfigPath := path.Config("queue.go")
	queueServiceProvider := "&queue.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			modify.WhenFacade(queueFacade,
				// Add the queue service provider to the application service provider
				modify.GoFile(appConfigPath).
					Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
					Find(match.Providers()).Modify(modify.Register(queueServiceProvider)),

				// Add the queue configuration file
				modify.File(queueConfigPath).Overwrite(stubs.Config(moduleName)),

				// Add the Register method to the AppServiceProvider to register the queue jobs.
				modify.GoFile(appServiceProviderPath).
					Find(match.Imports()).Modify(modify.AddImport(queueImport)).
					Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
					Find(match.RegisterFunc()).Modify(modify.Add(registerJobs)),

				// Add the queue facade to the facades file
				modify.File(queueFacadePath).Overwrite(stubs.QueueFacade()),
			),

			// Add the database driver
			modify.WhenDriver(databaseDriver, modify.GoFile(queueConfigPath).Find(match.Config("queue")).Modify(modify.AddConfig("default", `"database"`))),
		).
		Uninstall(
			// Remove the queue facade
			modify.WhenFacade(queueFacade, modify.File(queueFacadePath).Remove()),

			// Remove the Register method from the AppServiceProvider
			modify.GoFile(appServiceProviderPath).
				Find(match.RegisterFunc()).Modify(modify.Remove(registerJobs)).
				Find(match.Imports()).Modify(modify.RemoveImport(queueImport)).
				Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)),

			modify.WhenNoFacades([]string{queueFacade},
				// Remove the queue service provider from the application service provider
				modify.GoFile(appConfigPath).
					Find(match.Providers()).Modify(modify.Unregister(queueServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),

				// Remove the queue configuration file
				modify.File(queueConfigPath).Remove(),
			),
		).
		Execute()
}
