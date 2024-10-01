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

var appFacade foundation.Application

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(BindingOrm, func(app foundation.Application) (any, error) {
		ctx := context.Background()
		config := app.MakeConfig()
		connection := config.GetString("database.default")
		orm, err := BuildOrm(ctx, config, connection)
		if err != nil {
			return nil, fmt.Errorf("[Orm] Init %s connection error: %v", connection, err)
		}

		return orm, nil
	})
	app.Singleton(BindingSchema, func(app foundation.Application) (any, error) {
		orm := app.MakeOrm()
		config := app.MakeConfig()
		log := app.MakeLog()

		connection := config.GetString("database.default")
		prefix := config.GetString(fmt.Sprintf("database.connections.%s.prefix", connection))
		schema := config.GetString(fmt.Sprintf("database.connections.%s.schema", connection))
		blueprint := migration.NewBlueprint(prefix, schema)

		return migration.NewSchema(blueprint, config, connection, log, orm), nil
	})
	app.Singleton(BindingSeeder, func(app foundation.Application) (any, error) {
		return NewSeederFacade(), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	appFacade = app
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	if artisanFacade := app.MakeArtisan(); artisanFacade != nil {
		config := app.MakeConfig()
		seeder := app.MakeSeeder()
		artisanFacade.Register([]consolecontract.Command{
			console.NewMigrateMakeCommand(config),
			console.NewMigrateCommand(config),
			console.NewMigrateRollbackCommand(config),
			console.NewMigrateResetCommand(config),
			console.NewMigrateRefreshCommand(config, artisanFacade),
			console.NewMigrateFreshCommand(config, artisanFacade),
			console.NewMigrateStatusCommand(config),
			console.NewModelMakeCommand(),
			console.NewObserverMakeCommand(),
			console.NewSeedCommand(config, seeder),
			console.NewSeederMakeCommand(),
			console.NewFactoryMakeCommand(),
		})
	}
}
