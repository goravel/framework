package docker

import (
	"fmt"

	contractsconfig "github.com/goravel/framework/contracts/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
)

type Database struct {
	testing.DatabaseDriver
	artisan    contractsconsole.Artisan
	config     contractsconfig.Config
	connection string
	orm        contractsorm.Orm
}

func NewDatabase(app foundation.Application, connection string) (*Database, error) {
	config := app.MakeConfig()
	if config == nil {
		return nil, errors.ConfigFacadeNotSet
	}

	if connection == "" {
		connection = config.GetString("database.default")
	}

	artisanFacade := app.MakeArtisan()
	if artisanFacade == nil {
		return nil, errors.ArtisanFacadeNotSet
	}

	databaseDriverCallback, exist := config.Get(fmt.Sprintf("database.connections.%s.via", connection)).(func() (contractsorm.Driver, error))
	if !exist {
		return nil, errors.OrmDatabaseConfigNotFound
	}
	databaseDriver, err := databaseDriverCallback()
	if err != nil {
		return nil, err
	}

	databaseDocker, err := databaseDriver.Docker()
	if err != nil {
		return nil, err
	}
	if err = databaseDocker.Ready(); err != nil {
		return nil, err
	}

	return &Database{
		DatabaseDriver: databaseDocker,
		artisan:        artisanFacade,
		config:         config,
		connection:     connection,
		orm:            app.MakeOrm(),
	}, nil
}

func (r *Database) Build() error {
	if err := r.DatabaseDriver.Build(); err != nil {
		return err
	}

	r.config.Add(fmt.Sprintf("database.connections.%s.port", r.connection), r.DatabaseDriver.Config().Port)

	r.orm.Refresh()

	return nil
}

func (r *Database) Migrate() error {
	return r.artisan.Call("migrate")
}

func (r *Database) Seed(seeders ...seeder.Seeder) error {
	command := "db:seed"
	if len(seeders) > 0 {
		command += " --seeder"
		for _, seed := range seeders {
			command += fmt.Sprintf(" %s", seed.Signature())
		}
	}

	return r.artisan.Call(command)
}
