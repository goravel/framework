package queue

import (
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type Task struct {
	config     contractsqueue.Config
	connection string
	chain      bool
	delay      time.Time
	jobs       []contractsqueue.Jobs
	queueKey   string
}

func NewTask(config contractsqueue.Config, job contractsqueue.Job, args ...[]contractsqueue.Arg) *Task {
	var arg []contractsqueue.Arg
	if len(args) > 0 {
		arg = args[0]
	}

	connection, queue, _ := config.Default()

	return &Task{
		config:     config,
		connection: connection,
		jobs: []contractsqueue.Jobs{
			{
				Job:  job,
				Args: arg,
			},
		},
		queueKey: config.QueueKey(connection, queue),
	}
}

func NewChainTask(config contractsqueue.Config, jobs []contractsqueue.Jobs) *Task {
	connection, queue, _ := config.Default()

	return &Task{
		config:     config,
		connection: connection,
		chain:      true,
		jobs:       jobs,
		queueKey:   config.QueueKey(connection, queue),
	}
}

// Delay sets a delay time for the task
func (r *Task) Delay(delay time.Time) contractsqueue.Task {
	r.delay = delay
	return r
}

// Dispatch dispatches the task
func (r *Task) Dispatch() error {
	driver, err := NewDriver(r.connection, r.config)
	if err != nil {
		return err
	}

	if r.chain {
		return driver.Bulk(r.jobs, r.queueKey)
	} else {
		job := r.jobs[0]
		if !r.delay.IsZero() {
			return driver.Later(r.delay, job.Job, job.Args, r.queueKey)
		}
		return driver.Push(job.Job, job.Args, r.queueKey)
	}
}

// DispatchSync dispatches the task synchronously
func (r *Task) DispatchSync() error {
	if r.chain {
		for _, job := range r.jobs {
			var realArgs []any
			for _, arg := range job.Args {
				realArgs = append(realArgs, arg.Value)
			}

			if err := job.Job.Handle(realArgs...); err != nil {
				return err
			}
		}

		return nil
	} else {
		job := r.jobs[0]
		var realArgs []any
		for _, arg := range job.Args {
			realArgs = append(realArgs, arg.Value)
		}

		return job.Job.Handle(realArgs...)
	}
}

// OnConnection sets the connection name
func (r *Task) OnConnection(connection string) contractsqueue.Task {
	r.connection = connection
	return r
}

// OnQueue sets the queue name
func (r *Task) OnQueue(queue string) contractsqueue.Task {
	r.queueKey = r.config.QueueKey(r.connection, queue)
	return r
}
