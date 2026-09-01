package event

import (
	"slices"
	"sync"

	"github.com/goravel/framework/contracts/event"
	contractsqueue "github.com/goravel/framework/contracts/queue"
)

var _ event.Instance = (*Application)(nil)

type Application struct {
	events     map[event.Event][]event.Listener
	listeners  map[string][]*listener
	queue      contractsqueue.Queue
	registered map[string]any
	wildcards  []wildcardEntry
	mu         sync.RWMutex
}

func NewApplication(queue contractsqueue.Queue) *Application {
	return &Application{
		listeners:  make(map[string][]*listener),
		queue:      queue,
		registered: make(map[string]any),
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

	return NewTask(app.queue, args, e, slices.Clone(listeners))
}

// Register registers events and their listeners, the listeners are also
// registered on the Listen flow so that Dispatch can reach them.
//
// Deprecated: Use Listen instead, Register will be removed in a future version.
func (app *Application) Register(events map[event.Event][]event.Listener) {
	type registration struct {
		name      string
		named     bool
		listeners []*listener
	}

	var (
		jobs     []contractsqueue.Job
		jobNames []string
		regs     = make([]registration, 0, len(events))
		cloned   = make(map[event.Event][]event.Listener, len(events))
	)

	// Resolving a signature calls into the listener. That must not happen while
	// the registry is locked: a panicking Signature would leave the mutex held
	// forever, and one that registers another listener would deadlock.
	for e, listeners := range events {
		listeners = slices.Clone(listeners)
		cloned[e] = listeners

		name, err := getEventName(e)
		reg := registration{name: name, named: err == nil}

		for _, l := range listeners {
			// The job resolves the signature, asking the listener again would call
			// into it twice for every registration.
			job := newQueueJob(l)
			signature := job.Signature()
			if !slices.Contains(jobNames, signature) {
				jobs = append(jobs, job)
				jobNames = append(jobNames, signature)
			}

			if reg.named {
				reg.listeners = append(reg.listeners, newLegacyListener(l, signature))
			}
		}

		regs = append(regs, reg)
	}

	func() {
		app.mu.Lock()
		defer app.mu.Unlock()

		if app.events == nil {
			app.events = map[event.Event][]event.Listener{}
		}

		for e, listeners := range cloned {
			app.events[e] = listeners
		}

		// Register overwrites the listeners of an event, so only the ones it
		// registered before are dropped, the ones added by Listen are kept.
		for _, reg := range regs {
			if !reg.named {
				continue
			}

			mirrored := slices.DeleteFunc(slices.Clone(app.listeners[reg.name]), func(l *listener) bool {
				return l.legacy
			})
			app.listeners[reg.name] = append(mirrored, reg.listeners...)
		}

		// The queue jobs are already deduplicated within this call, remember them
		// so that a later Listen doesn't register the same signature twice.
		for i, name := range jobNames {
			app.registered[name] = listenerIdentity(jobs[i].(*queueJob).listener)
		}
	}()

	app.queue.Register(jobs)
}
