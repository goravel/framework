package docker

import (
	"context"
	"fmt"

	"github.com/ory/dockertest/v3"

	contractsconfig "github.com/goravel/framework/contracts/config"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/database"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/support/docker"
)

type Database struct {
	app            foundation.Application
	config         contractsconfig.Config
	connection     string
	driver         testing.DatabaseDriver
	gormInitialize gorm.Initialize
	image          *testing.Image
	pool           *dockertest.Pool
	resource       *dockertest.Resource
}

func NewDatabase(app foundation.Application, connection string, gormInitialize gorm.Initialize) (*Database, error) {
	config := app.MakeConfig()

	if connection == "" {
		connection = config.GetString("database.default")
	}
	driver := config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))

	var databaseDriver testing.DatabaseDriver
	switch contractsorm.Driver(driver) {
	case contractsorm.DriverMysql:
		databaseDriver = NewMysql(config, connection)
	case contractsorm.DriverPostgresql:
		databaseDriver = NewPostgresql(config, connection)
	case contractsorm.DriverSqlserver:
		databaseDriver = NewSqlserver(config, connection)
	case contractsorm.DriverSqlite:
		databaseDriver = NewSqlite(config, connection)
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
	pool, err := docker.Pool()
	if err != nil {
		return err
	}
	receiver.pool = pool

	var opts *dockertest.RunOptions
	if receiver.image != nil {
		opts = &dockertest.RunOptions{
			Repository: receiver.image.Repository,
			Tag:        receiver.image.Tag,
			Env:        receiver.image.Env,
		}
	} else {
		opts = receiver.driver.Image()
	}
	resource, err := docker.Resource(pool, opts)
	if err != nil {
		return err
	}
	receiver.resource = resource

	if receiver.image != nil && receiver.image.Timeout > 0 {
		_ = resource.Expire(receiver.image.Timeout)
	} else {
		_ = resource.Expire(3600)
	}

	dbConfig := receiver.driver.Config(resource)
	receiver.config.Add(fmt.Sprintf("database.connections.%s.host", receiver.connection), dbConfig.Host)
	receiver.config.Add(fmt.Sprintf("database.connections.%s.port", receiver.connection), dbConfig.Port)
	receiver.config.Add(fmt.Sprintf("database.connections.%s.database", receiver.connection), dbConfig.Database)
	receiver.config.Add(fmt.Sprintf("database.connections.%s.username", receiver.connection), dbConfig.Username)
	receiver.config.Add(fmt.Sprintf("database.connections.%s.password", receiver.connection), dbConfig.Password)

	if err := pool.Retry(func() error {
		_, err := receiver.gormInitialize.InitializeQuery(context.Background(), receiver.config, receiver.driver.Name().String())
		if err != nil {
			return err
		}

		receiver.app.MakeArtisan().Call("migrate")

		return nil
	}); err != nil {
		return err
	}

	receiver.app.Singleton(database.BindingOrm, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		defaultConnection := config.GetString("database.default")

		orm, err := database.InitializeOrm(context.Background(), config, defaultConnection)
		if err != nil {
			return nil, fmt.Errorf("[Orm] Init %s connection error: %v", defaultConnection, err)
		}

		return orm, nil
	})

	return nil
}

func (receiver *Database) Config() testing.Config {
	return receiver.driver.Config(receiver.resource)
}

func (receiver *Database) Clear() error {
	return receiver.driver.Clear(receiver.pool, receiver.resource)
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
