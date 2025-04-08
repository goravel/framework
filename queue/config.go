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

func (r *Config) Config() contractsconfig.Config {
	return r.config
}

func (r *Config) Debug() bool {
	return r.config.GetBool("app.debug")
}

func (r *Config) Default() (connection, queue string, concurrent int) {
	connection = r.config.GetString("queue.default")
	queue = r.config.GetString(fmt.Sprintf("queue.connections.%s.queue", connection), "default")
	concurrent = r.config.GetInt(fmt.Sprintf("queue.connections.%s.concurrent", connection), 1)

	if concurrent < 1 {
		concurrent = 1
	}

	return
}

func (r *Config) Driver(connection string) string {
	return r.config.GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
}

func (r *Config) FailedJobsQuery() db.Query {
	connection := r.config.GetString("queue.failed.database")
	table := r.config.GetString("queue.failed.table")

	return r.db.Connection(connection).Table(table)
}

func (r *Config) QueueKey(connection, queue string) string {
	appName := r.config.GetString("app.name")
	if appName == "" {
		appName = "goravel"
	}

	return fmt.Sprintf("%s_queues:%s_%s", appName, connection, queue)
}

func (r *Config) Via(connection string) any {
	return r.config.Get(fmt.Sprintf("queue.connections.%s.via", connection))
}
