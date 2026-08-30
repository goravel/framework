package event

import (
	"github.com/goravel/framework/contracts/event"
	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type Task struct {
	event     event.Event
	queue     contractsqueue.Queue
	args      []event.Arg
	listeners []*listener
}

func NewTask(queue contractsqueue.Queue, args []event.Arg, evt event.Event, listeners []event.Listener) *Task {
	normalized := make([]*listener, 0, len(listeners))
	for _, l := range listeners {
		normalized = append(normalized, newLegacyListener(l))
	}

	return &Task{
		args:      args,
		event:     evt,
		listeners: normalized,
		queue:     queue,
	}
}

// Dispatch runs the event through the same pipeline as Dispatch, in the mode
// that keeps the deprecated behaviour: an unbound event is an error, and the
// first failing listener stops the ones behind it.
func (receiver *Task) Dispatch() error {
	name, err := getEventName(receiver.event)
	if err != nil {
		return err
	}

	errs := dispatch(receiver.queue, receiver.event, name, receiver.listeners, receiver.args, dispatchModeTask)
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

// dispatchToQueue pushes a job onto the queue, or runs it synchronously through
// the queue facade when the listener doesn't enable queueing.
func dispatchToQueue(queue contractsqueue.Queue, job contractsqueue.Job, options event.Queue, args []contractsqueue.Arg) error {
	task := queue.Job(job, args)

	if options.Connection != "" {
		task.OnConnection(options.Connection)
	}
	if options.Queue != "" {
		task.OnQueue(options.Queue)
	}

	if options.Enable {
		return task.Dispatch()
	}

	return task.DispatchSync()
}

// queueArgs builds the arguments of a queued listener, the event name leads the
// payload for listeners that expect it, since the queue only carries scalars.
func queueArgs(l *listener, eventName string, args []event.Arg) []contractsqueue.Arg {
	if !l.withEvent {
		return eventArgsToQueueArgs(args)
	}

	queued := make([]contractsqueue.Arg, 0, len(args)+1)
	queued = append(queued, contractsqueue.Arg{Type: "string", Value: eventName})
	for _, arg := range args {
		queued = append(queued, contractsqueue.Arg{Type: arg.Type, Value: arg.Value})
	}

	return queued
}

// argValues extracts the values of event arguments, they are what listeners and
// queue options receive.
func argValues(args []event.Arg) []any {
	values := make([]any, 0, len(args))
	for _, arg := range args {
		values = append(values, arg.Value)
	}

	return values
}

// eventArgsToQueueArgs converts event arguments to queue arguments, the two
// systems carry the same Type and Value pair behind different types.
func eventArgsToQueueArgs(args []event.Arg) []contractsqueue.Arg {
	var queueArgs []contractsqueue.Arg
	for _, arg := range args {
		queueArgs = append(queueArgs, contractsqueue.Arg{
			Type:  arg.Type,
			Value: arg.Value,
		})
	}

	return queueArgs
}
