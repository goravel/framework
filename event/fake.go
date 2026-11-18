package event

import (
	"slices"
	"sync"

	"github.com/goravel/framework/contracts/event"
)

var _ event.Fake = (*Fake)(nil)

// Fake records the events dispatched while faking. The Event binding is a
// singleton, so the fake belongs to the Application and every Dispatch
// through the facade sees it.
type Fake struct {
	app        *Application
	mu         sync.RWMutex
	enabled    bool
	all        bool
	patterns   []string
	dispatched []event.Dispatched
}

func newFake(app *Application) *Fake {
	return &Fake{app: app}
}

func (r *Fake) AssertDispatched(evt any, assertion ...func(dispatched event.Dispatched) bool) bool {
	return len(r.Dispatched(evt, assertion...)) > 0
}

func (r *Fake) AssertDispatchedTimes(evt any, times int) bool {
	return len(r.Dispatched(evt)) == times
}

func (r *Fake) AssertListening(evt any, target any) bool {
	name, err := getEventName(evt)
	if err != nil {
		return false
	}

	var signature string
	switch l := target.(type) {
	case string:
		signature = l
	case event.Listener:
		if l == nil {
			return false
		}
		signature = l.Signature()
	default:
		return false
	}

	if signature == "" {
		return false
	}

	return slices.ContainsFunc(r.app.prepareListeners(name), func(l *listener) bool {
		return l.signature == signature
	})
}

func (r *Fake) AssertNotDispatched(evt any, assertion ...func(dispatched event.Dispatched) bool) bool {
	return !r.AssertDispatched(evt, assertion...)
}

func (r *Fake) AssertNothingDispatched() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.dispatched) == 0
}

func (r *Fake) Dispatched(evt any, filter ...func(dispatched event.Dispatched) bool) []event.Dispatched {
	pattern, err := getEventName(evt)
	if err != nil {
		return nil
	}

	r.mu.RLock()
	records := slices.Clone(r.dispatched)
	r.mu.RUnlock()

	// Filters run without the lock, so they may call the fake again.
	result := make([]event.Dispatched, 0, len(records))
	for _, record := range records {
		if !matchWildcard(pattern, record.Name) {
			continue
		}

		record.Args = slices.Clone(record.Args)
		if matchesAll(record, filter) {
			result = append(result, record)
		}
	}

	return slices.Clip(result)
}

func (r *Fake) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled = false
	r.all = false
	r.patterns = nil
	r.dispatched = nil
}

// enable starts faking the given events, every event when there is none. An
// event that has no name can never be dispatched, so it fakes nothing.
func (r *Fake) enable(events []any) {
	patterns := make([]string, 0, len(events))
	for _, evt := range events {
		names, err := eventNames(evt)
		if err != nil {
			continue
		}

		patterns = append(patterns, names...)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled = true
	r.all = len(events) == 0
	r.patterns = patterns
}

// record keeps the dispatch when the event is faked, and reports whether it
// did, in which case the listeners must not be called.
func (r *Fake) record(evt any, name string, args []event.Arg) bool {
	if r == nil {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.fakes(name) {
		return false
	}

	r.dispatched = append(r.dispatched, event.Dispatched{
		Event: evt,
		Name:  name,
		Args:  slices.Clone(args),
	})

	return true
}

func (r *Fake) fakes(name string) bool {
	if !r.enabled {
		return false
	}
	if r.all {
		return true
	}

	return slices.ContainsFunc(r.patterns, func(pattern string) bool {
		return matchWildcard(pattern, name)
	})
}

func matchesAll(dispatched event.Dispatched, filters []func(dispatched event.Dispatched) bool) bool {
	for _, filter := range filters {
		if filter != nil && !filter(dispatched) {
			return false
		}
	}

	return true
}

// fakeableTask is the task Job returns. Whether the event is faked is decided
// when the task is dispatched, like Dispatch does.
type fakeableTask struct {
	fake  *Fake
	event event.Event
	name  string
	args  []event.Arg
	task  event.Task
}

func (r *fakeableTask) Dispatch() error {
	if r.fake.record(r.event, r.name, r.args) {
		return nil
	}

	return r.task.Dispatch()
}
