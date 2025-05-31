// Will be removed in v1.17
package queue

import (
	"fmt"

	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryconfig "github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	machinerylog "github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"

	"github.com/goravel/framework/contracts/config"
	contractslog "github.com/goravel/framework/contracts/log"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

type Machinery struct {
	appName       string
	log           contractslog.Log
	queueToServer map[string]*machinery.Server
	redisDatabase int
	redisDSN      string
}

func NewMachinery(config config.Config, log contractslog.Log, connection string) *Machinery {
	redisConnection := config.GetString(fmt.Sprintf("queue.connections.%s.connection", connection))
	redisHost := config.GetString(fmt.Sprintf("database.redis.%s.host", redisConnection))
	redisPassword := config.GetString(fmt.Sprintf("database.redis.%s.password", redisConnection))
	redisPort := config.GetInt(fmt.Sprintf("database.redis.%s.port", redisConnection))
	redisDatabase := config.GetInt(fmt.Sprintf("database.redis.%s.database", redisConnection))

	appName := config.GetString("app.name")
	if appName == "" {
		appName = "goravel"
	}

	var redisDSN string
	if redisPassword == "" {
		redisDSN = fmt.Sprintf("%s:%d", redisHost, redisPort)
	} else {
		redisDSN = fmt.Sprintf("%s@%s:%d", redisPassword, redisHost, redisPort)
	}

	debug := config.GetBool("app.debug")
	machinerylog.DEBUG = NewDebug(debug, log)
	machinerylog.INFO = NewInfo(debug, log)
	machinerylog.WARNING = NewWarning(debug, log)
	machinerylog.ERROR = NewError(debug, log)
	machinerylog.FATAL = NewFatal(debug, log)

	return &Machinery{
		appName:       appName,
		log:           log,
		queueToServer: make(map[string]*machinery.Server),
		redisDatabase: redisDatabase,
		redisDSN:      redisDSN,
	}
}

func (r *Machinery) Driver() string {
	return contractsqueue.DriverMachinery
}

// Machinery server will pop tasks automatically
func (r *Machinery) Pop(queue string) (contractsqueue.ReservedJob, error) {
	return nil, nil
}

func (r *Machinery) Push(task contractsqueue.Task, queue string) error {
	if len(task.Chain) > 0 {
		return r.pushChain(task, queue)
	}

	var realArgs []tasks.Arg
	for _, arg := range task.Args {
		realArgs = append(realArgs, tasks.Arg{
			Type:  arg.Type,
			Value: arg.Value,
		})
	}

	_, err := r.Server(queue).SendTask(&tasks.Signature{
		Name: task.Job.Signature(),
		Args: realArgs,
		ETA:  &task.Delay,
	})

	return err
}

func (r *Machinery) Run(jobs []contractsqueue.Job, queue string, concurrent int) (*machinery.Worker, error) {
	jobTasks, err := jobs2Tasks(jobs)
	if err != nil {
		return nil, err
	}
	if len(jobTasks) == 0 {
		return nil, nil
	}

	server := r.Server(queue)
	if err := server.RegisterTasks(jobTasks); err != nil {
		return nil, err
	}

	worker := server.NewWorker(r.queueKey(queue), concurrent)

	go func() {
		if err := worker.Launch(); err != nil {
			r.log.Errorf("Failed to launch machinery worker: %v", err)
		}
	}()

	return worker, nil
}

func (r *Machinery) Server(queue string) *machinery.Server {
	if server, ok := r.queueToServer[queue]; ok {
		return server
	}

	machineryConfig := &machineryconfig.Config{
		DefaultQueue: r.queueKey(queue),
		Redis:        &machineryconfig.RedisConfig{},
	}

	broker := redisbroker.NewGR(machineryConfig, []string{r.redisDSN}, r.redisDatabase)
	backend := redisbackend.NewGR(machineryConfig, []string{r.redisDSN}, r.redisDatabase)

	server := machinery.NewServer(machineryConfig, broker, backend, eager.New())
	r.queueToServer[queue] = server

	return server
}

func (r *Machinery) pushChain(task contractsqueue.Task, queue string) error {
	server := r.Server(queue)
	chainJobs := append([]contractsqueue.ChainJob{task.ChainJob}, task.Chain...)

	var signatures []*tasks.Signature
	for _, chainJob := range chainJobs {
		var realArgs []tasks.Arg
		for _, arg := range chainJob.Args {
			realArgs = append(realArgs, tasks.Arg{
				Type:  arg.Type,
				Value: arg.Value,
			})
		}

		signatures = append(signatures, &tasks.Signature{
			Name: chainJob.Job.Signature(),
			Args: realArgs,
			ETA:  &chainJob.Delay,
		})
	}

	chain, err := tasks.NewChain(signatures...)
	if err != nil {
		return err
	}

	_, err = server.SendChain(chain)

	return err
}

func (r *Machinery) queueKey(queue string) string {
	return fmt.Sprintf("%s_%s:%s", r.appName, "queues", queue)
}

func jobs2Tasks(jobs []contractsqueue.Job) (map[string]any, error) {
	tasks := make(map[string]any)

	for _, job := range jobs {
		signature := job.Signature()
		if signature == "" {
			return nil, errors.QueueEmptyJobSignature
		}

		if tasks[signature] != nil {
			return nil, errors.QueueDuplicateJobSignature.Args(signature)
		}

		tasks[signature] = job.Handle
	}

	return tasks, nil
}
