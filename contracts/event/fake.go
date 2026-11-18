package event

import (
	"reflect"
	"slices"
)

// Fake records the events dispatched while faking instead of calling their
// listeners.
//
// The event parameter of the methods below is the event to look for: its
// name, an instance of it (a nil pointer such as (*OrderShipped)(nil) is
// enough), or a wildcard pattern such as "user.*" that matches every event
// sharing the prefix.
type Fake interface {
	// AssertDispatched reports whether the event was dispatched, and when
	// assertions are given, whether one of its dispatches matches all of them.
	AssertDispatched(event any, assertion ...func(dispatched Dispatched) bool) bool
	// AssertDispatchedTimes reports whether the event was dispatched exactly
	// times times.
	AssertDispatchedTimes(event any, times int) bool
	// AssertListening reports whether the listener is registered for the
	// event, directly or through a matching wildcard pattern. The listener is
	// a Listener or its signature; closures have no signature, so they can
	// not be asserted.
	AssertListening(event any, listener any) bool
	// AssertNotDispatched reports whether the event was not dispatched, and
	// when assertions are given, whether none of its dispatches matches all
	// of them.
	AssertNotDispatched(event any, assertion ...func(dispatched Dispatched) bool) bool
	// AssertNothingDispatched reports whether no faked event was dispatched.
	AssertNothingDispatched() bool
	// Dispatched returns the dispatches of the event, only those matching
	// every given filter.
	Dispatched(event any, filter ...func(dispatched Dispatched) bool) []Dispatched
	// Reset stops faking: the listeners are called again and the records are
	// cleared.
	Reset()
}

// Dispatched is an event recorded by the fake.
type Dispatched struct {
	// Event is the value passed to Dispatch: the event itself, or its name
	// for events dispatched by name.
	Event any
	// Name is the name the event was dispatched under, the one a wildcard
	// listener would receive.
	Name string
	// Args is the payload passed to Dispatch. The listeners are not called,
	// so Event.Handle does not prepare it either.
	Args []Arg
}

// Arg returns the value of the argument at index, or nil when there is none.
func (r Dispatched) Arg(index int) any {
	if index < 0 || index >= len(r.Args) {
		return nil
	}

	return r.Args[index].Value
}

// HasArg reports whether one of the arguments has the value, compared with
// reflect.DeepEqual.
func (r Dispatched) HasArg(value any) bool {
	return slices.ContainsFunc(r.Args, func(arg Arg) bool {
		return reflect.DeepEqual(arg.Value, value)
	})
}
