package main

import (
	"fmt"
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	dbFacade := "DB"
	ormFacade := "Orm"
	schemaFacade := "Schema"
	seederFacade := "Seeder"
	modulePath := packages.GetModulePath()
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	appConfigPath := path.Config("app.go")
	databaseConfigPath := path.Config("database.go")
	kernelPath := path.Database("kernel.go")
	dbFacadePath := path.Facades("db.go")
	ormFacadePath := path.Facades("orm.go")
	schemaFacadePath := path.Facades("schema.go")
	seederFacadePath := path.Facades("seeder.go")
	appServiceProviderPath := path.App("providers", "app_service_provider.go")
	databaseServiceProvider := "&database.ServiceProvider{}"
	registerMigration := "facades.Schema().Register(database.Kernel{}.Migrations())"
	registerSeeder := "facades.Seeder().Register(database.Kernel{}.Seeders())"
	databaseImport := fmt.Sprintf("%s/database", moduleName)
	facadesImport := fmt.Sprintf("%s/app/facades", moduleName)

	packages.Setup(os.Args).
		Install(
			// Register the DB, Orm, Schema and Seeder facades
			modify.WhenFacade(dbFacade, modify.File(dbFacadePath).Overwrite(stubs.DBFacade())),
			modify.WhenFacade(ormFacade, modify.File(ormFacadePath).Overwrite(stubs.OrmFacade())),
			modify.WhenFacade(schemaFacade,
				// Register the Schema facade
				modify.File(schemaFacadePath).Overwrite(stubs.SchemaFacade()),

				// Create the console kernel file if it does not exist.
				modify.When(func() bool {
					return !file.Exists(kernelPath)
				}, modify.File(kernelPath).Overwrite(stubs.Kernel())),

				// Modify app/providers/app_service_provider.go to register migrations
				modify.GoFile(appServiceProviderPath).
					Find(match.Imports()).Modify(modify.AddImport(databaseImport)).
					Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
					Find(match.RegisterFunc()).Modify(modify.Add(registerMigration)),
			),
			modify.WhenFacade(seederFacade,
				// Register the Seeder facade
				modify.File(seederFacadePath).Overwrite(stubs.SeederFacade()),

				// Create the console kernel file if it does not exist.
				modify.When(func() bool {
					return !file.Exists(kernelPath)
				}, modify.File(kernelPath).Overwrite(stubs.Kernel())),

				// Modify app/providers/app_service_provider.go to register seeders
				modify.GoFile(appServiceProviderPath).
					Find(match.Imports()).Modify(modify.AddImport(databaseImport)).
					Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
					Find(match.RegisterFunc()).Modify(modify.Add(registerSeeder)),
			),

			// Create config/database.go and database/kernel.go
			modify.File(databaseConfigPath).Overwrite(stubs.Config(moduleName)),

			// Add the database service provider to the providers array in config/app.go
			modify.GoFile(appConfigPath).
				Find(match.Imports()).Modify(modify.AddImport(modulePath)).
				Find(match.Providers()).Modify(modify.Register(databaseServiceProvider)),
		).
		Uninstall(
			modify.WhenNoFacades([]string{dbFacade, ormFacade, schemaFacade, seederFacade},
				modify.File(kernelPath).Remove(),

				// Remove the database service provider from the providers array in config/app.go
				modify.GoFile(appConfigPath).
					Find(match.Providers()).Modify(modify.Unregister(databaseServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(modulePath)),

				// Remove config/database.go
				modify.File(databaseConfigPath).Remove(),
			),

			// Remove the DB, Orm, Schema and Seeder facades
			modify.WhenFacade(seederFacade,
				// Revert modifications in app/providers/app_service_provider.go
				modify.GoFile(appServiceProviderPath).
					Find(match.RegisterFunc()).Modify(modify.Remove(registerSeeder)).
					Find(match.Imports()).Modify(modify.RemoveImport(databaseImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)),

				// Remove the database kernel file if it was not modified.
				modify.When(isKernelNotModified, modify.File(kernelPath).Remove()),

				// Remove the seeder facade file.
				modify.File(seederFacadePath).Remove(),
			),
			modify.WhenFacade(schemaFacade,
				// Revert modifications in app/providers/app_service_provider.go
				modify.GoFile(appServiceProviderPath).
					Find(match.RegisterFunc()).Modify(modify.Remove(registerMigration)).
					Find(match.Imports()).Modify(modify.RemoveImport(databaseImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)),

				// Remove the database kernel file if it was not modified.
				modify.When(isKernelNotModified, modify.File(kernelPath).Remove()),

				// Remove the schema facade file.
				modify.File(schemaFacadePath).Remove(),
			),
			modify.WhenFacade(ormFacade, modify.File(ormFacadePath).Remove()),
			modify.WhenFacade(dbFacade, modify.File(dbFacadePath).Remove()),
		).
		Execute()
}

func isKernelNotModified() bool {
	content, err := file.GetContent(path.Database("kernel.go"))
	if err != nil {
		return false
	}

	return content == Stubs{}.Kernel()
}
