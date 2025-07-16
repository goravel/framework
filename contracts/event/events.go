package event

type Event interface{}

type Instance interface {
	// Dispatch fires an event and calls the listeners.
	Dispatch(event any, payload ...[]Arg) []any

	// Flush flushes all pushed events.
	Flush(event any)

	// Forget removes all listeners for an event.
	Forget(event any)

	// ForgetPushed removes all pushed events.
	ForgetPushed()

	// GetListeners gets all listeners for an event.
	GetListeners(event any) []any

	// HasListeners determines if a given event has listeners.
	HasListeners(event any) bool

	// HasWildcardListeners determines if a given event has wildcard listeners.
	HasWildcardListeners(event any) bool

	// Listen registers an event listener with the dispatcher.
	// events can be: string, []string, Event, []Event, or any other type
	// listener can be: function, class, or any callable
	Listen(events any, listener ...any) error

	// Push pushes an event to be fired at the end of the request.
	Push(event any, payload []Arg)

	// Until fires an event until the first non-null response.
	Until(event any, payload []Arg) any
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

type Listener interface {
	Handle(args ...any) error
}

type EventListener interface {
	Handle(event any, args ...any) error
}

type ShouldQueue interface {
	ShouldQueue() bool
}

type Signature interface {
	Signature() string
}

type QueueListener interface {
	Listener
	ShouldQueue
	Signature
}

type EventQueueListener interface {
	EventListener
	ShouldQueue
	Signature
}
