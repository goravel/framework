package console

type Stubs struct {
}

func (receiver Stubs) Event() string {
	return `package DummyPackage

import "github.com/goravel/framework/contracts/event"

type DummyEvent struct {
}

func (receiver *DummyEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}
`
}

func (receiver Stubs) EventBroadcast() string {
	return `package DummyPackage

import "github.com/goravel/framework/contracts/broadcasting"

var _ broadcasting.ShouldBroadcast = (*DummyEvent)(nil)

type DummyEvent struct {
}

func (receiver *DummyEvent) BroadcastOn() []string {
	return []string{}
}

func (receiver *DummyEvent) BroadcastAs() string {
	return ""
}

func (receiver *DummyEvent) BroadcastWith() map[string]any {
	return map[string]any{}
}

func (receiver *DummyEvent) BroadcastWhen() bool {
	return true
}
`
}

func (receiver Stubs) EventBroadcastNow() string {
	return receiver.EventBroadcast() + `
func (receiver *DummyEvent) BroadcastNow() bool {
	return true
}
`
}

func (receiver Stubs) Listener() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/event"
)

type DummyListener struct {
}

func (receiver *DummyListener) Signature() string {
	return "DummyName"
}

func (receiver *DummyListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *DummyListener) Handle(eventName string, args ...any) error {
	return nil
}
`
}
