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

// call - Safe function call with recovery
func (app *Application) call(fn reflect.Value, args []reflect.Value) any {
	defer func() { _ = recover() }()
	if results := fn.Call(args); len(results) > 0 {
		return results[0].Interface()
	}
	return nil
}

func (app *Application) callListener(listener any, evt any, args []event.Arg) any {
	// Handle queueing
	if ql, ok := listener.(event.QueueListener); ok && ql.ShouldQueue() {
		return app.queueListener(ql, evt, args)
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

// canUse - Simple type compatibility check
func (app *Application) canUse(val any, target reflect.Type) bool {
	if val == nil {
		return true
	}
	return reflect.TypeOf(val).AssignableTo(target) || target.Kind() == reflect.String
}

// convert - Convert value to target type
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

func (app *Application) getWildcardListeners(eventName any) []any {
	var wildcardListeners []any
	eventNameStr := app.getEventName(eventName)

	for event, wildcard := range app.wildcards {
		if str.Of(eventNameStr).Is(app.getEventName(event)) {
			wildcardListeners = append(wildcardListeners, wildcard...)
		}
	}

	// Cache results for performance
	app.wildcardsCache[eventNameStr] = wildcardListeners

	return wildcardListeners
}

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

func (app *Application) prepareListeners(event string) []any {
	var allListeners []any

	// Add direct listeners
	if listeners, exists := app.listeners[event]; exists {
		allListeners = append(allListeners, listeners...)
	}

	// Add wildcard listeners (use cache if available)
	if listeners, exists := app.wildcardsCache[event]; exists {
		allListeners = append(allListeners, listeners...)
	} else {
		wildcardListeners := app.getWildcardListeners(app.getEventName(event))
		allListeners = append(allListeners, wildcardListeners...)
	}

	return allListeners
}

func (app *Application) queueEventListener(listener event.EventQueueListener, evt any, args []event.Arg) any {
	// Create a dynamic queue job using closure
	job := &dynamicQueueJob{
		signature: listener.Signature(),
		handler: func(queueArgs ...any) error {
			// Convert queue args back to event args format
			eventArgs := make([]any, len(queueArgs))
			copy(eventArgs, queueArgs)
			return listener.Handle(evt, eventArgs...)
		},
		shouldQueue: listener.ShouldQueue(),
		queueConfig: listener, // For queue configuration methods
	}

	return app.queueListener(job, evt, args)
}

func (app *Application) queueListener(listener event.QueueListener, evt any, payload []event.Arg) any {
	task := app.queue.Job(listener, eventArgsToQueueArgs(payload))

	// Use reflection to call optional configuration methods
	ref := reflect.ValueOf(listener)

	// Configure connection if ViaConnection method exists
	if method := ref.MethodByName("ViaConnection"); method.IsValid() {
		result := method.Call(nil)
		if len(result) > 0 {
			task.OnConnection(result[0].String())
		}
	}

	// Configure queue if ViaQueue method exists
	if method := ref.MethodByName("ViaQueue"); method.IsValid() {
		result := method.Call(nil)
		if len(result) > 0 {
			task.OnQueue(result[0].String())
		}
	}

	// Configure delay if WithDelay method exists
	if method := ref.MethodByName("WithDelay"); method.IsValid() {
		result := method.Call(nil)
		if len(result) > 0 {
			task.Delay(time.Now().Add(time.Duration(result[0].Int()) * time.Second))
		}
	}

	_ = task.Dispatch()

	return nil
}
