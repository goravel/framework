package event

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/goravel/framework/contracts/event"
)

// eventQueueWrapper wraps EventQueueListener to implement QueueListener interface
type eventQueueWrapper struct {
	listener event.EventQueueListener
	event    any
}

func (w *eventQueueWrapper) Handle(args ...any) error {
	spew.Dump("============", w.event)
	return w.listener.Handle(w.event, args...)
}

func (w *eventQueueWrapper) ShouldQueue() bool {
	return w.listener.ShouldQueue()
}

func (w *eventQueueWrapper) Signature() string {
	return w.listener.Signature()
}

func eventToQueueListener(listener event.EventQueueListener, event any) event.QueueListener {
	return &eventQueueWrapper{
		listener: listener,
		event:    event,
	}
}
