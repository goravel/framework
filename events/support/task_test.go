package support

import (
	"errors"
	"testing"

	"github.com/goravel/framework/contracts/events"
	"github.com/stretchr/testify/assert"
)

type TestEvent struct {
}

func (receiver *TestEvent) Signature() string {
	return "test_event"
}

func (receiver *TestEvent) Handle(args []events.Arg) ([]events.Arg, error) {
	return args, nil
}

type TestEventNoRegister struct {
}

func (receiver *TestEventNoRegister) Signature() string {
	return "test_event1"
}

func (receiver *TestEventNoRegister) Handle(args []events.Arg) ([]events.Arg, error) {
	return args, nil
}

type TestEventHandleError struct {
}

func (receiver *TestEventHandleError) Signature() string {
	return "test_event1"
}

func (receiver *TestEventHandleError) Handle(args []events.Arg) ([]events.Arg, error) {
	return nil, errors.New("some errors")
}

type TestListener struct {
}

func (receiver *TestListener) Signature() string {
	return "test_listener"
}

func (receiver *TestListener) Queue(args ...interface{}) events.Queue {
	return events.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListener) Handle(args ...interface{}) error {
	return nil
}

type TestListenerHandleError struct {
}

func (receiver *TestListenerHandleError) Signature() string {
	return "test_listener"
}

func (receiver *TestListenerHandleError) Queue(args ...interface{}) events.Queue {
	return events.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListenerHandleError) Handle(args ...interface{}) error {
	return errors.New("some errors")
}

func TestDispatch(t *testing.T) {
	var task Task

	tests := []struct {
		Name  string
		Setup func()
		Event events.Event
		err   bool
	}{
		{
			Name: "dispatchSync success",
			Setup: func() {
				task = Task{
					Events: map[events.Event][]events.Listener{
						&TestEvent{}: {
							&TestListener{},
						},
					},
					Event: &TestEvent{},
					Args: []events.Arg{
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
					Events: map[events.Event][]events.Listener{
						&TestEvent{}: {
							&TestListenerHandleError{},
						},
					},
					Event: &TestEvent{},
					Args: []events.Arg{
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
					Events: map[events.Event][]events.Listener{
						&TestEvent{}: {
							&TestListener{},
						},
					},
					Event: &TestEventNoRegister{},
					Args: []events.Arg{
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
					Events: map[events.Event][]events.Listener{
						&TestEventHandleError{}: {
							&TestListener{},
						},
					},
					Event: &TestEventHandleError{},
					Args: []events.Arg{
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
