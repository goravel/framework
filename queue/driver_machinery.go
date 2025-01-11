// TODO: Will be removed in v1.17

package queue

import (
	"reflect"
	"time"

	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"

	contractslog "github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
)

type Machinery struct {
	connection string
	config     queue.Config
	log        contractslog.Log
}

func NewMachinery(connection string, config queue.Config, log contractslog.Log) *Machinery {
	return &Machinery{
		connection: connection,
		config:     config,
		log:        log,
	}
}

func (m *Machinery) Connection() string {
	return m.connection
}

func (m *Machinery) Driver() string {
	return queue.DriverMachinery
}

func (m *Machinery) Push(job queue.Job, args []any, queue string) error {
	server := m.server(queue)
	_, err := server.SendTask(&tasks.Signature{
		Name: job.Signature(),
		Args: m.argsToMachineryArgs(args),
	})
	return err
}

func (m *Machinery) Bulk(jobs []queue.Jobs, queue string) error {
	var signatures []*tasks.Signature
	for _, job := range jobs {
		signatures = append(signatures, &tasks.Signature{
			Name: job.Job.Signature(),
			Args: m.argsToMachineryArgs(job.Args),
			ETA:  &job.Delay,
		})
	}

	chain, err := tasks.NewChain(signatures...)
	if err != nil {
		return err
	}

	server := m.server(queue)
	_, err = server.SendChain(chain)

	return err
}

func (m *Machinery) Later(delay time.Time, job queue.Job, args []any, queue string) error {
	server := m.server(queue)
	_, err := server.SendTask(&tasks.Signature{
		Name: job.Signature(),
		Args: m.argsToMachineryArgs(args),
		ETA:  &delay,
	})
	return err
}

func (m *Machinery) Pop(queue string) (queue.Job, []any, error) {
	return nil, nil, nil
}

func (m *Machinery) Run(jobs []queue.Job, queue string, concurrent int) error {
	server := m.server(queue)
	if server == nil {
		return nil
	}

	jobTasks, err := jobs2Tasks(jobs)
	if err != nil {
		return err
	}

	if err = server.RegisterTasks(jobTasks); err != nil {
		return err
	}

	if queue == "" {
		queue = server.GetConfig().DefaultQueue
	}

	worker := server.NewWorker(queue, concurrent)
	return worker.Launch()
}

func (m *Machinery) server(queue string) *machinery.Server {
	redisConfig, database, defaultQueue := m.config.Redis(m.connection)
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

	debug := m.config.Debug()
	log.DEBUG = NewDebug(debug, m.log)
	log.INFO = NewInfo(debug, m.log)
	log.WARNING = NewWarning(debug, m.log)
	log.ERROR = NewError(debug, m.log)
	log.FATAL = NewFatal(debug, m.log)

	return machinery.NewServer(cnf, broker, backend, lock)
}

func (m *Machinery) argsToMachineryArgs(args []any) []tasks.Arg {
	var realArgs []tasks.Arg
	for _, arg := range args {
		reflected := reflect.ValueOf(arg)
		realArgs = append(realArgs, tasks.Arg{
			Type:  reflected.Type().String(),
			Value: reflected.Interface(),
		})
	}
	return realArgs
}
