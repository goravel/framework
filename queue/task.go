package queue

import (
	"errors"
	"time"

	"github.com/RichardKnop/machinery/v2"
	"github.com/RichardKnop/machinery/v2/tasks"

	"github.com/goravel/framework/contracts/queue"
)

type Task struct {
	config     *Config
	connection string
	chain      bool
	delay      *time.Time
	machinery  *Machinery
	jobs       []queue.Jobs
	queue      string
	server     *machinery.Server
}

func NewTask(config *Config, job queue.Job, args []queue.Arg) *Task {
	return &Task{
		config:     config,
		connection: config.DefaultConnection(),
		machinery:  NewMachinery(config),
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
		machinery:  NewMachinery(config),
		jobs:       jobs,
	}
}

func (receiver *Task) Delay(delay time.Time) queue.Task {
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
	var realArgs []tasks.Arg
	for _, arg := range args {
		realArgs = append(realArgs, tasks.Arg{
			Type:  arg.Type,
			Value: arg.Value,
		})
	}

	_, err := receiver.server.SendTask(&tasks.Signature{
		Name: job.Signature(),
		Args: realArgs,
		ETA:  receiver.delay,
	})
	if err != nil {
		return err
	}

	return nil
}

func (receiver *Task) handleSync(job queue.Job, args []queue.Arg) error {
	var realArgs []any
	for _, arg := range args {
		realArgs = append(realArgs, arg.Value)
	}

	return job.Handle(realArgs...)
}
