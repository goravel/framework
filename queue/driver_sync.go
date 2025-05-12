package queue

import (
	"time"

	"github.com/goravel/framework/contracts/queue"
)

var (
	Name              = "sync"
	_    queue.Driver = &Sync{}
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

func (r *Sync) Name() string {
	return Name
}

func (r *Sync) Pop(_ string) (queue.Task, error) {
	// sync driver does not support pop
	return queue.Task{}, nil
}

func (r *Sync) Push(task queue.Task, _ string) error {
	if err := push(task.Jobs); err != nil {
		return err
	}

	if len(task.Chain) > 0 {
		for _, chain := range task.Chain {
			if err := push(chain); err != nil {
				return err
			}
		}
	}

	return nil
}

func push(job queue.Jobs) error {
	if !job.Delay.IsZero() {
		time.Sleep(time.Until(job.Delay))
	}

	var realArgs []any
	for _, arg := range job.Args {
		realArgs = append(realArgs, arg.Value)
	}

	if err := job.Job.Handle(realArgs...); err != nil {
		return err
	}

	return nil
}
