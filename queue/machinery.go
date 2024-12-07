package queue

import (
	logcontract "github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/machinery"
	redisbackend "github.com/goravel/machinery/backends/redis"
	redisbroker "github.com/goravel/machinery/brokers/redis"
	"github.com/goravel/machinery/config"
	"github.com/goravel/machinery/locks/eager"
)

type Machinery struct {
	config *Config
	log    logcontract.Log
}

func NewMachinery(config *Config, log logcontract.Log) *Machinery {
	return &Machinery{config: config, log: log}
}

func (m *Machinery) Server(connection string, queue string) (*machinery.Server, error) {
	driver := m.config.Driver(connection)

	switch driver {
	case DriverSync:
		color.Warningln("Queue sync driver doesn't need to be run")

		return nil, nil
	case DriverRedis:
		return m.redisServer(connection, queue), nil
	}

	return nil, errors.QueueDriverNotSupported.Args(driver)
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

	broker := redisbroker.New(cnf, []string{redisConfig}, database)
	backend := redisbackend.New(cnf, []string{redisConfig}, database)
	lock := eager.New()

	/*debug := m.config.config.GetBool("app.debug")
	log.DEBUG = NewDebug(debug, m.log)
	log.INFO = NewInfo(debug, m.log)
	log.WARNING = NewWarning(debug, m.log)
	log.ERROR = NewError(debug, m.log)
	log.FATAL = NewFatal(debug, m.log)*/

	return machinery.NewServer(cnf, broker, backend, lock)
}
