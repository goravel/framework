package database

import (
	"context"

	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/database/console"
	consolemigration "github.com/goravel/framework/database/console/migration"
	"github.com/goravel/framework/database/migration"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/database/schema"
	"github.com/goravel/framework/database/seeder"
	"github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(orm.BindingOrm, func(app foundation.Application) (any, error) {
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
		orm, err := orm.BuildOrm(ctx, config, connection, log, app.Refresh)
		if err != nil {
			return nil, errors.OrmInitConnection.Args(connection, err).SetModule(errors.ModuleOrm)
		}

		return orm, nil
	})
	app.Singleton(schema.BindingSchema, func(app foundation.Application) (any, error) {
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

		return schema.NewSchema(config, log, orm, nil), nil
	})
	app.Singleton(seeder.BindingSeeder, func(app foundation.Application) (any, error) {
		return seeder.NewSeederFacade(), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	artisan := app.MakeArtisan()
	config := app.MakeConfig()
	log := app.MakeLog()
	schema := app.MakeSchema()
	seeder := app.MakeSeeder()

	if artisan != nil && config != nil && log != nil && schema != nil && seeder != nil {
		var migrator contractsmigration.Migrator

		driver := config.GetString("database.migrations.driver")
		if driver == contractsmigration.MigratorDefault {
			migrator = migration.NewDefaultMigrator(artisan, schema, config.GetString("database.migrations.table"))
		} else if driver == contractsmigration.MigratorSql {
			var err error
			migrator, err = migration.NewSqlMigrator(config)
			if err != nil {
				log.Error(errors.MigrationSqlMigratorInit.Args(err).SetModule(errors.ModuleMigration))
				return
			}
		} else {
			log.Error(errors.MigrationUnsupportedDriver.Args(driver).SetModule(errors.ModuleMigration))
			return
		}

		artisan.Register([]contractsconsole.Command{
			consolemigration.NewMigrateMakeCommand(migrator),
			consolemigration.NewMigrateCommand(migrator),
			consolemigration.NewMigrateRollbackCommand(migrator),
			consolemigration.NewMigrateResetCommand(migrator),
			consolemigration.NewMigrateRefreshCommand(artisan),
			consolemigration.NewMigrateFreshCommand(artisan, migrator),
			consolemigration.NewMigrateStatusCommand(migrator),
			console.NewModelMakeCommand(),
			console.NewObserverMakeCommand(),
			console.NewSeedCommand(config, seeder),
			console.NewSeederMakeCommand(),
			console.NewFactoryMakeCommand(),
			console.NewWipeCommand(config, schema),
		})
	}
}
