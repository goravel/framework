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
// normalized to when it is registered, so that dispatching stays free of type
// switches.
type listener struct {
	// handle invokes the listener in the current goroutine, it is nil for
	// listeners that always go through the queue.
	handle func(eventName string, evt any, args []event.Arg) error
	// job is pushed onto the queue when the listener is queueable, nil otherwise.
	job queue.Job
	// queueOptions returns the queue options of the listener.
	queueOptions func(args ...any) event.Queue
	// signature is the unique identifier of the listener, empty for closures.
	signature string
	// identity distinguishes listeners sharing a signature, see listenerIdentity.
	identity any
	// legacy reports whether the listener came from the deprecated Register.
	legacy bool
	// wildcard reports whether the listener was registered on a wildcard pattern,
	// those listeners receive the event name instead of the event itself.
	wildcard bool
	// withEvent reports whether the event name is prepended to the queue arguments.
	withEvent bool
}

// wildcardEntry keeps wildcard registrations ordered, a map would make the
// invocation order of overlapping patterns such as "user.*" and "*.created"
// depend on Go's map iteration.
type wildcardEntry struct {
	pattern   string
	listeners []*listener
}

// queueJob adapts an event.QueueListener to the queue.Job interface, the queue
// only carries scalar arguments so the event name travels as the first one.
type queueJob struct {
	listener event.QueueListener
}

func (j *queueJob) Signature() string {
	return j.listener.Signature()
}

// Handle runs the listener from a queue worker. It recovers from a panic so that
// a faulty listener fails the job through the queue's own retry and failure
// handling instead of taking the worker process down.
func (j *queueJob) Handle(args ...any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.EventListenerPanic.Args(j.listener.Signature(), r)
		}
	}()

	if len(args) == 0 {
		return errors.EventQueueMissingEvent.Args(j.listener.Signature())
	}

	eventName, ok := args[0].(string)
	if !ok {
		return errors.EventQueueMissingEvent.Args(j.listener.Signature())
	}

	return j.listener.Handle(eventName, args[1:]...)
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

	// The whole request is normalized before anything is registered, so that an
	// invalid listener can't leave half of the events wired up.
	normalized := make([]*listener, 0, len(listeners))
	var errs []error
	for _, l := range listeners {
		n, err := newListener(names, l)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		normalized = append(normalized, n)
	}

	if len(errs) > 0 {
		return stderrors.Join(errs...)
	}

	for _, name := range names {
		for _, n := range normalized {
			// Each event gets its own listener value, they carry per-registration
			// state such as the wildcard flag.
			if err := app.addListener(name, n.clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return stderrors.Join(errs...)
}

// listenClosure registers a func(event *SomeEvent) error closure, the event is
// resolved from the type of its only parameter.
func (app *Application) listenClosure(events any) error {
	fn := reflect.ValueOf(events)
	if fn.Kind() != reflect.Func || fn.IsNil() {
		return errors.EventInvalidListener.Args(typeName(reflect.TypeOf(events)))
	}

	name, err := closureEventName(fn.Type())
	if err != nil {
		return err
	}

	return app.addListener(name, newClosureListener(fn))
}

func (app *Application) addListener(name string, l *listener) error {
	job, err := app.claimQueueJob(l)
	if err != nil {
		return err
	}

	// Registering with the queue calls into code outside this module, it must not
	// run while the registry is locked.
	if job != nil {
		app.queue.Register([]queue.Job{job})
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	if strings.Contains(name, "*") {
		l.wildcard = true
		for i := range app.wildcards {
			if app.wildcards[i].pattern == name {
				app.wildcards[i].listeners = append(app.wildcards[i].listeners, l)

				return nil
			}
		}

		app.wildcards = append(app.wildcards, wildcardEntry{pattern: name, listeners: []*listener{l}})

		return nil
	}

	app.listeners[name] = append(app.listeners[name], l)

	return nil
}

// claimQueueJob reserves the listener's signature, returning the job to register
// with the queue, or nil when the listener isn't queueable or its job is already
// registered. The queue resolves jobs by signature alone, so two different
// listeners sharing one signature would silently run each other's code.
func (app *Application) claimQueueJob(l *listener) (queue.Job, error) {
	if l.job == nil {
		return nil, nil
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	if registered, ok := app.registered[l.signature]; ok {
		if registered != l.identity {
			return nil, errors.EventQueueDuplicateSignature.Args(l.signature)
		}

		return nil, nil
	}

	app.registered[l.signature] = l.identity

	return l.job, nil
}

// listenerIdentity returns a comparable identity for a listener, so that
// registering the same listener on several events is not mistaken for two
// listeners fighting over one signature. Listen only accepts pointers, the type
// fallback exists for the deprecated Register, which deduplicates by signature
// within each of its own calls.
func listenerIdentity(l any) any {
	if value := reflect.ValueOf(l); value.Kind() == reflect.Pointer {
		return value.Pointer()
	}

	return reflect.TypeOf(l)
}

func (l *listener) clone() *listener {
	cloned := *l

	return &cloned
}

// newListener normalizes an event.QueueListener or a closure into a listener.
// The event names are only used to reject listeners that could never be called.
func newListener(eventNames []string, l any) (*listener, error) {
	switch v := l.(type) {
	case event.QueueListener:
		// The queue resolves a job by its signature, so listeners sharing one must
		// be distinguishable. Two values of the same type never are, which would
		// let one listener be queued and another executed.
		if value := reflect.ValueOf(v); value.Kind() != reflect.Pointer || value.IsNil() {
			return nil, errors.EventListenerNotPointer.Args(v.Signature())
		}

		return &listener{
			handle: func(eventName string, evt any, args []event.Arg) error {
				return v.Handle(eventName, argValues(args)...)
			},
			job:          &queueJob{listener: v},
			identity:     listenerIdentity(v),
			queueOptions: v.Queue,
			signature:    v.Signature(),
			withEvent:    true,
		}, nil
	case func(evt any, args ...any) error:
		return &listener{
			handle: func(eventName string, evt any, args []event.Arg) error {
				return v(evt, argValues(args)...)
			},
		}, nil
	}

	fn := reflect.ValueOf(l)
	if fn.Kind() != reflect.Func || fn.IsNil() {
		return nil, errors.EventInvalidListener.Args(strings.Join(eventNames, ", "))
	}

	// A typed closure can only ever be called for the event its parameter names,
	// registering it on any other event, or on a pattern, is a mistake.
	closureName, err := closureEventName(fn.Type())
	if err != nil {
		return nil, errors.EventInvalidListener.Args(strings.Join(eventNames, ", "))
	}

	for _, name := range eventNames {
		if name != closureName {
			return nil, errors.EventListenerEventMismatch.Args(closureName, name)
		}
	}

	return newClosureListener(fn), nil
}

// newLegacyListener wraps a listener registered through the deprecated Register,
// those listeners are always executed through the queue facade.
func newLegacyListener(l event.Listener) *listener {
	return &listener{
		job:          l,
		legacy:       true,
		queueOptions: l.Queue,
	}
}

// newClosureListener wraps a func(event *SomeEvent) error closure, the closure
// type has already been validated by closureEventName.
func newClosureListener(fn reflect.Value) *listener {
	param := fn.Type().In(0)

	return &listener{
		handle: func(eventName string, evt any, args []event.Arg) error {
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

	name := typeName(t.In(0))
	if name == "" {
		return "", errors.EventInvalidListener.Args(typeName(t))
	}

	return name, nil
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

// matchWildcard reports whether a name matches a pattern in which "*" stands for
// any sequence of characters. It allocates nothing, unlike compiling the pattern
// into a regular expression on every match.
func matchWildcard(pattern, name string) bool {
	if pattern == name {
		return true
	}

	// The segments between the asterisks have to appear in order, the first one
	// anchored at the start and the last one at the end.
	first := true
	for {
		star := strings.IndexByte(pattern, '*')
		if star < 0 {
			break
		}

		segment := pattern[:star]
		if first {
			if !strings.HasPrefix(name, segment) {
				return false
			}
		} else {
			index := strings.Index(name, segment)
			if index < 0 {
				return false
			}

			name = name[index:]
		}

		name = name[len(segment):]
		pattern = pattern[star+1:]
		first = false
	}

	if first {
		return pattern == name
	}

	return strings.HasSuffix(name, pattern)
}
