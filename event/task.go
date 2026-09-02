package event

import (
	"github.com/goravel/framework/contracts/event"
	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type Task struct {
	event     event.Event
	queue     contractsqueue.Queue
	err       error
	args      []event.Arg
	listeners []*listener
}

func NewTask(queue contractsqueue.Queue, args []event.Arg, evt event.Event, listeners []event.Listener) *Task {
	normalized := make([]*listener, 0, len(listeners))
	for _, l := range listeners {
		normalized = append(normalized, newLegacyListener(l, l.Signature()))
	}

	return newTask(queue, args, evt, normalized)
}

// newTask builds a task from listeners that are already normalized, which is how
// Job reaches the listeners registered through Listen.
func newTask(queue contractsqueue.Queue, args []event.Arg, evt event.Event, listeners []*listener) *Task {
	return &Task{
		args:      args,
		event:     evt,
		listeners: listeners,
		queue:     queue,
	}
}

// newFailedTask carries an error that only surfaces when the task is dispatched,
// Job has no way to report one itself.
func newFailedTask(err error) *Task {
	return &Task{err: err}
}

// Dispatch runs the event through the same pipeline as Dispatch, in the mode
// that keeps the deprecated behaviour: an unbound event is an error, and the
// first failing listener stops the ones behind it.
func (receiver *Task) Dispatch() error {
	if receiver.err != nil {
		return receiver.err
	}

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

// dispatchToQueue pushes a job onto the queue. A listener that doesn't enable
// queueing never reaches here, it runs in process.
func dispatchToQueue(queue contractsqueue.Queue, job contractsqueue.Job, options event.Queue, args []contractsqueue.Arg) error {
	task := queue.Job(job, args)

	if options.Connection != "" {
		task.OnConnection(options.Connection)
	}
	if options.Queue != "" {
		task.OnQueue(options.Queue)
	}

	return task.Dispatch()
}

// queueArgs builds the arguments of a queued listener, the event name leads the
// payload for listeners that expect it, since the queue only carries scalars.
func queueArgs(eventName string, args []event.Arg) []contractsqueue.Arg {
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
