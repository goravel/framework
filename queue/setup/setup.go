package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	queueFacade := "Queue"
	databaseDriver := "database"
	packageName := setup.Paths().Main().Package()
	queueFacadePath := path.Facades("queue.go")
	queueConfigPath := path.Config("queue.go")
	queueServiceProvider := "&queue.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	migrationPath := support.Config.Paths.Migrations
	migrationPkg := filepath.Base(migrationPath)
	migrationPkgPath := fmt.Sprintf("%s/%s", packageName, migrationPath)
	jobMigrationFileName, jobMigrationStruct, jobMigrationContent := stubs.JobMigration(migrationPkg, packageName)
	jobMigrationFilePath := path.Base(migrationPath, jobMigrationFileName)
	jobMigrationStructWithPkg := fmt.Sprintf("&%s.%s", migrationPkg, jobMigrationStruct)

	setup.Install(
		modify.WhenFacade(queueFacade,
			// Add the queue service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(moduleImport, queueServiceProvider),

			// Add the queue configuration file
			modify.File(queueConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), packageName)),

			// Add the queue facade to the facades file
			modify.File(queueFacadePath).Overwrite(stubs.QueueFacade(setup.Paths().Facades().Package())),

			// Add the job migration file
			modify.File(jobMigrationFilePath).Overwrite(jobMigrationContent),

			// Register the job migration
			modify.AddMigrationApply(migrationPkgPath, jobMigrationStructWithPkg),
		),

		// Add the database driver
		modify.WhenDriver(databaseDriver, modify.GoFile(queueConfigPath).Find(match.Config("queue")).Modify(modify.AddConfig("default", `"database"`))),
	).Uninstall(
		// Unregister the job migration
		modify.RemoveMigrationApply(migrationPkgPath, jobMigrationStructWithPkg),

		// Remove the job migration file
		modify.File(jobMigrationFilePath).Remove(),

		// Remove the queue facade
		modify.File(queueFacadePath).Remove(),

		// Remove the queue configuration file
		modify.File(queueConfigPath).Remove(),

		// Remove the queue service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, queueServiceProvider),
	).Execute()
}
