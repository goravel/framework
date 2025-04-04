package queue

import (
	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryconfig "github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	machinerylog "github.com/RichardKnop/machinery/v2/log"

	"github.com/goravel/framework/contracts/config"
	contractslog "github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
)

type Machinery struct {
	config queue.Config
	log    contractslog.Log
}

func NewMachinery(config config.Config, log contractslog.Log) *machinery.Server {
	redisConfig, database, defaultQueue := config.Redis(connection)
	if queue == "" {
		queue = defaultQueue
	}

	cnf := &machineryconfig.Config{
		DefaultQueue: queue,
		Redis:        &machineryconfig.RedisConfig{},
	}

	broker := redisbroker.NewGR(cnf, []string{redisConfig}, database)
	backend := redisbackend.NewGR(cnf, []string{redisConfig}, database)
	lock := eager.New()

	debug := config.GetBool("app.debug")
	machinerylog.DEBUG = NewDebug(debug, log)
	machinerylog.INFO = NewInfo(debug, log)
	machinerylog.WARNING = NewWarning(debug, log)
	machinerylog.ERROR = NewError(debug, log)
	machinerylog.FATAL = NewFatal(debug, log)

	return machinery.NewServer(cnf, broker, backend, lock)
}
