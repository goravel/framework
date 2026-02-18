package event

import (
	"slices"
	"sync"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
)

type Application struct {
	listeners      map[string][]any
	mu             sync.RWMutex
	wildcards      map[string][]any
	wildcardsCache sync.Map
	events         map[event.Event][]event.Listener
	queue          queue.Queue
}

func NewApplication(queue queue.Queue) *Application {
	return &Application{
		listeners:      make(map[string][]any),
		mu:             sync.RWMutex{},
		wildcards:      make(map[string][]any),
		wildcardsCache: sync.Map{},
		queue:          queue,
	}
}

// GetEvents returns all registered events and their listeners.
func (app *Application) GetEvents() map[event.Event][]event.Listener {
	return app.events
}

// Job returns a new Task for the given event and arguments.
// Deprecated: Use Dispatch instead. Job will be removed in a future version.
// The new Dispatch method fires events immediately and returns listener responses.
//
// Migration example:
//
//	// Old way
//	task := app.Job(event, args)
//	task.Dispatch()
//
//	// New way
//	responses := app.Dispatch(event, args)
func (app *Application) Job(e event.Event, args []event.Arg) event.Task {
	listeners, ok := app.events[e]
	if !ok {
		listeners = make([]event.Listener, 0)
	}

	return NewTask(app.queue, args, e, listeners)
}

// Register registers events and their listeners.
func (app *Application) Register(events map[event.Event][]event.Listener) {
	var (
		jobs     []queue.Job
		jobNames []string
	)

	if app.events == nil {
		app.events = map[event.Event][]event.Listener{}
	}

	for e, listeners := range events {
		app.events[e] = listeners
		for _, listener := range listeners {
			if !slices.Contains(jobNames, listener.Signature()) {
				jobs = append(jobs, listener)
				jobNames = append(jobNames, listener.Signature())
			}
		}
	}

	app.queue.Register(jobs)
}
