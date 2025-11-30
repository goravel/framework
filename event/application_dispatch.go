package event

import (
	"reflect"
	"time"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/support/str"
)

// Dispatch fires an event and calls all registered listeners for that event.
// It supports both direct event names and wildcard patterns, and returns all listener responses.
// This method is thread-safe and can be called concurrently.
//
// Parameters:
//   - evt: Can be a string, event.Event interface, or any other type
//   - payload: Optional arguments to pass to event listeners
//
// Returns: A slice containing all non-nil responses from event listeners
//
// Example:
//
//	// Dispatch a string event
//	responses := app.Dispatch("user.created", []event.Arg{{Value: user, Type: "User"}})
//
//	// Dispatch a struct event
//	type UserCreated struct { User User }
//	responses := app.Dispatch(&UserCreated{User: user})
//
//	// Dispatch without payload
//	responses := app.Dispatch("app.started")
func (app *Application) Dispatch(evt any, payload ...[]event.Arg) []any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	// as payload is optional, we need to check if it is empty
	var args []event.Arg
	if len(payload) > 0 {
		args = payload[0]
	}

	return app.invokeListeners(evt, args, false)
}

// call invokes a function with the given arguments safely, recovering from panics.
// If a panic occurs, it returns nil. This allows event processing to continue
// even if a single listener panics.
//
// Parameters:
//   - fn: The function to invoke
//   - args: The arguments to pass to the function
//
// Returns: The first return value from the function, or nil if function panics or returns nothing
func (app *Application) call(fn reflect.Value, args []reflect.Value) any {
	defer func() {
		// Recover from panics to prevent one bad listener from breaking
		// the entire event dispatch chain
		_ = recover()
	}()
	if results := fn.Call(args); len(results) > 0 {
		return results[0].Interface()
	}
	return nil
}

// callListener invokes a single listener for the given event.
// It handles three types of listeners:
//   - QueueListener: Queued execution if ShouldQueue() returns true
//   - EventQueueListener: Queued execution for event-aware listeners
//   - Regular listeners: Direct synchronous execution
//
// The method uses reflection to dynamically match listener parameters with
// the event and payload arguments.
//
// Parameters:
//   - listener: The listener to invoke (function, struct with Handle method, or queue listener)
//   - evt: The event being dispatched
//   - args: Optional payload arguments
//
// Returns: The listener's return value, or nil if no value returned
func (app *Application) callListener(listener any, evt any, args []event.Arg) any {
	// Handle queueing
	if ql, ok := listener.(event.QueueListener); ok && ql.ShouldQueue() {
		return app.queueListener(ql, args)
	}
	if eql, ok := listener.(event.EventQueueListener); ok && eql.ShouldQueue() {
		return app.queueEventListener(eql, evt, args) // Direct queueing without wrapper
	}

	// Get the callable (function or method)
	v := reflect.ValueOf(listener)
	if v.Kind() != reflect.Func {
		if method := v.MethodByName("Handle"); method.IsValid() {
			v = method
		} else {
			return nil
		}
	}

	// Call with dynamic args
	return app.callListenerHandle(v, evt, args)
}

// callListenerHandle invokes a listener function with dynamically matched arguments.
// It builds the argument list by matching the event and payload to function parameters.
// Supports variadic parameters and automatic type conversion.
//
// Parameters:
//   - fn: The function to invoke
//   - evt: The event being dispatched
//   - args: Optional payload arguments
//
// Returns: The function's return value, or nil if no value returned
func (app *Application) callListenerHandle(fn reflect.Value, evt any, args []event.Arg) any {
	t := fn.Type()
	numIn := t.NumIn()

	if numIn == 0 {
		return app.call(fn, nil)
	}

	// Build args in one pass
	callArgs := make([]reflect.Value, numIn)
	payloadIdx := 0

	for i := 0; i < numIn; i++ {
		param := t.In(i)

		// First param gets event if it fits
		if i == 0 && app.canUse(evt, param) {
			callArgs[i] = app.convert(evt, param)
			continue
		}

		// Use payload args
		if payloadIdx < len(args) && app.canUse(args[payloadIdx].Value, param) {
			callArgs[i] = app.convert(args[payloadIdx].Value, param)
			payloadIdx++
			continue
		}

		// Variadic slice (consumes all remaining)
		if param.Kind() == reflect.Slice && param.Elem().Kind() == reflect.Interface {
			remaining := make([]any, len(args)-payloadIdx)
			for j, arg := range args[payloadIdx:] {
				remaining[j] = arg.Value
			}
			callArgs[i] = reflect.ValueOf(remaining)
			break
		}

		// Zero value
		callArgs[i] = reflect.Zero(param)
	}

	return app.call(fn, callArgs)
}

// canUse checks if a value can be used for a given parameter type.
// Returns true if the value is nil, assignable to target, or target is string.
func (app *Application) canUse(val any, target reflect.Type) bool {
	if val == nil {
		return true
	}
	return reflect.TypeOf(val).AssignableTo(target) || target.Kind() == reflect.String
}

// convert converts a value to the target reflect type.
// Handles nil values, string conversions, and type assignments.
func (app *Application) convert(val any, target reflect.Type) reflect.Value {
	if val == nil {
		return reflect.Zero(target)
	}
	if target.Kind() == reflect.String {
		return reflect.ValueOf(app.getEventName(val))
	}
	v := reflect.ValueOf(val)
	if v.Type().AssignableTo(target) {
		return v
	}
	return reflect.Zero(target)
}

// getWildcardListeners returns all wildcard listeners that match the given event name.
// It iterates through registered wildcard patterns (e.g., "user.*") and returns listeners
// for patterns that match the event name.
//
// The results are cached in wildcardsCache for performance. The cache is automatically
// cleared when new wildcard listeners are registered.
//
// Parameters:
//   - eventName: The event name to match against wildcard patterns
//
// Returns: A slice of listeners from all matching wildcard patterns
func (app *Application) getWildcardListeners(eventName any) []any {
	var wildcardListeners []any
	eventNameStr := app.getEventName(eventName)

	for event, wildcard := range app.wildcards {
		if str.Of(eventNameStr).Is(app.getEventName(event)) {
			wildcardListeners = append(wildcardListeners, wildcard...)
		}
	}

	// Cache results for performance - thread-safe with sync.Map
	app.wildcardsCache.Store(eventNameStr, wildcardListeners)

	return wildcardListeners
}

// invokeListeners calls all registered listeners for an event.
// It collects responses from all listeners and can optionally halt on first response.
//
// Parameters:
//   - evt: The event being dispatched
//   - payload: Optional payload arguments
//   - halt: If true, stops after first non-nil response
//
// Returns: Slice of all non-nil listener responses
func (app *Application) invokeListeners(evt any, payload []event.Arg, halt bool) []any {
	var responses []any
	listeners := app.prepareListeners(app.getEventName(evt))
	for _, listener := range listeners {
		response := app.callListener(listener, evt, payload)
		if response != nil {
			responses = append(responses, response)
			if halt {
				break
			}
		}
	}

	return responses
}

// prepareListeners gathers all listeners for a given event.
// Combines direct listeners and wildcard listeners, using cache when available.
//
// Parameters:
//   - event: The event name to get listeners for
//
// Returns: Combined slice of all matching listeners
func (app *Application) prepareListeners(event string) []any {
	var allListeners []any

	// Add direct listeners
	if listeners, exists := app.listeners[event]; exists {
		allListeners = append(allListeners, listeners...)
	}

	// Add wildcard listeners (use cache if available)
	if cached, exists := app.wildcardsCache.Load(event); exists {
		if listeners, ok := cached.([]any); ok {
			allListeners = append(allListeners, listeners...)
		}
	} else {
		// event is already a string, no need to call getEventName again
		wildcardListeners := app.getWildcardListeners(event)
		allListeners = append(allListeners, wildcardListeners...)
	}

	return allListeners
}

// queueEventListener creates a dynamic queue job for an event-aware listener.
// Wraps the listener in a closure that preserves the event context.
//
// Parameters:
//   - listener: The event queue listener to wrap
//   - evt: The event being dispatched
//   - args: Optional payload arguments
//
// Returns: nil (queuing is asynchronous)
func (app *Application) queueEventListener(listener event.EventQueueListener, evt any, args []event.Arg) any {
	// Capture event in local scope to avoid closure variable capture bug
	eventCopy := evt

	// Create a dynamic queue job using closure
	job := &dynamicQueueJob{
		signature: listener.Signature(),
		handler: func(queueArgs ...any) error {
			// Convert queue args back to event args format
			eventArgs := make([]any, len(queueArgs))
			copy(eventArgs, queueArgs)
			return listener.Handle(eventCopy, eventArgs...)
		},
		shouldQueue: listener.ShouldQueue(),
		queueConfig: listener, // For queue configuration methods
	}

	return app.queueListener(job, args)
}

// queueListener dispatches a listener to the queue system.
// Configures the queue task based on listener configuration methods.
//
// Parameters:
//   - listener: The queue listener to dispatch
//   - payload: Optional payload arguments
//
// Returns: nil (queuing is asynchronous)
func (app *Application) queueListener(listener event.QueueListener, payload []event.Arg) any {
	task := app.queue.Job(listener, eventArgsToQueueArgs(payload))

	// Configure queue task using interface assertions instead of reflection
	// This is significantly faster than MethodByName approach

	// Configure connection if listener implements ViaConnection
	if configured, ok := listener.(interface{ ViaConnection() string }); ok {
		task.OnConnection(configured.ViaConnection())
	}

	// Configure queue if listener implements ViaQueue
	if configured, ok := listener.(interface{ ViaQueue() string }); ok {
		task.OnQueue(configured.ViaQueue())
	}

	// Configure delay if listener implements WithDelay
	if configured, ok := listener.(interface{ WithDelay() int64 }); ok {
		task.Delay(time.Now().Add(time.Duration(configured.WithDelay()) * time.Second))
	}

	// Dispatch task - errors are intentionally not returned as per framework pattern
	// Queued jobs handle their own error recovery and reporting
	_ = task.Dispatch()

	return nil
}
