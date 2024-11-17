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
	supportdocker "github.com/goravel/framework/support/docker"
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

	driver := config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))
	database := config.GetString(fmt.Sprintf("database.connections.%s.database", connection))
	username := config.GetString(fmt.Sprintf("database.connections.%s.username", connection))
	password := config.GetString(fmt.Sprintf("database.connections.%s.password", connection))
	containerManager := supportdocker.NewContainerManager()
	databaseDriver, err := containerManager.Create(supportdocker.ContainerType(driver), database, username, password)
	if err != nil {
		return nil, err
	}
	if err = databaseDriver.Ready(); err != nil {
		return nil, err
	}

	return &Database{
		DatabaseDriver: databaseDriver,
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

	if err := r.artisan.Call("migrate"); err != nil {
		return err
	}

	r.orm.Refresh()

	return nil
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
