package event

import (
	"fmt"

	"github.com/goravel/framework/contracts/event"
	queuecontract "github.com/goravel/framework/contracts/queue"
)

type Task struct {
	args      []event.Arg
	event     event.Event
	listeners []event.Listener
	queue     queuecontract.Queue
}

func NewTask(queue queuecontract.Queue, args []event.Arg, event event.Event, listeners []event.Listener) *Task {
	return &Task{
		args:      args,
		event:     event,
		listeners: listeners,
		queue:     queue,
	}
}

func (receiver *Task) Dispatch() error {
	if len(receiver.listeners) == 0 {
		return fmt.Errorf("event %v doesn't bind listeners", receiver.event)
	}

	handledArgs, err := receiver.event.Handle(receiver.args)
	if err != nil {
		return err
	}

	var mapArgs []any
	for _, arg := range handledArgs {
		mapArgs = append(mapArgs, arg.Value)
	}

	for _, listener := range receiver.listeners {
		var err error
		task := receiver.queue.Job(listener, eventArgsToQueueArgs(handledArgs))
		queue := listener.Queue(mapArgs...)
		if queue.Connection != "" {
			task.OnConnection(queue.Connection)
		}
		if queue.Queue != "" {
			task.OnQueue(queue.Queue)
		}
		if queue.Enable {
			err = task.Dispatch()
		} else {
			err = task.DispatchSync()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func eventArgsToQueueArgs(args []event.Arg) []queuecontract.Arg {
	var queueArgs []queuecontract.Arg
	for _, arg := range args {
		queueArgs = append(queueArgs, queuecontract.Arg{
			Type:  arg.Type,
			Value: arg.Value,
		})
	}

	return queueArgs
}
