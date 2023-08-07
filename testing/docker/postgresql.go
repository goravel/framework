package docker

import (
	"fmt"

	"github.com/ory/dockertest/v3"
	"github.com/spf13/cast"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/testing"
)

type Postgresql struct {
	config     contractsconfig.Config
	connection string
}

func NewPostgresql(config contractsconfig.Config, connection string) *Postgresql {
	return &Postgresql{
		config:     config,
		connection: connection,
	}
}

func (receiver *Postgresql) Config(resource *dockertest.Resource) testing.Config {
	return testing.Config{
		Host:     "127.0.0.1",
		Port:     cast.ToInt(resource.GetPort("5432/tcp")),
		Database: receiver.config.GetString(fmt.Sprintf("database.connections.%s.database", receiver.connection)),
		Username: receiver.config.GetString(fmt.Sprintf("database.connections.%s.username", receiver.connection)),
		Password: receiver.config.GetString(fmt.Sprintf("database.connections.%s.password", receiver.connection)),
	}
}

func (receiver *Postgresql) Clear(pool *dockertest.Pool, resource *dockertest.Resource) error {
	return pool.Purge(resource)
}

func (receiver *Postgresql) Name() orm.Driver {
	return orm.DriverPostgresql
}

func (receiver *Postgresql) Image() *dockertest.RunOptions {
	database := receiver.config.GetString(fmt.Sprintf("database.connections.%s.database", receiver.connection))
	username := receiver.config.GetString(fmt.Sprintf("database.connections.%s.username", receiver.connection))
	password := receiver.config.GetString(fmt.Sprintf("database.connections.%s.password", receiver.connection))

	return &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_USER=" + username,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + database,
			"listen_addresses = '*'",
		},
	}
}
