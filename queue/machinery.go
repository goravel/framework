package queue

import (
	"fmt"

	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/gookit/color"
)

type Machinery struct {
	config *Config
}

func NewMachinery(config *Config) *Machinery {
	return &Machinery{config: config}
}

func (m *Machinery) Server(connection string, queue string) (*machinery.Server, error) {
	driver := m.config.Driver(connection)

	switch driver {
	case DriverSync:
		color.Yellowln("Queue sync driver doesn't need to be run")

		return nil, nil
	case DriverRedis:
		return m.redisServer(connection, queue), nil
	}

	return nil, fmt.Errorf("unknown queue driver: %s", driver)
}

func (m *Machinery) redisServer(connection string, queue string) *machinery.Server {
	redisConfig, database, defaultQueue := m.config.Redis(connection)
	if queue == "" {
		queue = defaultQueue
	}

	cnf := &config.Config{
		DefaultQueue: queue,
		Redis:        &config.RedisConfig{},
	}

	broker := redisbroker.NewGR(cnf, []string{redisConfig}, database)
	backend := redisbackend.NewGR(cnf, []string{redisConfig}, database)
	lock := eager.New()

	return machinery.NewServer(cnf, broker, backend, lock)
}
