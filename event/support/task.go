package support

import (
	"fmt"

	"github.com/RichardKnop/machinery/v2"
	"github.com/RichardKnop/machinery/v2/tasks"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue/support"
)

type Task struct {
	Events      map[event.Event][]event.Listener
	Event       event.Event
	Args        []event.Arg
	handledArgs []event.Arg
	mapArgs     []any
}

func (receiver *Task) Dispatch() error {
	listeners, ok := receiver.Events[receiver.Event]
	if !ok {
		return fmt.Errorf("event not found: %v", receiver.Event)
	}

	handledArgs, err := receiver.Event.Handle(receiver.Args)
	if err != nil {
		return err
	}

	receiver.handledArgs = handledArgs

	var mapArgs []any
	for _, arg := range receiver.handledArgs {
		mapArgs = append(mapArgs, arg.Value)
	}
	receiver.mapArgs = mapArgs

	for _, listener := range listeners {
		var err error
		queue := listener.Queue(receiver.mapArgs...)
		if queue.Enable {
			err = receiver.dispatchAsync(listener)
		} else {
			err = receiver.dispatchSync(listener)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (receiver *Task) dispatchAsync(listener event.Listener) error {
	queueServer, err := receiver.getQueueServer(listener)
	if err != nil {
		return err
	}
	if queueServer == nil {
		return receiver.dispatchSync(listener)
	}

	var args []tasks.Arg
	for _, arg := range receiver.handledArgs {
		args = append(args, tasks.Arg{
			Type:  arg.Type,
			Value: arg.Value,
		})
	}

	_, err = queueServer.SendTask(&tasks.Signature{
		Name: listener.Signature(),
		Args: args,
	})
	if err != nil {
		return err
	}

	return nil
}

func (receiver *Task) dispatchSync(listen event.Listener) error {
	return listen.Handle(receiver.mapArgs...)
}

func (receiver *Task) getQueueServer(listener event.Listener) (*machinery.Server, error) {
	queue := listener.Queue(receiver.mapArgs)
	connection := queue.Connection
	if connection == "" {
		connection = facades.Config.GetString("queue.default")
	}

	driver := facades.Config.GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
	if driver == support.DriverSync || driver == "" {
		return nil, nil
	}

	server, err := support.GetServer(connection, queue.Queue)
	if err != nil {
		return nil, err
	}

	return server, nil
}
