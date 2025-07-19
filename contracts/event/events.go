package event

type Arg struct {
	Type  string
	Value any
}

type Event interface {
	// Handle the event.
	Handle(args []Arg) ([]Arg, error)
}

type EventListener interface {
	Handle(event any, args ...any) error
}

type EventQueueListener interface {
	// Queue configure the event queue options.
	Queue(args ...any) Queue
	EventListener
	Signature
	ShouldQueue
}

type Instance interface {
	// Dispatch fires an event and calls the listeners.
	Dispatch(event any, payload ...[]Arg) []any
	// GetEvents gets all registered events.
	GetEvents() map[Event][]Listener
	// Job create a new event task.
	Job(event Event, args []Arg) Task
	// Listen registers an event listener with the dispatcher.
	// events can be: string, []string, Event, []Event, or any other type
	// listener can be: function, class, or any callable
	Listen(events any, listener ...any) error
	// Register event listeners to the application.
	Register(map[Event][]Listener)
}

type Listener interface {
	// Handle the event.
	Handle(args ...any) error
	// Queue configure the event queue options.
	Queue(args ...any) Queue
	// Signature returns the unique identifier for the listener.
	Signature() string
}

type Queue struct {
	Connection string
	Enable     bool
	Queue      string
}

type QueueListener interface {
	Listener
	Signature
	ShouldQueue
}

type ShouldQueue interface {
	ShouldQueue() bool
}

type Signature interface {
	Signature() string
}

type Task interface {
	// Dispatch an event and call the listeners.
	Dispatch() error
}
