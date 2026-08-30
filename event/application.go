package event

import (
	"slices"
	"sync"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
)

var _ event.Instance = (*Application)(nil)

type Application struct {
	events         map[event.Event][]event.Listener
	listeners      map[string][]*listener
	queue          queue.Queue
	registered     map[string]bool
	wildcards      map[string][]*listener
	wildcardsCache sync.Map
	mu             sync.RWMutex
}

func NewApplication(queue queue.Queue) *Application {
	return &Application{
		listeners:  make(map[string][]*listener),
		queue:      queue,
		registered: make(map[string]bool),
		wildcards:  make(map[string][]*listener),
	}
}

// GetEvents returns a copy of the events registered through Register.
//
// Deprecated: Use Listen instead, GetEvents will be removed in a future version.
func (app *Application) GetEvents() map[event.Event][]event.Listener {
	app.mu.RLock()
	defer app.mu.RUnlock()

	if app.events == nil {
		return nil
	}

	events := make(map[event.Event][]event.Listener, len(app.events))
	for e, listeners := range app.events {
		events[e] = slices.Clone(listeners)
	}

	return events
}

// Job creates a new event task.
//
// Deprecated: Use Dispatch instead, Job will be removed in a future version.
func (app *Application) Job(e event.Event, args []event.Arg) event.Task {
	app.mu.RLock()
	defer app.mu.RUnlock()

	listeners, ok := app.events[e]
	if !ok {
		listeners = make([]event.Listener, 0)
	}

	return NewTask(app.queue, args, e, listeners)
}

// Register registers events and their listeners, the listeners are also
// registered on the Listen flow so that Dispatch can reach them.
//
// Deprecated: Use Listen instead, Register will be removed in a future version.
func (app *Application) Register(events map[event.Event][]event.Listener) {
	var (
		jobs     []queue.Job
		jobNames []string
	)

	app.mu.Lock()

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

		// Mirror the listeners on the Listen flow, replacing any previous
		// registration for the event to keep Register's overwrite semantics.
		if name, err := getEventName(e); err == nil {
			mirrored := make([]*listener, 0, len(listeners))
			for _, l := range listeners {
				mirrored = append(mirrored, newLegacyListener(l))
			}
			app.listeners[name] = mirrored
		}
	}

	// The queue jobs are already deduplicated within this call, remember them so
	// that a later Listen doesn't register the same signature twice.
	for _, name := range jobNames {
		app.registered[name] = true
	}

	app.mu.Unlock()

	app.queue.Register(jobs)
}
