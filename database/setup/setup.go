package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/facades"
	contractsmodify "github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	supportstubs "github.com/goravel/framework/support/stubs"
)

func main() {
	stubs := Stubs{}
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

	installConfigActionsFunc := func() []contractsmodify.Action {
		var actions []contractsmodify.Action
		content, err := file.GetContent(databaseConfigPath)
		if err != nil {
			color.Errorln("failed to get database configuration content")
			return actions
		}

		for _, config := range stubs.Config() {
			// Skip if the configuration already exists
			if strings.Contains(content, fmt.Sprintf(`%q`, config.Key)) {
				continue
			}
			actions = append(actions, modify.AddConfig(config.Key, config.Value, config.Annotations...))
		}

		return actions
	}

	uninstallConfigActionsFunc := func() []contractsmodify.Action {
		var actions []contractsmodify.Action
		content, err := file.GetContent(databaseConfigPath)
		if err != nil {
			color.Errorln("failed to get database configuration content")
			return actions
		}

		for _, config := range stubs.Config() {
			// Skip if the configuration does not exist
			if !strings.Contains(content, fmt.Sprintf(`%q`, config.Key)) {
				continue
			}
			actions = append(actions, modify.RemoveConfig(config.Key))
		}

		return actions
	}

	packages.Setup(os.Args).
		Install(
			// Create config/database.go
			modify.WhenFileNotExists(databaseConfigPath, modify.File(databaseConfigPath).Overwrite(supportstubs.DatabaseConfig(moduleName))),

			// Add database configuration to config/database.go
			modify.GoFile(databaseConfigPath).Find(match.Config("database")).Modify(installConfigActionsFunc()...),

			// Add the database service provider to the providers array in config/app.go
			modify.GoFile(appConfigPath).
				Find(match.Imports()).Modify(modify.AddImport(modulePath)).
				Find(match.Providers()).Modify(modify.Register(databaseServiceProvider)),

			// Register the DB, Orm, Schema and Seeder facades
			modify.WhenFacade(facades.DB, modify.File(dbFacadePath).Overwrite(stubs.DBFacade())),
			modify.WhenFacade(facades.Orm, modify.File(ormFacadePath).Overwrite(stubs.OrmFacade())),
			modify.WhenFacade(facades.Schema,
				// Register the Schema facade
				modify.File(schemaFacadePath).Overwrite(stubs.SchemaFacade()),

				// Create the console kernel file if it does not exist.
				modify.WhenFileNotExists(kernelPath, modify.File(kernelPath).Overwrite(stubs.Kernel())),

				// Modify app/providers/app_service_provider.go to register migrations
				modify.GoFile(appServiceProviderPath).
					Find(match.Imports()).Modify(modify.AddImport(databaseImport)).
					Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
					Find(match.RegisterFunc()).Modify(modify.Add(registerMigration)),
			),
			modify.WhenFacade(facades.Seeder,
				// Register the Seeder facade
				modify.File(seederFacadePath).Overwrite(stubs.SeederFacade()),

				// Create the console kernel file if it does not exist.
				modify.WhenFileNotExists(kernelPath, modify.File(kernelPath).Overwrite(stubs.Kernel())),

				// Modify app/providers/app_service_provider.go to register seeders
				modify.GoFile(appServiceProviderPath).
					Find(match.Imports()).Modify(modify.AddImport(databaseImport)).
					Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
					Find(match.RegisterFunc()).Modify(modify.Add(registerSeeder)),
			),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.DB, facades.Orm, facades.Schema, facades.Seeder},
				modify.File(kernelPath).Remove(),

				// Remove the database service provider from the providers array in config/app.go
				modify.GoFile(appConfigPath).
					Find(match.Providers()).Modify(modify.Unregister(databaseServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(modulePath)),

				// Remove database configuration from config/database.go
				modify.GoFile(databaseConfigPath).Find(match.Config("database")).Modify(uninstallConfigActionsFunc()...),

				// Remove config/database.go
				modify.When(func(_ map[string]any) bool {
					content, err := file.GetContent(databaseConfigPath)
					if err != nil {
						return false
					}
					return content == supportstubs.DatabaseConfig(moduleName)
				}, modify.File(databaseConfigPath).Remove()),
			),

			// Remove the DB, Orm, Schema and Seeder facades
			modify.WhenFacade(facades.Seeder,
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
			modify.WhenFacade(facades.Schema,
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
			modify.WhenFacade(facades.Orm, modify.File(ormFacadePath).Remove()),
			modify.WhenFacade(facades.DB, modify.File(dbFacadePath).Remove()),
		).
		Execute()
}

func isKernelNotModified(_ map[string]any) bool {
	content, err := file.GetContent(path.Database("kernel.go"))
	if err != nil {
		return false
	}

	return content == Stubs{}.Kernel()
}
