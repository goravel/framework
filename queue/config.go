package queue

import (
	"fmt"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/db"
)

type Config struct {
	config contractsconfig.Config
	db     db.DB
}

func NewConfig(config contractsconfig.Config, db db.DB) *Config {
	return &Config{
		config: config,
		db:     db,
	}
}

func (r *Config) Debug() bool {
	return r.config.GetBool("app.debug")
}

func (r *Config) DefaultConnection() string {
	return r.config.GetString("queue.default")
}

func (r *Config) Driver(connection string) string {
	if connection == "" {
		connection = r.DefaultConnection()
	}

	return r.config.GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
}

func (r *Config) FailedJobsQuery() db.Query {
	connection := r.config.GetString("queue.failed.database")
	table := r.config.GetString("queue.failed.table")

	return r.db.Connection(connection).Table(table)
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

func (r *Config) Size(connection string) int {
	if connection == "" {
		connection = r.DefaultConnection()
	}

	return r.config.GetInt(fmt.Sprintf("queue.connections.%s.size", connection), 100)
}

func (r *Config) Via(connection string) any {
	if connection == "" {
		connection = r.DefaultConnection()
	}

	return r.config.Get(fmt.Sprintf("queue.connections.%s.via", connection))
}
