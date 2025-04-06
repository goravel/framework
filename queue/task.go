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
	queue      string
}

func NewTask(config contractsqueue.Config, job contractsqueue.Job, args ...[]any) *Task {
	var arg []any
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
		queue: config.Queue(connection, queue),
	}
}

func NewChainTask(config contractsqueue.Config, jobs []contractsqueue.Jobs) *Task {
	connection, queue, _ := config.Default()

	return &Task{
		config:     config,
		connection: connection,
		chain:      true,
		jobs:       jobs,
		queue:      config.Queue(connection, queue),
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
		return driver.Bulk(r.jobs, r.queue)
	} else {
		job := r.jobs[0]
		if !r.delay.IsZero() {
			return driver.Later(r.delay, job.Job, job.Args, r.queue)
		}
		return driver.Push(job.Job, job.Args, r.queue)
	}
}

// DispatchSync dispatches the task synchronously
func (r *Task) DispatchSync() error {
	if r.chain {
		for _, job := range r.jobs {
			if err := job.Job.Handle(job.Args...); err != nil {
				return err
			}
		}
		return nil
	} else {
		job := r.jobs[0]
		return job.Job.Handle(job.Args...)
	}
}

// OnConnection sets the connection name
func (r *Task) OnConnection(connection string) contractsqueue.Task {
	r.connection = connection
	return r
}

// OnQueue sets the queue name
func (r *Task) OnQueue(queue string) contractsqueue.Task {
	r.queue = r.config.Queue(r.connection, queue)
	return r
}
