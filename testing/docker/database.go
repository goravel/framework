package docker

import (
	"context"
	"fmt"
	"time"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/gorm"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/testing"
	frameworkdatabase "github.com/goravel/framework/database"
	supportdocker "github.com/goravel/framework/support/docker"
)

type Database struct {
	app            foundation.Application
	config         contractsconfig.Config
	connection     string
	driver         testing.DatabaseDriver
	gormInitialize gorm.Initialize
	image          *testing.Image
}

func NewDatabase(app foundation.Application, connection string, gormInitialize gorm.Initialize) (*Database, error) {
	config := app.MakeConfig()

	if connection == "" {
		connection = config.GetString("database.default")
	}
	driver := config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))
	database := config.GetString(fmt.Sprintf("database.connections.%s.database", connection))
	username := config.GetString(fmt.Sprintf("database.connections.%s.username", connection))
	password := config.GetString(fmt.Sprintf("database.connections.%s.password", connection))

	var databaseDriver testing.DatabaseDriver
	switch contractsorm.Driver(driver) {
	case contractsorm.DriverMysql:
		databaseDriver = supportdocker.NewMysql(database, username, password)
	case contractsorm.DriverPostgresql:
		databaseDriver = supportdocker.NewPostgresql(database, username, password)
	case contractsorm.DriverSqlserver:
		databaseDriver = supportdocker.NewSqlserver(database, username, password)
	case contractsorm.DriverSqlite:
		databaseDriver = supportdocker.NewSqlite(database)
	default:
		return nil, fmt.Errorf("not found database connection: %s", connection)
	}

	return &Database{
		app:            app,
		config:         config,
		connection:     connection,
		driver:         databaseDriver,
		gormInitialize: gormInitialize,
	}, nil
}

func (receiver *Database) Build() error {
	if receiver.image != nil {
		receiver.driver.Image(*receiver.image)
	}

	if err := receiver.driver.Build(); err != nil {
		return err
	}

	config := receiver.driver.Config()
	receiver.config.Add(fmt.Sprintf("database.connections.%s.port", receiver.connection), config.Port)

	var query contractsorm.Query
	for i := 0; i < 60; i++ {
		query1, err := receiver.gormInitialize.InitializeQuery(context.Background(), receiver.config, receiver.driver.Name().String())
		if err == nil {
			query = query1
			break
		}

		time.Sleep(2 * time.Second)
	}
	if query == nil {
		return fmt.Errorf("connect to %s failed", receiver.driver.Name().String())
	}

	receiver.app.MakeArtisan().Call("migrate")
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

func (receiver *Database) Config() testing.DatabaseConfig {
	return receiver.driver.Config()
}

// Deprecated: Use Stop instead.
func (receiver *Database) Clear() error {
	return receiver.Stop()
}

func (receiver *Database) Image(image testing.Image) {
	receiver.image = &image
}

func (receiver *Database) Seed(seeds ...seeder.Seeder) {
	command := "db:seed"
	if len(seeds) > 0 {
		command += " --seeder"
		for _, seed := range seeds {
			command += fmt.Sprintf(" %s", seed.Signature())
		}
	}

	receiver.app.MakeArtisan().Call(command)
}

func (receiver *Database) Stop() error {
	return receiver.driver.Stop()
}
