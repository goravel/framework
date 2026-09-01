package event

import (
	"slices"

	"github.com/goravel/framework/contracts/event"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

// dispatchMode carries the differences between the two entry points into the
// shared pipeline. Dispatch runs every listener and reports all the failures,
// while the deprecated Task keeps its fail fast behaviour, so that a listener
// error still prevents the jobs behind it from being queued.
type dispatchMode struct {
	requireListeners bool
	stopOnError      bool
}

var (
	dispatchModeDispatch = dispatchMode{}
	dispatchModeTask     = dispatchMode{requireListeners: true, stopOnError: true}
)

// Dispatch fires an event and calls all the listeners registered for it, both
// the ones registered on the event itself and the ones registered on a matching
// wildcard pattern. Every listener runs, the returned result carries the errors
// of the ones that failed.
func (app *Application) Dispatch(evt any, args ...[]event.Arg) event.Result {
	name, err := getEventName(evt)
	if err != nil {
		return NewResult([]error{err})
	}

	// The payload is optional.
	var payload []event.Arg
	if len(args) > 0 {
		payload = args[0]
	}

	return NewResult(dispatch(app.queue, evt, name, app.prepareListeners(name), payload, dispatchModeDispatch))
}

// dispatch is the single pipeline behind both Dispatch and the deprecated Task.
// The listeners are snapshotted by the caller, so a listener registering another
// listener only affects later dispatches.
func dispatch(queue contractsqueue.Queue, evt any, eventName string, listeners []*listener, args []event.Arg, mode dispatchMode) []error {
	if len(listeners) == 0 {
		if mode.requireListeners {
			return []error{errors.EventListenerNotBind.Args(evt)}
		}

		// Without a listener there is nothing to prepare the payload for, so the
		// event's own Handle is not run either.
		return nil
	}

	// An event.Event prepares its own payload before the listeners see it, the
	// behaviour the deprecated Task has always had.
	if e, ok := evt.(event.Event); ok {
		handled, err := e.Handle(args)
		if err != nil {
			return []error{err}
		}

		args = handled
	}

	var errs []error
	for _, l := range listeners {
		if err := callListener(queue, l, eventName, evt, args); err != nil {
			errs = append(errs, err)
			if mode.stopOnError {
				break
			}
		}
	}

	return errs
}

// callListener runs a single listener, a panicking listener fails on its own
// without breaking the rest of the dispatch.
func callListener(queue contractsqueue.Queue, l *listener, eventName string, evt any, args []event.Arg) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.EventListenerPanic.Args(eventName, r)
		}
	}()

	var options event.Queue
	if l.queueOptions != nil {
		options = l.queueOptions(argValues(args)...)
	}

	if l.job != nil && options.Enable {
		return dispatchToQueue(queue, l.job, options, queueArgs(eventName, args))
	}

	// Wildcard listeners are registered on a pattern rather than on an event, so
	// they receive the name of the event that matched.
	if l.wildcard {
		return l.handle(eventName, eventName, args)
	}

	return l.handle(eventName, evt, args)
}

// prepareListeners snapshots the listeners of an event, the exact matches first
// and then the wildcard ones in registration order.
func (app *Application) prepareListeners(eventName string) []*listener {
	app.mu.RLock()
	defer app.mu.RUnlock()

	listeners := slices.Clone(app.listeners[eventName])
	for _, wildcard := range app.wildcards {
		if matchWildcard(wildcard.pattern, eventName) {
			listeners = append(listeners, wildcard.listeners...)
		}
	}

	return listeners
}
