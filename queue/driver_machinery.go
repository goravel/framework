// TODO: Will be removed in v1.17

package queue

import (
	"time"

	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/log"

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
	//TODO implement me
	panic("implement me")
}

func (m *Machinery) Push(job queue.Job, args []any, queue string) error {
	//TODO implement me
	panic("implement me")
}

func (m *Machinery) Bulk(jobs []queue.Jobs, queue string) error {
	//TODO implement me
	panic("implement me")
}

func (m *Machinery) Later(delay time.Time, job queue.Job, args []any, queue string) error {
	//TODO implement me
	panic("implement me")
}

func (m *Machinery) Pop(queue string) (queue.Job, []any, error) {
	//TODO implement me
	panic("implement me")
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
