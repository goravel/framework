package docker

import (
	"context"
	"errors"
	"fmt"

	contractsconfig "github.com/goravel/framework/contracts/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/testing"
	frameworkdatabase "github.com/goravel/framework/database"
	"github.com/goravel/framework/support/color"
	supportdocker "github.com/goravel/framework/support/docker"
)

type Database struct {
	testing.DatabaseDriver
	app        foundation.Application
	artisan    contractsconsole.Artisan
	config     contractsconfig.Config
	connection string
}

func NewDatabase(app foundation.Application, connection string) *Database {
	config := app.MakeConfig()

	if config == nil {
		return nil, errors.New("config facade is not set")
	}

	if connection == "" {
		connection = config.GetString("database.default")
	}

	driver := config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))
	database := config.GetString(fmt.Sprintf("database.connections.%s.database", connection))
	username := config.GetString(fmt.Sprintf("database.connections.%s.username", connection))
	password := config.GetString(fmt.Sprintf("database.connections.%s.password", connection))
	databaseDriver := supportdocker.DatabaseDriver(supportdocker.ContainerType(driver), database, username, password)

	return &Database{
		app:            app,
		artisan:        app.MakeArtisan(),
		config:         config,
		connection:     connection,
		DatabaseDriver: databaseDriver,
	}
}

func (receiver *Database) Build() error {
	if err := receiver.DatabaseDriver.Build(); err != nil {
		return err
	}

	receiver.config.Add(fmt.Sprintf("database.connections.%s.port", receiver.connection), receiver.DatabaseDriver.Config().Port)
	artisan := receiver.app.MakeArtisan()
	if artisan == nil {
		return errors.New("artisan instance is not available")
	}

	artisan.Call("migrate")

	// TODO Find a better way to refresh the database connection
	receiver.app.Singleton(frameworkdatabase.BindingOrm, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		defaultConnection := config.GetString("database.default")

		orm, err := frameworkdatabase.InitializeOrm(context.Background(), config, defaultConnection)
		if err != nil {
			return nil, fmt.Errorf("[Orm] Init %s connection error: %v", defaultConnection, err)
		}

		return orm, nil
	})

	return nil
}

func (receiver *Database) Seed(seeds ...seeder.Seeder) {
	command := "db:seed"
	if len(seeds) > 0 {
		command += " --seeder"
		for _, seed := range seeds {
			command += fmt.Sprintf(" %s", seed.Signature())
		}
	}

	artisan := receiver.app.MakeArtisan()
	if artisan == nil {
		color.Red().Println("artisan instance is not available")
		return
	}

	artisan.Call(command)
}
