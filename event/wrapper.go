package event

import (
	"github.com/goravel/framework/contracts/event"
)

type eventQueueWrapper struct {
	listener event.EventQueueListener
	event    any
}

// Handle processes the event using the wrapped EventQueueListener.
func (w *eventQueueWrapper) Handle(args ...any) error {
	return w.listener.Handle(w.event, args...)
}

// ShouldQueue determines if this listener should be queued.
func (w *eventQueueWrapper) ShouldQueue() bool {
	return w.listener.ShouldQueue()
}

// Signature returns the unique identifier for this listener.
func (w *eventQueueWrapper) Signature() string {
	return w.listener.Signature()
}

func eventToQueueListener(listener event.EventQueueListener, event any) event.QueueListener {
	return &eventQueueWrapper{
		listener: listener,
		event:    event,
	}
}
