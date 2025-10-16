package queue

import (
	"github.com/goravel/framework/contracts/config"
	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type QueueRunner struct {
	config config.Config
	worker contractsqueue.Worker
}

func NewQueueRunner(config config.Config, queue contractsqueue.Queue) *QueueRunner {
	var worker contractsqueue.Worker
	if queue != nil {
		worker = queue.Worker()
	}

	return &QueueRunner{
		config: config,
		worker: worker,
	}
}

func (r *QueueRunner) ShouldRun() bool {
	return r.worker != nil && r.config.GetString("queue.default") != ""
}

func (r *QueueRunner) Run() error {
	return r.worker.Run()
}

func (r *QueueRunner) Shutdown() error {
	return r.worker.Shutdown()
}
