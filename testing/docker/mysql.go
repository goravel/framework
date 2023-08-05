package docker

import (
	"fmt"

	"github.com/ory/dockertest/v3"
	"github.com/spf13/cast"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/testing"
)

type Mysql struct {
	config     contractsconfig.Config
	connection string
}

func NewMysql(config contractsconfig.Config, connection string) *Mysql {
	return &Mysql{
		config:     config,
		connection: connection,
	}
}

func (receiver *Mysql) Config(resource *dockertest.Resource) testing.Config {
	return testing.Config{
		Host:     "127.0.0.1",
		Port:     cast.ToInt(resource.GetPort("3306/tcp")),
		Database: receiver.config.GetString(fmt.Sprintf("database.connections.%s.database", receiver.connection)),
		Username: receiver.config.GetString(fmt.Sprintf("database.connections.%s.username", receiver.connection)),
		Password: receiver.config.GetString(fmt.Sprintf("database.connections.%s.password", receiver.connection)),
	}
}

func (receiver *Mysql) Clear(pool *dockertest.Pool, resource *dockertest.Resource) error {
	return pool.Purge(resource)
}

func (receiver *Mysql) Name() orm.Driver {
	return orm.DriverMysql
}

func (receiver *Mysql) Image() *dockertest.RunOptions {
	database := receiver.config.GetString(fmt.Sprintf("database.connections.%s.database", receiver.connection))
	username := receiver.config.GetString(fmt.Sprintf("database.connections.%s.username", receiver.connection))
	password := receiver.config.GetString(fmt.Sprintf("database.connections.%s.password", receiver.connection))
	env := []string{
		"MYSQL_ROOT_PASSWORD=" + password,
		"MYSQL_DATABASE=" + database,
	}
	if username != "root" {
		env = append(env, "MYSQL_USER="+username)
		env = append(env, "MYSQL_PASSWORD="+password)
	}

	return &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "latest",
		Env:        env,
	}
}
