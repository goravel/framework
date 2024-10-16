package database

import (
	"context"

	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/database/console"
	consolemigration "github.com/goravel/framework/database/console/migration"
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
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleOrm)
		}

		log := app.MakeLog()
		if log == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleOrm)
		}

		connection := config.GetString("database.default")
		orm, err := BuildOrm(ctx, config, connection, log, app.Refresh)
		if err != nil {
			return nil, errors.OrmInitConnection.Args(connection, err).SetModule(errors.ModuleOrm)
		}

		return orm, nil
	})
	app.Singleton(BindingSchema, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleSchema)
		}

		log := app.MakeLog()
		if log == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleSchema)
		}

		orm := app.MakeOrm()
		if orm == nil {
			return nil, errors.OrmFacadeNotSet.SetModule(errors.ModuleSchema)
		}

		return migration.NewSchema(config, log, orm, nil), nil
	})
	app.Singleton(BindingSeeder, func(app foundation.Application) (any, error) {
		return NewSeederFacade(), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	artisan := app.MakeArtisan()
	config := app.MakeConfig()
	schema := app.MakeSchema()
	seeder := app.MakeSeeder()

	if artisan != nil && config != nil && schema != nil && seeder != nil {
		artisan.Register([]contractsconsole.Command{
			consolemigration.NewMigrateMakeCommand(config, schema),
			consolemigration.NewMigrateCommand(config, schema),
			consolemigration.NewMigrateRollbackCommand(config),
			consolemigration.NewMigrateResetCommand(config),
			consolemigration.NewMigrateRefreshCommand(config, artisan),
			consolemigration.NewMigrateFreshCommand(config, artisan),
			consolemigration.NewMigrateStatusCommand(config),
			console.NewModelMakeCommand(),
			console.NewObserverMakeCommand(),
			console.NewSeedCommand(config, seeder),
			console.NewSeederMakeCommand(),
			console.NewFactoryMakeCommand(),
		})
	}
}
