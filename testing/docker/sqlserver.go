package docker

import (
	"fmt"

	"github.com/ory/dockertest/v3"
	"github.com/spf13/cast"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/testing"
)

type Sqlserver struct {
	config     contractsconfig.Config
	connection string
}

func NewSqlserver(config contractsconfig.Config, connection string) *Sqlserver {
	return &Sqlserver{
		config:     config,
		connection: connection,
	}
}

func (receiver *Sqlserver) Config(resource *dockertest.Resource) testing.Config {
	return testing.Config{
		Host:     "127.0.0.1",
		Port:     cast.ToInt(resource.GetPort("1433/tcp")),
		Database: "msdb",
		Username: "sa",
		Password: receiver.config.GetString(fmt.Sprintf("database.connections.%s.password", receiver.connection)),
	}
}

func (receiver *Sqlserver) Clear(pool *dockertest.Pool, resource *dockertest.Resource) error {
	return pool.Purge(resource)
}

func (receiver *Sqlserver) Name() orm.Driver {
	return orm.DriverSqlserver
}

func (receiver *Sqlserver) Image() *dockertest.RunOptions {
	password := receiver.config.GetString(fmt.Sprintf("database.connections.%s.password", receiver.connection))

	return &dockertest.RunOptions{
		Repository: "mcr.microsoft.com/mssql/server",
		Tag:        "latest",
		Env: []string{
			"MSSQL_SA_PASSWORD=" + password,
			"ACCEPT_EULA=Y",
		},
	}
}
