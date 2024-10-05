package database

import (
	"context"

	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/database/console"
	"github.com/goravel/framework/database/migration"
	"github.com/goravel/framework/errors"
)

const BindingOrm = "goravel.orm"
const BindingSchema = "goravel.schema"
const BindingSeeder = "goravel.seeder"

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(BindingOrm, func(app foundation.Application) (any, error) {
		ctx := context.Background()
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ErrConfigFacadeNotSet.SetModule(errors.ModuleOrm)
		}

		log := app.MakeLog()
		if log == nil {
			return nil, errors.ErrLogFacadeNotSet.SetModule(errors.ModuleOrm)
		}

		connection := config.GetString("database.default")
		return BuildOrm(ctx, config, connection, log, app.Refresh)
	})
	app.Singleton(BindingSchema, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ErrConfigFacadeNotSet.SetModule(errors.ModuleSchema)
		}

		log := app.MakeLog()
		if log == nil {
			return nil, errors.ErrLogFacadeNotSet.SetModule(errors.ModuleSchema)
		}

		orm := app.MakeOrm()
		if orm == nil {
			return nil, errors.ErrOrmFacadeNotSet.SetModule(errors.ModuleSchema)
		}

		connection := config.GetString("database.default")

		return migration.NewSchema(config, connection, log, orm), nil
	})
	app.Singleton(BindingSeeder, func(app foundation.Application) (any, error) {
		return NewSeederFacade(), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
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
