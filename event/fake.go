package event

import (
	"reflect"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/goravel/framework/contracts/event"
)

var _ event.Fake = (*Fake)(nil)

// Fake records the events dispatched while faking. The Event binding is a
// singleton, so the fake belongs to the Application and every Dispatch
// through the facade sees it.
type Fake struct {
	app        *Application
	mu         sync.RWMutex
	enabled    atomic.Bool
	all        bool
	patterns   []string
	dispatched []*dispatched
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
		if isNil(l) {
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

	// The records are never changed once kept, so they can be handed out
	// as they are.
	var records []event.Dispatched
	r.mu.RLock()
	for _, record := range r.dispatched {
		if matchWildcard(pattern, record.name) {
			records = append(records, record)
		}
	}
	r.mu.RUnlock()

	// Filters run without the lock, so they may call the fake again.
	return slices.DeleteFunc(records, func(record event.Dispatched) bool {
		return !matchesAll(record, filter)
	})
}

func (r *Fake) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled.Store(false)
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

	r.enabled.Store(true)
	r.all = len(events) == 0
	r.patterns = patterns
}

// record keeps the dispatch when the event is faked, and reports whether it
// did, in which case the listeners must not be called. It takes no lock while
// nothing is faked, so dispatching outside of tests stays lock free.
func (r *Fake) record(evt any, name string, args []event.Arg) bool {
	if r == nil || !r.enabled.Load() {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.fakes(name) {
		return false
	}

	r.dispatched = append(r.dispatched, &dispatched{
		event: evt,
		name:  name,
		args:  slices.Clone(args),
	})

	return true
}

func (r *Fake) fakes(name string) bool {
	if !r.enabled.Load() {
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

// isNil reports whether the value is nil, a nil pointer held in an interface
// included.
func isNil(value any) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

var _ event.Dispatched = (*dispatched)(nil)

// dispatched is a record of the fake. It is never changed once kept, and Args
// hands out a copy, so a test can not change it.
type dispatched struct {
	event any
	name  string
	args  []event.Arg
}

func (r *dispatched) Arg(index int) any {
	if index < 0 || index >= len(r.args) {
		return nil
	}

	return r.args[index].Value
}

func (r *dispatched) Args() []event.Arg {
	return slices.Clone(r.args)
}

func (r *dispatched) Event() any {
	return r.event
}

func (r *dispatched) HasArg(value any) bool {
	return slices.ContainsFunc(r.args, func(arg event.Arg) bool {
		return reflect.DeepEqual(arg.Value, value)
	})
}

func (r *dispatched) Name() string {
	return r.name
}
