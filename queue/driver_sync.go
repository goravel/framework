package queue

import (
	"time"

	"github.com/goravel/framework/contracts/queue"
)

type Sync struct {
	connection string
}

func NewSync(connection string) *Sync {
	return &Sync{
		connection: connection,
	}
}

func (r *Sync) Connection() string {
	return r.connection
}

func (r *Sync) Driver() string {
	return queue.DriverSync
}

func (r *Sync) Push(job queue.Job, args []any, _ string) error {
	return job.Handle(args...)
}

func (r *Sync) Bulk(jobs []queue.Jobs, _ string) error {
	for _, job := range jobs {
		if job.Delay > 0 {
			time.Sleep(job.Delay)
		}
		if err := job.Job.Handle(job.Args...); err != nil {
			return err
		}
	}

	return nil
}

func (r *Sync) Later(delay time.Time, job queue.Job, args []any, _ string) error {
	time.Sleep(time.Until(delay))
	return job.Handle(args...)
}

func (r *Sync) Pop(_ string) (queue.Job, []any, error) {
	// sync driver does not support pop
	return nil, nil, nil
}
