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
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	supportstubs "github.com/goravel/framework/support/stubs"
)

func main() {
	stubs := Stubs{}
	modulePath := packages.GetModulePath()
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	databaseConfigPath := path.Config("database.go")
	dbFacadePath := path.Facades("db.go")
	ormFacadePath := path.Facades("orm.go")
	schemaFacadePath := path.Facades("schema.go")
	seederFacadePath := path.Facades("seeder.go")
	databaseServiceProvider := "&database.ServiceProvider{}"
	env := `
DB_HOST=
DB_PORT=
DB_DATABASE=
DB_USERNAME=
DB_PASSWORD=
`

	databaseConfigContent, err := file.GetContent(databaseConfigPath)
	if err != nil {
		// If the file does not exist, use the default content
		databaseConfigContent = supportstubs.DatabaseConfig(moduleName)
	}

	installConfigActionsFunc := func() []contractsmodify.Action {
		var actions []contractsmodify.Action

		for _, config := range stubs.Config() {
			// Skip if the configuration already exists
			if strings.Contains(databaseConfigContent, fmt.Sprintf(`%q`, config.Key)) {
				continue
			}
			actions = append(actions, modify.AddConfig(config.Key, config.Value, config.Annotations...))
		}

		return actions
	}

	uninstallConfigActionsFunc := func() []contractsmodify.Action {
		var actions []contractsmodify.Action

		for _, config := range stubs.Config() {
			// Skip if the configuration does not exist
			if !strings.Contains(databaseConfigContent, fmt.Sprintf(`%q`, config.Key)) {
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

			// Add the database service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, databaseServiceProvider),

			// Register the DB, Orm, Schema and Seeder facades
			modify.WhenFacade(facades.DB, modify.File(dbFacadePath).Overwrite(stubs.DBFacade())),
			modify.WhenFacade(facades.Orm, modify.File(ormFacadePath).Overwrite(stubs.OrmFacade())),
			modify.WhenFacade(facades.Schema,
				// Register the Schema facade
				modify.File(schemaFacadePath).Overwrite(stubs.SchemaFacade()),
			),
			modify.WhenFacade(facades.Seeder,
				// Register the Seeder facade
				modify.File(seederFacadePath).Overwrite(stubs.SeederFacade()),
			),

			// Add configurations to the .env and .env.example files
			modify.WhenFileNotContains(path.Base(".env"), "DB_HOST", modify.File(path.Base(".env")).Append(env)),
			modify.WhenFileNotContains(path.Base(".env.example"), "DB_HOST", modify.File(path.Base(".env.example")).Append(env)),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.DB, facades.Orm, facades.Schema, facades.Seeder},
				// Remove the database service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, databaseServiceProvider),

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
				// Remove the seeder facade file.
				modify.File(seederFacadePath).Remove(),
			),
			modify.WhenFacade(facades.Schema,
				// Remove the schema facade file.
				modify.File(schemaFacadePath).Remove(),
			),
			modify.WhenFacade(facades.Orm, modify.File(ormFacadePath).Remove()),
			modify.WhenFacade(facades.DB, modify.File(dbFacadePath).Remove()),
		).
		Execute()
}
