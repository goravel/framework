package queue

import (
	"fmt"

	contractsconfig "github.com/goravel/framework/contracts/config"
)

type Config struct {
	contractsconfig.Config

	appName           string
	debug             bool
	defaultConnection string
	defaultQueue      string
	defaultConcurrent int
	failedDatabase    string
	failedTable       string
}

func NewConfig(config contractsconfig.Config) *Config {
	defaultConnection := config.GetString("queue.default")
	defaultQueue := config.GetString(fmt.Sprintf("queue.connections.%s.queue", defaultConnection), "default")
	defaultConcurrent := config.GetInt(fmt.Sprintf("queue.connections.%s.concurrent", defaultConnection), 1)

	if defaultConcurrent < 1 {
		defaultConcurrent = 1
	}

	c := &Config{
		Config: config,

		appName:           config.GetString("app.name", "goravel"),
		debug:             config.GetBool("app.debug"),
		defaultConnection: defaultConnection,
		defaultQueue:      defaultQueue,
		defaultConcurrent: defaultConcurrent,
		failedDatabase:    config.GetString("queue.failed.database"),
		failedTable:       config.GetString("queue.failed.table"),
	}

	return c
}

func (r *Config) Debug() bool {
	return r.debug
}

func (r *Config) DefaultConnection() string {
	return r.defaultConnection
}

func (r *Config) DefaultQueue() string {
	return r.defaultQueue
}

func (r *Config) DefaultConcurrent() int {
	return r.defaultConcurrent
}

func (r *Config) Driver(connection string) string {
	return r.GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
}

func (r *Config) FailedDatabase() string {
	return r.failedDatabase
}

func (r *Config) FailedTable() string {
	return r.failedTable
}

func (r *Config) QueueKey(connection, queue string) string {
	return fmt.Sprintf("%s_queues:%s_%s", r.appName, connection, queue)
}

func (r *Config) Via(connection string) any {
	return r.Get(fmt.Sprintf("queue.connections.%s.via", connection))
}
