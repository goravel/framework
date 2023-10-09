package queue

import (
	"github.com/goravel/framework/contracts/queue"
)

type Worker struct {
	concurrent int
	connection string
	driver     queue.Driver
	queue      string
}

func NewWorker(config *Config, concurrent int, connection string, queue string) *Worker {
	return &Worker{
		concurrent: concurrent,
		connection: connection,
		driver:     NewDriver(connection, config),
		queue:      queue,
	}
}

func (receiver *Worker) Run() error {
	receiver.driver.Server(receiver.concurrent, receiver.queue)

	return nil
}
