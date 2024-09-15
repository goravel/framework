package database

import (
	"context"
	"fmt"

	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/database/console"
	"github.com/goravel/framework/database/migration"
)

const BindingOrm = "goravel.orm"
const BindingSchema = "goravel.schema"
const BindingSeeder = "goravel.seeder"

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(BindingOrm, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		defaultConnection := config.GetString("database.default")

		orm, err := InitializeOrm(context.Background(), config, defaultConnection)
		if err != nil {
			return nil, fmt.Errorf("[Orm] Init %s connection error: %v", defaultConnection, err)
		}

		return orm, nil
	})
	app.Singleton(BindingSchema, func(app foundation.Application) (any, error) {
		orm := app.MakeOrm()
		config := app.MakeConfig()
		log := app.MakeLog()

		return migration.NewSchema(config, "", log, orm)
	})
	app.Singleton(BindingSeeder, func(app foundation.Application) (any, error) {
		return NewSeederFacade(), nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {
	database.registerCommands(app)
}

func (database *ServiceProvider) registerCommands(app foundation.Application) {
	config := app.MakeConfig()
	seeder := app.MakeSeeder()
	artisan := app.MakeArtisan()
	app.MakeArtisan().Register([]consolecontract.Command{
		console.NewMigrateMakeCommand(config),
		console.NewMigrateCommand(config),
		console.NewMigrateRollbackCommand(config),
		console.NewMigrateResetCommand(config),
		console.NewMigrateRefreshCommand(config, artisan),
		console.NewMigrateFreshCommand(config, artisan),
		console.NewMigrateStatusCommand(config),
		console.NewModelMakeCommand(),
		console.NewObserverMakeCommand(),
		console.NewSeedCommand(config, seeder),
		console.NewSeederMakeCommand(),
		console.NewFactoryMakeCommand(),
	})
}
