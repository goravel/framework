package queue

import (
	"fmt"

	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryconfig "github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	machinerylog "github.com/RichardKnop/machinery/v2/log"

	"github.com/goravel/framework/contracts/config"
	contractslog "github.com/goravel/framework/contracts/log"
)

type Machinery struct {
	config     config.Config
	log        contractslog.Log
	connection string
	queue      string
}

func NewMachinery(config config.Config, log contractslog.Log, connection, queue string) *Machinery {
	return &Machinery{
		config:     config,
		log:        log,
		connection: connection,
		queue:      queue,
	}
}

func (r *Machinery) Run() *machinery.Server {
	redisConnection := r.config.GetString(fmt.Sprintf("queue.connections.%s.connection", r.connection))
	redisHost := r.config.GetString(fmt.Sprintf("database.redis.%s.host", redisConnection))
	redisPassword := r.config.GetString(fmt.Sprintf("database.redis.%s.password", redisConnection))
	redisPort := r.config.GetInt(fmt.Sprintf("database.redis.%s.port", redisConnection))
	redisDatabase := r.config.GetInt(fmt.Sprintf("database.redis.%s.database", redisConnection))

	appName := r.config.GetString("app.name")
	if appName == "" {
		appName = "goravel"
	}

	redisQueue := fmt.Sprintf("%s_%s:%s", appName, "queues", r.queue)

	var redisDSN string
	if redisPassword == "" {
		redisDSN = fmt.Sprintf("%s:%d", redisHost, redisPort)
	} else {
		redisDSN = fmt.Sprintf("%s@%s:%d", redisPassword, redisHost, redisPort)
	}

	machineryConfig := &machineryconfig.Config{
		DefaultQueue: redisQueue,
		Redis:        &machineryconfig.RedisConfig{},
	}

	broker := redisbroker.NewGR(machineryConfig, []string{redisDSN}, redisDatabase)
	backend := redisbackend.NewGR(machineryConfig, []string{redisDSN}, redisDatabase)
	lock := eager.New()

	debug := r.config.GetBool("app.debug")
	machinerylog.DEBUG = NewDebug(debug, r.log)
	machinerylog.INFO = NewInfo(debug, r.log)
	machinerylog.WARNING = NewWarning(debug, r.log)
	machinerylog.ERROR = NewError(debug, r.log)
	machinerylog.FATAL = NewFatal(debug, r.log)

	a := machinery.NewServer(machineryConfig, broker, backend, lock)
	a.SendTask()
}

func (r *Machinery) IsQueueExists() bool {
	return r.server.IsQueueExists(r.queue)
}
