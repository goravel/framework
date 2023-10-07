package queue

import (
	"errors"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type Task struct {
	config     *Config
	connection string
	chain      bool
	delay      *carbon.Carbon
	driver     queue.Driver
	jobs       []queue.Jobs
	queue      string
}

func NewTask(config *Config, job queue.Job, args []queue.Arg) *Task {
	return &Task{
		config:     config,
		connection: config.DefaultConnection(),
		driver:     NewDriver(config.DefaultConnection(), config),
		jobs: []queue.Jobs{
			{
				Job:  job,
				Args: args,
			},
		},
	}
}

func NewChainTask(config *Config, jobs []queue.Jobs) *Task {
	return &Task{
		config:     config,
		connection: config.DefaultConnection(),
		chain:      true,
		driver:     NewDriver(config.DefaultConnection(), config),
		jobs:       jobs,
	}
}

func (receiver *Task) Delay(delay carbon.Carbon) queue.Task {
	receiver.delay = &delay

	return receiver
}

func (receiver *Task) Dispatch() error {
	driver := receiver.config.Driver(receiver.connection)
	if driver == "" {
		return errors.New("unknown queue driver")
	}
	if driver == DriverSync || driver == "" {
		return receiver.DispatchSync()
	}

	server, err := receiver.machinery.Server(receiver.connection, receiver.queue)
	if err != nil {
		return err
	}

	receiver.server = server

	if receiver.chain {
		for _, job := range receiver.jobs {
			if err := receiver.handleAsync(job.Job, job.Args); err != nil {
				return err
			}
		}

		return nil
	} else {
		job := receiver.jobs[0]

		return receiver.handleAsync(job.Job, job.Args)
	}
}

func (receiver *Task) DispatchSync() error {
	if receiver.chain {
		for _, job := range receiver.jobs {
			if err := receiver.handleSync(job.Job, job.Args); err != nil {
				return err
			}
		}

		return nil
	} else {
		job := receiver.jobs[0]

		return receiver.handleSync(job.Job, job.Args)
	}
}

func (receiver *Task) OnConnection(connection string) queue.Task {
	receiver.connection = connection

	return receiver
}

func (receiver *Task) OnQueue(queue string) queue.Task {
	receiver.queue = receiver.config.Queue(receiver.connection, queue)

	return receiver
}

func (receiver *Task) handleAsync(job queue.Job, args []queue.Arg) error {
	/*_, err := receiver.server.SendTask(&tasks.Signature{
		Name: job.Signature(),
		Args: realArgs,
		ETA:  receiver.delay,
	})*/
	if err != nil {
		return err
	}

	return nil
}

func (receiver *Task) handleSync(job queue.Job, args []queue.Arg) error {
	return job.Handle(realArgs...)
}
