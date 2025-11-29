package event

import (
	"reflect"
	"strings"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

// Listen registers event listeners for one or more events.
// It supports various event types and listener formats:
//   - Events: string, []string, event.Event, []event.Event, or any custom type
//   - Listeners: functions, structs implementing listener interfaces, or closures
//
// The method automatically handles queue registration for listeners that implement ShouldQueue.
// Wildcard patterns (containing "*") are supported for flexible event matching.
//
// Examples:
//
//	// Listen to a single string event
//	app.Listen("user.created", func(args ...any) error {
//		return nil
//	})
//
//	// Listen to multiple events
//	app.Listen([]string{"user.created", "user.updated"}, listener)
//
//	// Listen to wildcard events
//	app.Listen("user.*", func(args ...any) error {
//		return nil
//	})
//
//	// Listen with event structs
//	app.Listen(UserCreatedEvent{}, &UserCreatedListener{})
func (app *Application) Listen(events any, listener ...any) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	if len(listener) == 0 {
		// it may be a closure
		if fn := reflect.ValueOf(events); fn.Kind() == reflect.Func {
			return app.handleClosure(fn)
		}

		return errors.New("listener is required")
	}

	switch e := events.(type) {
	case string:
		app.setupEvents(e, listener[0])
	case []string:
		for _, eventName := range e {
			app.setupEvents(eventName, listener[0])
		}
	case []event.Event:
		for _, evt := range e {
			app.setupEvents(evt, listener[0])
		}
	case event.Event:
		app.setupEvents(e, listener[0])
	default:
		if eventName := app.getEventName(events); eventName != "" {
			app.setupEvents(eventName, listener[0])
		} else {
			return errors.New("invalid event type")
		}
	}

	return nil
}

func (app *Application) handleClosure(fn reflect.Value) error {
	fnType := fn.Type()

	if fnType.NumIn() != 1 || !app.isEventType(fnType.In(0)) {
		return errors.New("closure must accept exactly one event parameter")
	}

	// get first parameter and pass to setupEvents
	ptr := reflect.New(fnType.In(0).Elem()).Interface()
	app.setupEvents(ptr, fn)

	return nil
}

func (app *Application) isEventType(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Implements(reflect.TypeOf((*event.Event)(nil)).Elem())
}

func (app *Application) setupEvents(e any, listener any) {
	eventName := app.getEventName(e)

	// Register queued listeners with the queue system
	if l, ok := listener.(event.EventQueueListener); ok {
		// Capture event in local scope to avoid closure variable capture bug
		eventCopy := e
		app.queue.Register([]queue.Job{
			&dynamicQueueJob{
				signature: l.Signature(),
				handler: func(queueArgs ...any) error {
					// Convert queue args back to event args format
					eventArgs := make([]any, len(queueArgs))
					copy(eventArgs, queueArgs)
					return l.Handle(eventCopy, eventArgs...)
				},
				shouldQueue: l.ShouldQueue(),
				queueConfig: l,
			},
		})
	} else if l, ok := listener.(event.QueueListener); ok {
		app.queue.Register([]queue.Job{l})
	}

	// Register the listener based on event type
	if strings.Contains(eventName, "*") {
		app.setupWildcardListen(eventName, listener)
	} else {
		app.listeners[eventName] = append(app.listeners[eventName], listener)
	}
}

func (app *Application) setupWildcardListen(eventName string, listener any) {
	app.wildcards[eventName] = append(app.wildcards[eventName], listener)
	// Clear cache to ensure fresh wildcard matching
	app.wildcardsCache.Range(func(key, value any) bool {
		app.wildcardsCache.Delete(key)
		return true
	})
}
