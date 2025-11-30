package event

import (
	"reflect"

	"github.com/goravel/framework/contracts/event"
)

// dynamicQueueJob - Minimal queue job implementation using closures
type dynamicQueueJob struct {
	signature   string
	handler     func(...any) error
	shouldQueue bool
	queueConfig any // Original listener for configuration
}

// Handle processes the job with the given arguments
func (j *dynamicQueueJob) Handle(args ...any) error {
	return j.handler(args...)
}

// Queue method using reflection to call original listener's Queue method
func (j *dynamicQueueJob) Queue(args ...any) event.Queue {
	if queueable, ok := j.queueConfig.(interface{ Queue(...any) event.Queue }); ok {
		return queueable.Queue(args...)
	}
	return event.Queue{}
}

// ShouldQueue returns whether this job should be queued
func (j *dynamicQueueJob) ShouldQueue() bool {
	return j.shouldQueue
}

// Signature returns the unique identifier for this job
func (j *dynamicQueueJob) Signature() string {
	return j.signature
}

// getEventName returns the name of the event.
// examples:
//
//	getEventName("user.created") // "user.created"
//	getEventName(&UserCreated{}) // "UserCreated"
func (app *Application) getEventName(evt any) string {
	switch e := evt.(type) {
	case string:
		return e
	case event.Event:
		// Check if event has custom signature method
		if signature, ok := e.(event.Signature); ok {
			return signature.Signature()
		}

		// Use reflection to get struct type name
		eventType := reflect.TypeOf(e)
		if eventType.Kind() == reflect.Ptr {
			eventType = eventType.Elem()
		}

		return eventType.Name()
	default:
		// Use reflection for any other type
		eventType := reflect.TypeOf(evt)
		if eventType != nil {
			if eventType.Kind() == reflect.Ptr {
				eventType = eventType.Elem()
			}
			// For unnamed structs, use the full type string
			name := eventType.Name()
			if name == "" {
				name = eventType.String()
			}
			return name
		}
		return ""
	}
}
