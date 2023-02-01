package support

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/event"
)

type TestEvent struct {
}

func (receiver *TestEvent) Signature() string {
	return "test_event"
}

func (receiver *TestEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type TestEventNoRegister struct {
}

func (receiver *TestEventNoRegister) Signature() string {
	return "test_event1"
}

func (receiver *TestEventNoRegister) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type TestEventHandleError struct {
}

func (receiver *TestEventHandleError) Signature() string {
	return "test_event1"
}

func (receiver *TestEventHandleError) Handle(args []event.Arg) ([]event.Arg, error) {
	return nil, errors.New("some errors")
}

type TestListener struct {
}

func (receiver *TestListener) Signature() string {
	return "test_listener"
}

func (receiver *TestListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListener) Handle(args ...any) error {
	return nil
}

type TestListenerHandleError struct {
}

func (receiver *TestListenerHandleError) Signature() string {
	return "test_listener"
}

func (receiver *TestListenerHandleError) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListenerHandleError) Handle(args ...any) error {
	return errors.New("some errors")
}

func TestDispatch(t *testing.T) {
	var task Task

	tests := []struct {
		Name  string
		Setup func()
		Event event.Event
		err   bool
	}{
		{
			Name: "dispatchSync success",
			Setup: func() {
				task = Task{
					Events: map[event.Event][]event.Listener{
						&TestEvent{}: {
							&TestListener{},
						},
					},
					Event: &TestEvent{},
					Args: []event.Arg{
						{Type: "sting", Value: "test"},
					},
				}
			},
			err: false,
		},
		{
			Name: "dispatchSync error",
			Setup: func() {
				task = Task{
					Events: map[event.Event][]event.Listener{
						&TestEvent{}: {
							&TestListenerHandleError{},
						},
					},
					Event: &TestEvent{},
					Args: []event.Arg{
						{Type: "sting", Value: "test"},
					},
				}
			},
			Event: &TestEvent{},
			err:   true,
		},
		{
			Name: "event not found",
			Setup: func() {
				task = Task{
					Events: map[event.Event][]event.Listener{
						&TestEvent{}: {
							&TestListener{},
						},
					},
					Event: &TestEventNoRegister{},
					Args: []event.Arg{
						{Type: "sting", Value: "test"},
					},
				}
			},
			err: true,
		},
		{
			Name: "event handle return error",
			Setup: func() {
				task = Task{
					Events: map[event.Event][]event.Listener{
						&TestEventHandleError{}: {
							&TestListener{},
						},
					},
					Event: &TestEventHandleError{},
					Args: []event.Arg{
						{Type: "sting", Value: "test"},
					},
				}
			},
			err: true,
		},
	}

	for _, test := range tests {
		test.Setup()
		err := task.Dispatch()
		assert.Equal(t, test.err, err != nil, test.Name)
	}
}
