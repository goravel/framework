package console

type EventStubs struct {
}

func (receiver EventStubs) Event() string {
	return `package events

import "github.com/goravel/framework/contracts/event"

type DummyEvent struct {
}

func (receiver *DummyEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}
`
}
