package event

type Instance interface {
	// Dispatch fires an event and calls all the listeners registered for it.
	Dispatch(event any, args ...[]Arg) Result
	// Listen registers one or more listeners for one or more events.
	// events can be a string, []string, Event, []Event or []any, and a wildcard
	// pattern such as "user.*" matches every event sharing the prefix.
	// listeners can be QueueListener implementations or
	// func(event any, args ...any) error closures. When no listener is given,
	// events must be a func(event *SomeEvent) error closure, and the event is
	// resolved from the parameter type.
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

// QueueListener is the listener interface used by Listen and Dispatch.
// Listeners registered through the deprecated Register keep using Listener.
type QueueListener interface {
	// Handle the event. eventName is the canonical name of the dispatched event,
	// never the event itself, so that a listener behaves the same whether it runs
	// in process or through the queue, where only scalar arguments survive.
	// The data a queued listener needs must therefore travel in args.
	Handle(eventName string, args ...any) error
	// Queue configure the event queue options, the listener is pushed onto the
	// queue instead of running synchronously when Queue().Enable is true.
	Queue(args ...any) Queue
	// Signature returns the unique identifier for the listener.
	Signature() string
}

// Listener is the listener interface used by the deprecated Register and Job.
type Listener interface {
	// Signature returns the unique identifier for the listener.
	Signature() string
	// Queue configure the event queue options.
	Queue(args ...any) Queue
	// Handle the event.
	Handle(args ...any) error
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
