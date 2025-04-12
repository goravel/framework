package queue

import (
	"time"

	"github.com/goravel/framework/contracts/queue"
	"github.com/spf13/cast"
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

func (r *Sync) Later(delay time.Time, task queue.Task, _ string) error {
	if !delay.IsZero() {
		time.Sleep(time.Until(delay))
	}

	return r.Push(task, "")
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

func filterArgsType(args []queue.Arg) []any {
	realArgs := make([]any, 0, len(args))
	for _, arg := range args {
		switch arg.Type {
		case "string":
			realArgs = append(realArgs, cast.ToString(arg.Value))
		case "int":
			realArgs = append(realArgs, cast.ToInt(arg.Value))
		case "bool":
			realArgs = append(realArgs, cast.ToBool(arg.Value))
		case "[]string":
			realArgs = append(realArgs, cast.ToStringSlice(arg.Value))
		case "[]int":
			realArgs = append(realArgs, cast.ToIntSlice(arg.Value))
		case "[]bool":
			realArgs = append(realArgs, cast.ToBoolSlice(arg.Value))
		}
	}
	return realArgs
}
