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

func (r *Sync) Pop(_ string) (*queue.Task, error) {
	// sync driver does not support pop
	return nil, nil
}

func (r *Sync) Push(task queue.Task, _ string) error {
	if task.Data.Delay != nil && !task.Data.Delay.IsZero() {
		time.Sleep(time.Until(*task.Data.Delay))
	}

	if err := task.Data.Job.Handle(filterArgsType(task.Data.Args)...); err != nil {
		return err
	}

	if len(task.Data.Chained) > 0 {
		for _, chained := range task.Data.Chained {
			if chained.Delay != nil && !chained.Delay.IsZero() {
				time.Sleep(time.Until(*chained.Delay))
			}

			if err := chained.Job.Handle(filterArgsType(chained.Args)...); err != nil {
				return err
			}
		}
	}

	return nil
}
