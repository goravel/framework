package support

import (
	"errors"
	"time"

	"github.com/goravel/framework/contracts/queue"

	"github.com/RichardKnop/machinery/v2"
	"github.com/RichardKnop/machinery/v2/tasks"
)

type Task struct {
	Job        queue.Job
	Jobs       []queue.Jobs
	Chain      bool
	Args       []queue.Arg
	connection string
	queue      string
	server     *machinery.Server
	delay      *time.Time
}

func (receiver *Task) Dispatch() error {
	driver := getDriver(receiver.connection)

	if driver == "" {
		return errors.New("unknown queue driver")
	}
	if driver == DriverSync || driver == "" {
		return receiver.DispatchSync()
	}

	server, err := GetServer(receiver.connection, receiver.queue)
	if err != nil {
		return err
	}
	receiver.server = server

	if receiver.Chain {
		for _, job := range receiver.Jobs {
			if err := receiver.handleAsync(job.Job, job.Args); err != nil {
				return err
			}
		}

		return nil
	} else {
		return receiver.handleAsync(receiver.Job, receiver.Args)
	}
}

func (receiver *Task) DispatchSync() error {
	if receiver.Chain {
		for _, job := range receiver.Jobs {
			if err := receiver.handleSync(job.Job, job.Args); err != nil {
				return err
			}
		}

		return nil
	} else {
		return receiver.handleSync(receiver.Job, receiver.Args)
	}
}

func (receiver *Task) handleSync(job queue.Job, args []queue.Arg) error {
	var realArgs []any
	for _, arg := range args {
		realArgs = append(realArgs, arg.Value)
	}

	return job.Handle(realArgs...)
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

func (receiver *Task) OnConnection(connection string) queue.Task {
	receiver.connection = connection

	return receiver
}

func (receiver *Task) OnQueue(queue string) queue.Task {
	receiver.queue = GetQueueName(receiver.connection, queue)

	return receiver
}

func (receiver *Task) Delay(delay time.Time) queue.Task {
	receiver.delay = &delay

	return receiver
}
