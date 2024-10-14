package queue

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

func jobs2Tasks(jobs []queue.Job) (map[string]any, error) {
	tasks := make(map[string]any)

	for _, job := range jobs {
		if job.Signature() == "" {
			return nil, errors.QueueEmptyJobSignature
		}

		if tasks[job.Signature()] != nil {
			return nil, errors.QueueDuplicateJobSignature.Args(job.Signature())
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
				return nil, errors.QueueEmptyListenerSignature
			}

			if tasks[listener.Signature()] != nil {
				continue
			}

			tasks[listener.Signature()] = listener.Handle
		}
	}

	return tasks, nil
}
