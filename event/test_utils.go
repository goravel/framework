package event

import (
	"errors"

	"github.com/goravel/framework/contracts/event"
)

var (
	testSyncListener        = 0
	testAsyncListener       = 0
	testCancelListener      = 0
	testCancelAfterListener = 0
)

type TestEvent struct{}

func (receiver *TestEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type TestEventNoRegister struct{}

func (receiver *TestEventNoRegister) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type TestEventHandleError struct{}

func (receiver *TestEventHandleError) Handle(args []event.Arg) ([]event.Arg, error) {
	return nil, errors.New("some errors")
}

type TestCancelEvent struct{}

func (receiver *TestCancelEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type TestListener struct{}

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

type TestListenerHandleError struct{}

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
	return errors.New("error")
}

type TestAsyncListener struct{}

func (receiver *TestAsyncListener) Signature() string {
	return "test_async_listener"
}

func (receiver *TestAsyncListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     true,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestAsyncListener) Handle(args ...any) error {
	testAsyncListener++

	return nil
}

type TestSyncListener struct{}

func (receiver *TestSyncListener) Signature() string {
	return "test_sync_listener"
}

func (receiver *TestSyncListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestSyncListener) Handle(args ...any) error {
	testSyncListener++

	return nil
}

type TestCancelListener struct{}

func (receiver *TestCancelListener) Signature() string {
	return "test_cancel_listener"
}

func (receiver *TestCancelListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestCancelListener) Handle(args ...any) error {
	testCancelListener++

	return errors.New("cancel")
}

type TestCancelAfterListener struct{}

func (receiver *TestCancelAfterListener) Signature() string {
	return "test_cancel_after_listener"
}

func (receiver *TestCancelAfterListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestCancelAfterListener) Handle(args ...any) error {
	testCancelAfterListener++

	return nil
}
