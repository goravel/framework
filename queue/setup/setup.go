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
	setup := packages.Setup(os.Args)
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

	setup.Install(
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

		// Add the database driver
		modify.WhenDriver(databaseDriver, modify.GoFile(queueConfigPath).Find(match.Config("queue")).Modify(modify.AddConfig("default", `"database"`))),
	).Uninstall(
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
	).Execute()
}
