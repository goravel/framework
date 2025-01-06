package queue

import (
	"time"

	"github.com/goravel/framework/contracts/queue"
)

type Task struct {
	config     queue.Config
	connection string
	chain      bool
	delay      time.Time
	jobs       []queue.Jobs
	queue      string
}

func NewTask(config queue.Config, job queue.Job, args []any) *Task {
	return &Task{
		config:     config,
		connection: config.DefaultConnection(),
		jobs: []queue.Jobs{
			{
				Job:  job,
				Args: args,
			},
		},
		queue: config.Queue(config.DefaultConnection(), ""),
	}
}

func NewChainTask(config queue.Config, jobs []queue.Jobs) *Task {
	return &Task{
		config:     config,
		connection: config.DefaultConnection(),
		chain:      true,
		jobs:       jobs,
		queue:      config.Queue(config.DefaultConnection(), ""),
	}
}

// Delay sets a delay time for the task
func (r *Task) Delay(delay time.Time) queue.Task {
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
func (r *Task) OnConnection(connection string) queue.Task {
	r.connection = connection
	return r
}

// OnQueue sets the queue name
func (r *Task) OnQueue(queue string) queue.Task {
	r.queue = r.config.Queue(r.connection, queue)
	return r
}
