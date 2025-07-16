// Package event wrapper provides utility functions for converting EventQueueListener
// to QueueListener interfaces, enabling event listeners to be processed through
// the queue system while maintaining event context.
package event

import (
	"github.com/goravel/framework/contracts/event"
)

// eventQueueWrapper wraps an EventQueueListener to make it compatible with QueueListener interface.
// This allows EventQueueListener instances to be queued while preserving the original event context.
type eventQueueWrapper struct {
	listener event.EventQueueListener // The original event queue listener
	event    any                      // The event that triggered this listener
}

// Handle processes the event using the wrapped EventQueueListener.
// It passes the original event along with the provided arguments to maintain context.
//
// Example:
//
//	wrapper := &eventQueueWrapper{listener: myListener, event: "user.created"}
//	err := wrapper.Handle("user_data")
func (w *eventQueueWrapper) Handle(args ...any) error {
	return w.listener.Handle(w.event, args...)
}

// ShouldQueue determines if this listener should be queued by delegating to the wrapped listener.
// This preserves the original listener's queueing preference.
func (w *eventQueueWrapper) ShouldQueue() bool {
	return w.listener.ShouldQueue()
}

// Signature returns the unique identifier for this listener by delegating to the wrapped listener.
// This ensures the wrapped listener's signature is preserved for queue management.
func (w *eventQueueWrapper) Signature() string {
	return w.listener.Signature()
}

// eventToQueueListener converts an EventQueueListener to a QueueListener by wrapping it
// with the event context. This enables EventQueueListener instances to be processed
// through the standard queue system while maintaining their event awareness.
//
// Example:
//
//	queueListener := eventToQueueListener(myEventListener, "user.created")
//	queue.Job(queueListener).Dispatch()
func eventToQueueListener(listener event.EventQueueListener, evt any) event.QueueListener {
	return &eventQueueWrapper{
		listener: listener,
		event:    evt,
	}
}
