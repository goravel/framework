package event

import (
	stderrors "errors"
	"reflect"
	"strings"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// listener is the internal representation every accepted listener form is
// normalized to when it is registered, so that dispatching stays free of
// reflection and type switches.
type listener struct {
	// handle invokes the listener in the current goroutine, it is nil for
	// listeners that always go through the queue.
	handle func(evt any, args []event.Arg) error
	// job is pushed onto the queue when the listener is queueable, nil otherwise.
	job queue.Job
	// queueOptions returns the queue options of the listener.
	queueOptions func(args ...any) event.Queue
	// signature is the unique identifier of the listener, empty for closures.
	signature string
	// wildcard reports whether the listener was registered on a wildcard pattern,
	// those listeners receive the event name instead of the event itself.
	wildcard bool
	// withEvent reports whether the event name is prepended to the queue arguments.
	withEvent bool
}

// queueJob adapts an event.QueueListener to the queue.Job interface, the queue
// only carries scalar arguments so the event name travels as the first one.
type queueJob struct {
	listener event.QueueListener
}

func (j *queueJob) Signature() string {
	return j.listener.Signature()
}

func (j *queueJob) Handle(args ...any) error {
	if len(args) == 0 {
		return errors.EventQueueMissingEvent.Args(j.listener.Signature())
	}

	return j.listener.Handle(args[0], args[1:]...)
}

// Listen registers one or more listeners for one or more events, see
// event.Instance for the accepted event and listener forms.
func (app *Application) Listen(events any, listeners ...any) error {
	if len(listeners) == 0 {
		return app.listenClosure(events)
	}

	names, err := eventNames(events)
	if err != nil {
		return err
	}

	var errs []error
	for _, name := range names {
		for _, l := range listeners {
			normalized, err := newListener(name, l)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			app.addListener(name, normalized)
		}
	}

	return stderrors.Join(errs...)
}

// listenClosure registers a func(event *SomeEvent) error closure, the event is
// resolved from the type of its only parameter.
func (app *Application) listenClosure(events any) error {
	fn := reflect.ValueOf(events)
	if fn.Kind() != reflect.Func {
		return errors.EventInvalidListener.Args(typeName(reflect.TypeOf(events)))
	}

	name, err := closureEventName(fn.Type())
	if err != nil {
		return err
	}

	app.addListener(name, newClosureListener(fn))

	return nil
}

func (app *Application) addListener(name string, l *listener) {
	app.mu.Lock()
	defer app.mu.Unlock()

	if l.job != nil && !app.registered[l.signature] {
		app.registered[l.signature] = true
		app.queue.Register([]queue.Job{l.job})
	}

	if strings.Contains(name, "*") {
		l.wildcard = true
		app.wildcards[name] = append(app.wildcards[name], l)
		// The cache maps event names to matching wildcard listeners, a new
		// pattern may match any of them.
		app.wildcardsCache.Clear()

		return
	}

	app.listeners[name] = append(app.listeners[name], l)
}

// newListener normalizes an event.QueueListener or a closure into a listener.
func newListener(eventName string, l any) (*listener, error) {
	switch v := l.(type) {
	case event.QueueListener:
		return &listener{
			handle: func(evt any, args []event.Arg) error {
				return v.Handle(evt, argValues(args)...)
			},
			job:          &queueJob{listener: v},
			queueOptions: v.Queue,
			signature:    v.Signature(),
			withEvent:    true,
		}, nil
	case func(evt any, args ...any) error:
		return &listener{
			handle: func(evt any, args []event.Arg) error {
				return v(evt, argValues(args)...)
			},
		}, nil
	}

	if fn := reflect.ValueOf(l); fn.Kind() == reflect.Func {
		if _, err := closureEventName(fn.Type()); err != nil {
			return nil, errors.EventInvalidListener.Args(eventName)
		}

		return newClosureListener(fn), nil
	}

	return nil, errors.EventInvalidListener.Args(eventName)
}

// newLegacyListener wraps a listener registered through the deprecated Register,
// those listeners are always executed through the queue facade.
func newLegacyListener(l event.Listener) *listener {
	return &listener{
		job:          l,
		queueOptions: l.Queue,
	}
}

// newClosureListener wraps a func(event *SomeEvent) error closure, the closure
// type has already been validated by closureEventName.
func newClosureListener(fn reflect.Value) *listener {
	param := fn.Type().In(0)

	return &listener{
		handle: func(evt any, args []event.Arg) error {
			value := reflect.ValueOf(evt)
			if !value.IsValid() || !value.Type().AssignableTo(param) {
				return errors.EventInvalidEvent.Args(evt)
			}

			if err, ok := fn.Call([]reflect.Value{value})[0].Interface().(error); ok {
				return err
			}

			return nil
		},
	}
}

// closureEventName validates a func(event *SomeEvent) error closure and returns
// the name of the event its parameter refers to.
func closureEventName(t reflect.Type) (string, error) {
	if t.NumIn() != 1 || t.IsVariadic() || t.NumOut() != 1 || t.Out(0) != errorType {
		return "", errors.EventInvalidListener.Args(typeName(t))
	}

	return typeName(t.In(0)), nil
}

// eventNames resolves the events to listen on to their names.
func eventNames(events any) ([]string, error) {
	switch e := events.(type) {
	case string:
		name, err := getEventName(e)

		return []string{name}, err
	case []string:
		return collectEventNames(e)
	case []event.Event:
		return collectEventNames(e)
	case []any:
		return collectEventNames(e)
	default:
		name, err := getEventName(events)

		return []string{name}, err
	}
}

func collectEventNames[T any](events []T) ([]string, error) {
	names := make([]string, 0, len(events))
	for _, e := range events {
		name, err := getEventName(e)
		if err != nil {
			return nil, err
		}

		names = append(names, name)
	}

	return names, nil
}

// getEventName returns the name of an event, a string event is its own name and
// any other event is named after its package qualified type.
func getEventName(evt any) (string, error) {
	if name, ok := evt.(string); ok {
		if name == "" {
			return "", errors.EventInvalidEvent.Args(evt)
		}

		return name, nil
	}

	name := typeName(reflect.TypeOf(evt))
	if name == "" {
		return "", errors.EventInvalidEvent.Args(evt)
	}

	return name, nil
}

// typeName returns the package qualified name of a type, pointers are followed
// so that *events.UserCreated and events.UserCreated share a name.
func typeName(t reflect.Type) string {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t == nil {
		return ""
	}

	return t.String()
}
