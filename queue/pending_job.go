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
	delay      time.Time
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
			Jobs: contractsqueue.Jobs{
				Job:  job,
				Args: arg,
			},
		},
	}
}

func NewPendingChainJob(config contractsqueue.Config, jobs []contractsqueue.Jobs) *PendingJob {
	if len(jobs) == 0 {
		return nil
	}

	var chain []contractsqueue.Jobs
	for _, job := range jobs[1:] {
		chain = append(chain, contractsqueue.Jobs{
			Job:   job.Job,
			Args:  job.Args,
			Delay: job.Delay,
		})
	}

	job := contractsqueue.Jobs{
		Job:   jobs[0].Job,
		Args:  jobs[0].Args,
		Delay: jobs[0].Delay,
	}

	connection, queue, _ := config.Default()

	return &PendingJob{
		config:     config,
		connection: connection,
		queueKey:   config.QueueKey(connection, queue),
		task: contractsqueue.Task{
			Uuid:  uuid.New().String(),
			Jobs:  job,
			Chain: chain,
		},
	}
}

// Delay sets a delay time for the task
func (r *PendingJob) Delay(delay time.Time) contractsqueue.PendingJob {
	r.delay = delay
	return r
}

// Dispatch dispatches the task
func (r *PendingJob) Dispatch() error {
	driver, err := NewDriver(r.connection, r.config)
	if err != nil {
		return err
	}

	if !r.delay.IsZero() {
		if !r.task.Delay.IsZero() {
			r.task.Delay = r.task.Delay.Add(carbon.Now().DiffAbsInDuration(carbon.FromStdTime(r.delay)))
		} else {
			r.task.Delay = r.delay
		}
	}

	return driver.Push(r.task, r.queueKey)
}

// DispatchSync dispatches the task synchronously
func (r *PendingJob) DispatchSync() error {
	syncDriver := NewSync(r.connection)

	return syncDriver.Push(r.task, r.queueKey)
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
