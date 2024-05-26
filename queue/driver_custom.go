package queue

import (
	"time"

	"github.com/goravel/framework/contracts/queue"
)

type Custom struct {
	connection string
}

func NewCustom(connection string) *Custom {
	return &Custom{
		connection: connection,
	}
}

func (r *Custom) Connection() string {
	return r.connection
}

func (r *Custom) Driver() string {
	return DriverCustom
}

func (r *Custom) Push(job queue.Job, args []any, _ string) error {
	return Call(job.Signature(), args)
}

func (r *Custom) Bulk(jobs []queue.Jobs, _ string) error {
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

func (r *Custom) Later(delay uint, job queue.Job, args []any, _ string) error {
	time.Sleep(time.Duration(delay) * time.Second)
	return Call(job.Signature(), args)
}

func (r *Custom) Pop(_ string) (queue.Job, []any, error) {
	return nil, nil, nil
}
