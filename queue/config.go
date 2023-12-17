package queue

import (
	"fmt"

	"github.com/redis/go-redis/v9"

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
	if len(appName) == 0 {
		appName = "goravel"
	}
	if len(connection) == 0 {
		connection = r.DefaultConnection()
	}
	if len(queue) == 0 {
		queue = r.config.GetString(fmt.Sprintf("queue.connections.%s.queue", connection), "default")
	}

	return fmt.Sprintf("%s_%s:%s", appName, "queues", queue)
}

func (r *Config) Driver(connection string) string {
	if len(connection) == 0 {
		connection = r.DefaultConnection()
	}

	return r.config.GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
}

func (r *Config) Redis(queueConnection string) *redis.Client {
	connection := r.config.GetString(fmt.Sprintf("queue.connections.%s.database", queueConnection))
	host := r.config.GetString(fmt.Sprintf("database.redis.%s.host", connection))
	password := r.config.GetString(fmt.Sprintf("database.redis.%s.password", connection))
	port := r.config.GetInt(fmt.Sprintf("database.redis.%s.port", connection))
	database := r.config.GetInt(fmt.Sprintf("database.redis.%s.database", connection))

	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       database,
	})
}

func (r *Config) Database(queueConnection string) orm.Query {
	connection := r.config.GetString(fmt.Sprintf("queue.connections.%s.database", queueConnection))
	table := r.config.GetString(fmt.Sprintf("queue.connections.%s.table", queueConnection))
	return OrmFacade.Connection(connection).Query().Table(table)
}

func (r *Config) FailedJobsQuery() orm.Query {
	connection := r.config.GetString("queue.failed.database")
	table := r.config.GetString("queue.failed.table")
	return OrmFacade.Connection(connection).Query().Table(table)
}
