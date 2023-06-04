package queue

import (
	"github.com/goravel/framework/contracts/queue"
)

const DriverSync string = "sync"
const DriverRedis string = "redis"

type Worker struct {
	concurrent int
	connection string
	machinery  *Machinery
	jobs       []queue.Job
	queue      string
}

func NewWorker(config *Config, concurrent int, connection string, jobs []queue.Job, queue string) *Worker {
	return &Worker{
		concurrent: concurrent,
		connection: connection,
		machinery:  NewMachinery(config),
		jobs:       jobs,
		queue:      queue,
	}
}

func (receiver *Worker) Run() error {
	server, err := receiver.machinery.Server(receiver.connection, receiver.queue)
	if err != nil {
		return err
	}
	if server == nil {
		return nil
	}

	jobTasks, err := jobs2Tasks(receiver.jobs)
	if err != nil {
		return err
	}

	if err := server.RegisterTasks(jobTasks); err != nil {
		return err
	}

	if receiver.queue == "" {
		receiver.queue = server.GetConfig().DefaultQueue
	}
	if receiver.concurrent == 0 {
		receiver.concurrent = 1
	}
	worker := server.NewWorker(receiver.queue, receiver.concurrent)
	if err := worker.Launch(); err != nil {
		return err
	}

	return nil
}
