package main

import (
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
	queueFacadePath := path.Facades("queue.go")
	queueConfigPath := path.Config("queue.go")
	queueServiceProvider := "&queue.ServiceProvider{}"
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			modify.WhenFacade(queueFacade,
				// Add the queue service provider to the providers array in bootstrap/providers.go
				modify.AddProviderApply(modulePath, queueServiceProvider),

				// Add the queue configuration file
				modify.File(queueConfigPath).Overwrite(stubs.Config(moduleName)),

				// Add the queue facade to the facades file
				modify.File(queueFacadePath).Overwrite(stubs.QueueFacade()),
			),

			// Add the database driver
			modify.WhenDriver(databaseDriver, modify.GoFile(queueConfigPath).Find(match.Config("queue")).Modify(modify.AddConfig("default", `"database"`))),
		).
		Uninstall(
			// Remove the queue facade
			modify.WhenFacade(queueFacade, modify.File(queueFacadePath).Remove()),

			modify.WhenNoFacades([]string{queueFacade},
				// Remove the queue configuration file
				modify.File(queueConfigPath).Remove(),

				// Remove the queue service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, queueServiceProvider),
			),
		).
		Execute()
}
