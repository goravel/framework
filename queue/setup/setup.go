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
	stubs := Stubs{}
	databaseDriver := "database"
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	queueFacadePath := path.Facades("queue.go")
	queueConfigPath := path.Config("queue.go")
	queueServiceProvider := "&queue.ServiceProvider{}"
	modulePath := packages.GetModulePath()
	migrationPath := support.Config.Paths.Migration
	migrationPkg := filepath.Base(migrationPath)
	migrationPkgPath := fmt.Sprintf("%s/%s", moduleName, migrationPath)
	jobMigrationFileName, jobMigrationStruct, jobMigrationContent := stubs.JobMigration(migrationPkg, moduleName)
	jobMigrationFilePath := path.Base(migrationPath, jobMigrationFileName)
	jobMigrationStructWithPkg := fmt.Sprintf("&%s.%s", migrationPkg, jobMigrationStruct)

	packages.Setup(os.Args).
		Install(
			// Add the queue service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, queueServiceProvider),

			// Add the queue configuration file
			modify.File(queueConfigPath).Overwrite(stubs.Config(moduleName)),

			// Add the queue facade to the facades file
			modify.File(queueFacadePath).Overwrite(stubs.QueueFacade()),

			// Add the job migration file
			modify.File(jobMigrationFilePath).Overwrite(jobMigrationContent),

			// Register the job migration
			modify.AddMigrationApply(migrationPkgPath, jobMigrationStructWithPkg),

			// Add the database driver
			modify.WhenDriver(databaseDriver, modify.GoFile(queueConfigPath).Find(match.Config("queue")).Modify(modify.AddConfig("default", `"database"`))),
		).
		Uninstall(
			// Unregister the job migration
			modify.RemoveMigrationApply(migrationPkgPath, jobMigrationStructWithPkg),

			// Remove the job migration file
			modify.File(jobMigrationFilePath).Remove(),

			// Remove the queue facade
			modify.File(queueFacadePath).Remove(),

			// Remove the queue configuration file
			modify.File(queueConfigPath).Remove(),

			// Remove the queue service provider from the providers array in bootstrap/providers.go
			modify.RemoveProviderApply(modulePath, queueServiceProvider),
		).
		Execute()
}
