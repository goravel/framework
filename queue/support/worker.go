package support

import (
	"github.com/goravel/framework/facades"
)

const DriverSync string = "sync"
const DriverRedis string = "redis"

type Worker struct {
	// Specify connection
	Connection string
	// Specify queue
	Queue string
	// Concurrent num
	Concurrent int
}

func (receiver *Worker) Run() error {
	server, err := GetServer(receiver.Connection, receiver.Queue)
	if err != nil {
		return err
	}

	if server == nil {
		return nil
	}

	jobTasks, err := jobs2Tasks(facades.Queue.GetJobs())
	if err != nil {
		return err
	}

	eventTasks, err := eventsToTasks(facades.Event.GetEvents())
	if err != nil {
		return err
	}

	if err := server.RegisterTasks(jobTasks); err != nil {
		return err
	}

	if err := server.RegisterTasks(eventTasks); err != nil {
		return err
	}

	if receiver.Queue == "" {
		receiver.Queue = server.GetConfig().DefaultQueue
	}
	if receiver.Concurrent == 0 {
		receiver.Concurrent = 1
	}
	worker := server.NewWorker(receiver.Queue, receiver.Concurrent)
	if err := worker.Launch(); err != nil {
		return err
	}

	return nil
}
