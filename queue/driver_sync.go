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

func (r *Sync) Push(job queue.Job, args []queue.Arg, _ string) error {
	var realArgs []any
	for _, arg := range args {
		realArgs = append(realArgs, arg.Value)
	}

	return job.Handle(realArgs...)
}

func (r *Sync) Bulk(jobs []queue.Jobs, _ string) error {
	for _, job := range jobs {
		if job.Delay != nil {
			time.Sleep(time.Until(*job.Delay))
		}
		var realArgs []any
		for _, arg := range job.Args {
			realArgs = append(realArgs, arg.Value)
		}
		if err := job.Job.Handle(realArgs...); err != nil {
			return err
		}
	}

	return nil
}

func (r *Sync) Later(delay time.Time, job queue.Job, args []queue.Arg, _ string) error {
	var realArgs []any
	for _, arg := range args {
		realArgs = append(realArgs, arg.Value)
	}
	time.Sleep(time.Until(delay))

	return job.Handle(realArgs...)
}

func (r *Sync) Pop(_ string) (queue.Job, []queue.Arg, error) {
	// sync driver does not support pop
	return nil, nil, nil
}
