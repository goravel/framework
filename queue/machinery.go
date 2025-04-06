// Will be removed in v1.17
package queue

import (
	"fmt"

	"github.com/RichardKnop/machinery/v2"
	backendsiface "github.com/RichardKnop/machinery/v2/backends/iface"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	brokersiface "github.com/RichardKnop/machinery/v2/brokers/iface"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryconfig "github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	machinerylog "github.com/RichardKnop/machinery/v2/log"

	"github.com/goravel/framework/contracts/config"
	contractslog "github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

type Machinery struct {
	backend    backendsiface.Backend
	broker     brokersiface.Broker
	concurrent int
	config     *machineryconfig.Config
	jobs       []queue.Job
	log        contractslog.Log
	queue      string
}

func NewMachinery(config config.Config, log contractslog.Log, jobs []queue.Job, connection, queue string, concurrent int) *Machinery {
	redisConnection := config.GetString(fmt.Sprintf("queue.connections.%s.connection", connection))
	redisHost := config.GetString(fmt.Sprintf("database.redis.%s.host", redisConnection))
	redisPassword := config.GetString(fmt.Sprintf("database.redis.%s.password", redisConnection))
	redisPort := config.GetInt(fmt.Sprintf("database.redis.%s.port", redisConnection))
	redisDatabase := config.GetInt(fmt.Sprintf("database.redis.%s.database", redisConnection))

	appName := config.GetString("app.name")
	if appName == "" {
		appName = "goravel"
	}

	redisQueue := fmt.Sprintf("%s_%s:%s", appName, "queues", queue)

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

	debug := config.GetBool("app.debug")
	machinerylog.DEBUG = NewDebug(debug, log)
	machinerylog.INFO = NewInfo(debug, log)
	machinerylog.WARNING = NewWarning(debug, log)
	machinerylog.ERROR = NewError(debug, log)
	machinerylog.FATAL = NewFatal(debug, log)

	return &Machinery{
		backend:    backend,
		broker:     broker,
		concurrent: concurrent,
		config:     machineryConfig,
		jobs:       jobs,
		log:        log,
		queue:      queue,
	}
}

func (r *Machinery) Run() (*machinery.Worker, error) {
	server := machinery.NewServer(r.config, r.broker, r.backend, eager.New())

	jobTasks, err := jobs2Tasks(r.jobs)
	if err != nil {
		return nil, err
	}
	if len(jobTasks) == 0 {
		return nil, nil
	}

	if err := server.RegisterTasks(jobTasks); err != nil {
		return nil, err
	}

	worker := server.NewWorker(r.queue, r.concurrent)

	go func() {
		if err := worker.Launch(); err != nil {
			r.log.Errorf("Failed to launch machinery worker: %v", err)
		}
	}()

	return worker, nil
}

func (r *Machinery) ExistTasks() bool {
	delayedTasks, err := r.broker.GetDelayedTasks()
	if err != nil {
		return false
	}

	pendingTasks, err := r.broker.GetPendingTasks(r.config.DefaultQueue)
	if err != nil {
		return false
	}

	return len(delayedTasks) > 0 || len(pendingTasks) > 0
}

func jobs2Tasks(jobs []queue.Job) (map[string]any, error) {
	tasks := make(map[string]any)

	for _, job := range jobs {
		if job.Signature() == "" {
			return nil, errors.QueueEmptyJobSignature
		}

		if tasks[job.Signature()] != nil {
			return nil, errors.QueueDuplicateJobSignature.Args(job.Signature())
		}

		tasks[job.Signature()] = job.Handle
	}

	return tasks, nil
}
