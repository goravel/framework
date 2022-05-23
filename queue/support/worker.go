package support

import (
	"github.com/goravel/framework/support/facades"
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
	if receiver.Connection == "" {
		receiver.Connection = facades.Config.GetString("queue.default")
	}

	server, err := getServer(receiver.Connection, receiver.Queue)
	if err != nil {
		return err
	}

	if server == nil {
		return nil
	}

	if err := server.RegisterTasks(jobs2Tasks(facades.Queue.GetJobs())); err != nil {
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
