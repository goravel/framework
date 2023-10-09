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
	if appName == "" {
		appName = "goravel"
	}
	if connection == "" {
		connection = r.DefaultConnection()
	}
	if queue == "" {
		queue = r.config.GetString(fmt.Sprintf("queue.connections.%s.queue", connection), "default")
	}

	return fmt.Sprintf("%s_%s:%s", appName, "queues", queue)
}

func (r *Config) Driver(connection string) string {
	if connection == "" {
		connection = r.config.GetString("queue.default")
	}

	return r.config.GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
}

func (r *Config) Redis(queueConnection string) *redis.Client {
	connection := r.config.GetString(fmt.Sprintf("queue.connections.%s.connection", queueConnection))
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

func (r *Config) Database(queueConnection string) orm.Orm {
	return OrmFacade.Connection(queueConnection)
}
