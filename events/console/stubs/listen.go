package stubs

type ListenerStubs struct {
}

func (receiver ListenerStubs) Listener() string {
	return `package listeners

import (
	"github.com/goravel/framework/contracts/events"
)

type DummyListener struct {
}

func (receiver *DummyListener) Signature() string {
	return "DummyName"
}

func (receiver *DummyListener) Queue(args ...interface{}) events.Queue {
	return events.Queue{
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
