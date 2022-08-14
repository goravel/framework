package stubs

type ListenerStubs struct {
}

func (receiver ListenerStubs) Listener() string {
	return `package listeners

import (
	"github.com/goravel/framework/contracts/event"
)

type DummyListener struct {
}

func (receiver *DummyListener) Signature() string {
	return "DummyName"
}

func (receiver *DummyListener) Queue(args ...interface{}) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *DummyListener) Handle(args ...interface{}) error {
	return nil
}
`
}
