package event

import (
	"github.com/goravel/framework/contracts/event"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

type Task struct {
	event     event.Event
	queue     contractsqueue.Queue
	args      []event.Arg
	listeners []event.Listener
}

func NewTask(queue contractsqueue.Queue, args []event.Arg, event event.Event, listeners []event.Listener) *Task {
	return &Task{
		args:      args,
		event:     event,
		listeners: listeners,
		queue:     queue,
	}
}

func (receiver *Task) Dispatch() error {
	if len(receiver.listeners) == 0 {
		return errors.EventListenerNotBind.Args(receiver.event)
	}

	handledArgs, err := receiver.event.Handle(receiver.args)
	if err != nil {
		return err
	}

	var (
		values    = argValues(handledArgs)
		queueArgs = eventArgsToQueueArgs(handledArgs)
	)

	for _, listener := range receiver.listeners {
		if err := dispatchToQueue(receiver.queue, listener, listener.Queue(values...), queueArgs); err != nil {
			return err
		}
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
