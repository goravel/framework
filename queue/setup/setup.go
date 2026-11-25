package main

import (
	"fmt"
	"os"

	"github.com/goravel/framework/contracts/facades"
	contractspackages "github.com/goravel/framework/contracts/packages"
	contractsmodify "github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)

	setup.Install(installActions(setup)...).Uninstall(uninstallActions(setup)...).Execute()
}

func installActions(setup contractspackages.Setup) []contractsmodify.Apply {
	stubs := Stubs{}
	databaseDriver := "database"
	queueFacadePath := path.Facade("queue.go")
	queueConfigPath := path.Config("queue.go")
	queueServiceProvider := "&queue.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesImport := setup.Paths().Facades().Import()
	migrationPkg := setup.Paths().Migrations().Package()
	migrationPkgPath := setup.Paths().Migrations().Import()
	facadesPackage := setup.Paths().Facades().Package()
	jobMigrationFileName, jobMigrationStruct, jobMigrationContent := stubs.JobMigration(migrationPkg, facadesImport, facadesPackage)
	jobMigrationFilePath := path.Migration(jobMigrationFileName)
	jobMigrationStructWithPkg := fmt.Sprintf("&%s.%s", migrationPkg, jobMigrationStruct)
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")

	return []contractsmodify.Apply{
		// Avoid duplicate installation when installing drivers
		modify.WhenFacade(facades.Queue,
			// Add the queue service provider to the providers array in bootstrap/providers.go
			modify.RegisterProvider(moduleImport, queueServiceProvider),

			// Add the queue configuration file
			modify.File(queueConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), facadesImport, facadesPackage)),

			// Add the queue facade to the facades file
			modify.File(queueFacadePath).Overwrite(stubs.QueueFacade(facadesPackage)),

			// Add the job migration file
			modify.File(jobMigrationFilePath).Overwrite(jobMigrationContent),

			// Register the job migration
			modify.RegisterMigration(migrationPkgPath, jobMigrationStructWithPkg),

			// Add the queue connection to the .env and .env.example files
			modify.WhenFileExists(envPath, modify.EnvWhenMissing(envPath, "QUEUE_CONNECTION", "sync")),
			modify.WhenFileExists(envExamplePath, modify.EnvWhenMissing(envExamplePath, "QUEUE_CONNECTION", "sync")),
		),

		// Selecting the sync driver overwrites the selector, so a stale value cannot survive.
		modify.WhenDriver("sync",
			modify.WhenFileExists(envPath, modify.Env(envPath, "QUEUE_CONNECTION", "sync")),
			modify.WhenFileExists(envExamplePath, modify.Env(envExamplePath, "QUEUE_CONNECTION", "sync")),
		),

		// Add the database driver and record the selection.
		modify.WhenDriver(databaseDriver,
			modify.GoFile(queueConfigPath).Find(match.Config("queue")).Modify(modify.AddConfig("default", `config.Env("QUEUE_CONNECTION", "database")`)),
			modify.WhenFileExists(envPath, modify.Env(envPath, "QUEUE_CONNECTION", "database")),
			modify.WhenFileExists(envExamplePath, modify.Env(envExamplePath, "QUEUE_CONNECTION", "database")),
		),
	}
}

func uninstallActions(setup contractspackages.Setup) []contractsmodify.Apply {
	stubs := Stubs{}
	queueFacadePath := path.Facade("queue.go")
	queueConfigPath := path.Config("queue.go")
	queueServiceProvider := "&queue.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesImport := setup.Paths().Facades().Import()
	migrationPkg := setup.Paths().Migrations().Package()
	migrationPkgPath := setup.Paths().Migrations().Import()
	facadesPackage := setup.Paths().Facades().Package()
	jobMigrationFileName, jobMigrationStruct, _ := stubs.JobMigration(migrationPkg, facadesImport, facadesPackage)
	jobMigrationFilePath := path.Migration(jobMigrationFileName)
	jobMigrationStructWithPkg := fmt.Sprintf("&%s.%s", migrationPkg, jobMigrationStruct)

	return []contractsmodify.Apply{
		modify.WhenFacade(facades.Queue,
			// Unregister the job migration
			modify.UnregisterMigration(migrationPkgPath, jobMigrationStructWithPkg),

			// Remove the job migration file
			modify.File(jobMigrationFilePath).Remove(),

			// Remove the queue facade
			modify.File(queueFacadePath).Remove(),

			// Remove the queue configuration file
			modify.File(queueConfigPath).Remove(),

			// Remove the queue service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(moduleImport, queueServiceProvider),
		),
	}
}
