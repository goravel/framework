package event

import (
	"slices"

	"github.com/goravel/framework/contracts/event"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/str"
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

	var errs []error
	for _, l := range app.prepareListeners(name) {
		if err := app.callListener(l, name, evt, payload); err != nil {
			errs = append(errs, err)
		}
	}

	return NewResult(errs)
}

// callListener runs a single listener, a panicking listener fails on its own
// without breaking the rest of the dispatch.
func (app *Application) callListener(l *listener, eventName string, evt any, args []event.Arg) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.EventListenerPanic.Args(eventName, r)
		}
	}()

	var options event.Queue
	if l.queueOptions != nil {
		options = l.queueOptions(argValues(args)...)
	}

	// Listeners registered through the deprecated Register have no in process
	// handler, they always go through the queue facade.
	if l.job != nil && (options.Enable || l.handle == nil) {
		queueArgs := eventArgsToQueueArgs(args)
		if l.withEvent {
			queueArgs = append([]contractsqueue.Arg{{Type: "string", Value: eventName}}, queueArgs...)
		}

		return dispatchToQueue(app.queue, l.job, options, queueArgs)
	}

	// Wildcard listeners are registered on a pattern rather than on an event, so
	// they receive the name of the event that matched.
	if l.wildcard {
		return l.handle(eventName, args)
	}

	return l.handle(evt, args)
}

// prepareListeners collects the listeners of an event, the wildcard matches are
// cached because they are the same for every dispatch of the same event.
func (app *Application) prepareListeners(eventName string) []*listener {
	app.mu.RLock()
	defer app.mu.RUnlock()

	// Cloned so that a listener registering another listener can't mutate the
	// slice being iterated.
	listeners := slices.Clone(app.listeners[eventName])

	if cached, ok := app.wildcardsCache.Load(eventName); ok {
		return append(listeners, cached.([]*listener)...)
	}

	return append(listeners, app.getWildcardListeners(eventName)...)
}

// getWildcardListeners returns the listeners whose pattern matches the event
// name, caching them for the next dispatch. It must be called with the lock held.
func (app *Application) getWildcardListeners(eventName string) []*listener {
	var listeners []*listener
	for pattern, wildcard := range app.wildcards {
		if str.Of(eventName).Is(pattern) {
			listeners = append(listeners, wildcard...)
		}
	}

	app.wildcardsCache.Store(eventName, listeners)

	return listeners
}
