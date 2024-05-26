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
	return DriverSync
}

func (r *Sync) Push(job queue.Job, args []queue.Arg, _ string) error {
	return Call(job.Signature(), args)
}

func (r *Sync) Bulk(jobs []queue.Jobs, _ string) error {
	for _, job := range jobs {
		if job.Delay > 0 {
			time.Sleep(time.Duration(job.Delay) * time.Second)
		}
		if err := Call(job.Job.Signature(), job.Args); err != nil {
			return err
		}
	}

	return nil
}

func (r *Sync) Later(delay uint, job queue.Job, args []queue.Arg, _ string) error {
	time.Sleep(time.Duration(delay) * time.Second)
	return Call(job.Signature(), args)
}

func (r *Sync) Pop(_ string) (queue.Job, []queue.Arg, error) {
	return nil, nil, nil
}
