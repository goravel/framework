package db

import (
	"fmt"

	"github.com/google/wire"

	"github.com/goravel/framework/contracts/config"
	databasecontract "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
)

var ConfigSet = wire.NewSet(NewConfigImpl, wire.Bind(new(Config), new(*ConfigImpl)))
var _ Config = &ConfigImpl{}

type Config interface {
	Reads() []databasecontract.Config
	Writes() []databasecontract.Config
}

type ConfigImpl struct {
	config     config.Config
	connection string
}

func NewConfigImpl(config config.Config, connection string) *ConfigImpl {
	return &ConfigImpl{
		config:     config,
		connection: connection,
	}
}

func (c *ConfigImpl) Reads() []databasecontract.Config {
	configs := c.config.Get(fmt.Sprintf("database.connections.%s.read", c.connection))
	if configs, ok := configs.([]databasecontract.Config); ok {
		return c.fillDefault(configs)
	}

	return []databasecontract.Config{}
}

func (c *ConfigImpl) Writes() []databasecontract.Config {
	configs := c.config.Get(fmt.Sprintf("database.connections.%s.write", c.connection))
	if configs, ok := configs.([]databasecontract.Config); ok {
		return c.fillDefault(configs)
	}

	return c.fillDefault([]databasecontract.Config{{}})
}

func (c *ConfigImpl) fillDefault(configs []databasecontract.Config) []databasecontract.Config {
	var newConfigs []databasecontract.Config
	driver := c.config.GetString(fmt.Sprintf("database.connections.%s.driver", c.connection))
	for _, item := range configs {
		if driver != orm.DriverSqlite.String() {
			if item.Host == "" {
				item.Host = c.config.GetString(fmt.Sprintf("database.connections.%s.host", c.connection))
			}
			if item.Port == 0 {
				item.Port = c.config.GetInt(fmt.Sprintf("database.connections.%s.port", c.connection))
			}
			if item.Username == "" {
				item.Username = c.config.GetString(fmt.Sprintf("database.connections.%s.username", c.connection))
			}
			if item.Password == "" {
				item.Password = c.config.GetString(fmt.Sprintf("database.connections.%s.password", c.connection))
			}
		}
		if item.Database == "" {
			item.Database = c.config.GetString(fmt.Sprintf("database.connections.%s.database", c.connection))
		}
		newConfigs = append(newConfigs, item)
	}

	return newConfigs
}
