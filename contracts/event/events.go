package event

type Instance interface {
	// Dispatch fires an event and calls all the listeners registered for it.
	Dispatch(event any, args ...[]Arg) Result
	// Listen registers one or more listeners for one or more events.
	// events can be a string, []string, Event, []Event or []any, and a wildcard
	// pattern such as "user.*" matches every event sharing the prefix.
	// listeners can be Listener implementations, func(event any, args ...any)
	// error closures, or func(event *SomeEvent) error closures. When no listener
	// is given, events must be the last closure form and the event is resolved
	// from its parameter type. That form can also be passed explicitly when its
	// parameter matches the registered event.
	Listen(events any, listeners ...any) error

	// GetEvents gets all registered events.
	//
	// Deprecated: Use Listen instead, GetEvents will be removed in a future version.
	GetEvents() map[Event][]Listener
	// Job create a new event task.
	//
	// Deprecated: Use Dispatch instead, Job will be removed in a future version.
	Job(event Event, args []Arg) Task
	// Register event listeners to the application.
	//
	// Deprecated: Use Listen instead, Register will be removed in a future version.
	Register(map[Event][]Listener)
}

type Event interface {
	// Handle the event.
	Handle(args []Arg) ([]Arg, error)
}

// Listener handles a dispatched event. It is registered through Listen, and
// through the deprecated Register.
type Listener interface {
	// Handle the event. A Listener receives only the canonical event name, never
	// the event object itself. Any event data it needs must travel in args so that
	// it behaves the same in process and through the queue.
	Handle(eventName string, args ...any) error
	// Queue configure the event queue options, the listener is pushed onto the
	// queue instead of running synchronously when Queue().Enable is true.
	Queue(args ...any) Queue
	// Signature returns the unique identifier for the listener.
	Signature() string
}

// Result aggregates the errors returned by the listeners of a single dispatch.
type Result interface {
	// Error returns all the listener errors joined into one, or nil when none failed.
	Error() error
	// Errors returns the error of each failed listener.
	Errors() []error
	// Failed reports whether any listener failed.
	Failed() bool
}

type Task interface {
	// Dispatch an event and call the listeners.
	Dispatch() error
}

type Arg struct {
	Value any
	Type  string
}

type Queue struct {
	Connection string
	Queue      string
	Enable     bool
}
