package queue

import (
	"fmt"

	configcontract "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
)

type Config struct {
	config configcontract.Config
}

func NewConfig(config configcontract.Config) *Config {
	return &Config{
		config: config,
	}
}

func (r *Config) DefaultConnection() string {
	return r.config.GetString("queue.default")
}

func (r *Config) Queue(connection, queue string) string {
	appName := r.config.GetString("app.name")
	if appName == "" {
		appName = "goravel"
	}
	if connection == "" {
		connection = r.DefaultConnection()
	}
	if queue == "" {
		queue = r.config.GetString(fmt.Sprintf("queue.connections.%s.queue", connection), "default")
	}

	return fmt.Sprintf("%s_queues:%s", appName, queue)
}

func (r *Config) Driver(connection string) string {
	if connection == "" {
		connection = r.DefaultConnection()
	}

	return r.config.GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
}

func (r *Config) Via(connection string) any {
	if connection == "" {
		connection = r.DefaultConnection()
	}

	return r.config.Get(fmt.Sprintf("queue.connections.%s.via", connection))
}

func (r *Config) FailedJobsQuery() orm.Query {
	connection := r.config.GetString("queue.failed.database")
	table := r.config.GetString("queue.failed.table")
	return OrmFacade.Connection(connection).Query().Table(table)
}
