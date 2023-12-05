package queue

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/convert"
)

func jobs2Tasks(jobs []queue.Job) (map[string]any, error) {
	tasks := make(map[string]any)

	for _, job := range jobs {
		if job.Signature() == "" {
			return nil, errors.New("the Signature of job can't be empty")
		}

		if tasks[job.Signature()] != nil {
			return nil, fmt.Errorf("job signature duplicate: %s, the names of Job and Listener cannot be duplicated", job.Signature())
		}

		tasks[job.Signature()] = job.Handle
	}

	return tasks, nil
}

func eventsToTasks(events map[event.Event][]event.Listener) (map[string]any, error) {
	tasks := make(map[string]any)

	for _, listeners := range events {
		for _, listener := range listeners {
			if listener.Signature() == "" {
				return nil, errors.New("the Signature of listener can't be empty")
			}

			if tasks[listener.Signature()] != nil {
				continue
			}

			tasks[listener.Signature()] = listener.Handle
		}
	}

	return tasks, nil
}

func ArgsToValues(args []queue.Arg) ([]reflect.Value, error) {
	values := make([]reflect.Value, len(args))

	for i, arg := range args {
		v, err := convert.ReflectValue(arg.Type, arg.Value)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}

	return values, nil
}
