package queue

import (
	"time"

	"github.com/google/uuid"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type PendingJob struct {
	config     contractsqueue.Config
	connection string
	chain      bool
	delay      *time.Time
	jobs       []contractsqueue.Jobs
	queueKey   string
	task       contractsqueue.Task
}

func NewPendingJob(config contractsqueue.Config, job contractsqueue.Job, args ...[]contractsqueue.Arg) *PendingJob {
	var arg []contractsqueue.Arg
	if len(args) > 0 {
		arg = args[0]
	}

	connection, queue, _ := config.Default()

	return &PendingJob{
		config:     config,
		connection: connection,
		queueKey:   config.QueueKey(connection, queue),
		task: contractsqueue.Task{
			Uuid: uuid.New().String(),
			Data: contractsqueue.TaskData{
				Job:  job,
				Args: arg,
			},
		},
	}
}

func NewPendingChainJob(config contractsqueue.Config, jobs []contractsqueue.Jobs) *PendingJob {
	var chained []contractsqueue.TaskData
	for _, job := range jobs[1:] {
		chained = append(chained, contractsqueue.TaskData{
			Job:   job.Job,
			Args:  job.Args,
			Delay: job.Delay,
		})
	}

	connection, queue, _ := config.Default()

	return &PendingJob{
		config:     config,
		connection: connection,
		queueKey:   config.QueueKey(connection, queue),
		task: contractsqueue.Task{
			Uuid: uuid.New().String(),
			Data: contractsqueue.TaskData{
				Job:     jobs[0].Job,
				Args:    jobs[0].Args,
				Delay:   jobs[0].Delay,
				Chained: chained,
			},
		},
	}
}

// Delay sets a delay time for the task
func (r *PendingJob) Delay(delay time.Time) contractsqueue.PendingJob {
	r.delay = &delay
	return r
}

// Dispatch dispatches the task
func (r *PendingJob) Dispatch() error {
	driver, err := NewDriver(r.connection, r.config)
	if err != nil {
		return err
	}

	if r.delay != nil && !r.delay.IsZero() {
		if r.task.Data.Delay != nil && !r.task.Data.Delay.IsZero() {
			*r.task.Data.Delay = r.task.Data.Delay.Add(carbon.Now().DiffAbsInDuration(carbon.FromStdTime(*r.delay)))
		} else {
			r.task.Data.Delay = r.delay
		}
	}

	return driver.Push(r.task, r.queueKey)
}

// DispatchSync dispatches the task synchronously
func (r *PendingJob) DispatchSync() error {
	if err := r.task.Data.Job.Handle(filterArgsType(r.task.Data.Args)...); err != nil {
		return err
	}

	if len(r.task.Data.Chained) > 0 {
		for _, job := range r.task.Data.Chained {
			if err := job.Job.Handle(filterArgsType(job.Args)...); err != nil {
				return err
			}
		}
	}

	return nil
}

// OnConnection sets the connection name
func (r *PendingJob) OnConnection(connection string) contractsqueue.PendingJob {
	r.connection = connection
	return r
}

// OnQueue sets the queue name
func (r *PendingJob) OnQueue(queue string) contractsqueue.PendingJob {
	r.queueKey = r.config.QueueKey(r.connection, queue)
	return r
}
