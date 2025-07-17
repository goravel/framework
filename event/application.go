// Package event provides a comprehensive event dispatching system for the Goravel framework.
// It supports synchronous and asynchronous event handling, wildcard event listening,
// and queued event processing. The event system is thread-safe and optimized for performance.
//
// Example usage:
//
//	// Create an event dispatcher
//	app := event.NewApplication(queue)
//
//	// Listen to specific events
//	app.Listen("user.created", func(args ...any) error {
//		// Handle user creation
//		return nil
//	})
//
//	// Listen with wildcard patterns
//	app.Listen("user.*", func(args ...any) error {
//		// Handle all user events
//		return nil
//	})
//
//	// Dispatch events
//	app.Dispatch("user.created", []event.Arg{{Value: user, Type: "User"}})
//
//	// Push events for later processing
//	app.Push("user.created", []event.Arg{{Value: user, Type: "User"}})
//	app.Flush("user.created") // Process all pushed events
package event

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/str"
)

// Application is the main event dispatcher that handles event registration, dispatch, and queuing.
// It provides thread-safe operations with read-write mutexes and maintains separate storage for
// direct listeners, wildcard listeners, and deferred events.
type Application struct {
	listeners      map[string][]any       // Direct event listeners mapped by event name
	wildcards      map[string][]any       // Wildcard pattern listeners (e.g., "user.*")
	wildcardsCache map[string][]any       // Cache for wildcard matching performance optimization
	pushed         map[string][]event.Arg // Deferred events for batch processing
	mu             sync.RWMutex           // Thread-safe operations mutex
	queue          queue.Queue            // Queue for asynchronous processing
}

// NewApplication creates a new event dispatcher instance with the provided queue system.
// The queue is used for asynchronous event processing when listeners implement ShouldQueue.
//
// Example:
//
//	queue := queue.NewApplication()
//	eventDispatcher := event.NewApplication(queue)
func NewApplication(queue queue.Queue) *Application {
	return &Application{
		queue:          queue,
		listeners:      make(map[string][]any),
		wildcards:      make(map[string][]any),
		wildcardsCache: make(map[string][]any),
		pushed:         make(map[string][]event.Arg),
	}
}

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

// Flush processes all pushed events for a specific event type and then clears them.
// This is useful for batch processing events that were deferred using Push().
// The method is thread-safe and dispatches events outside the lock to prevent deadlocks.
//
// Example:
//
//	// Push multiple events
//	app.Push("user.created", []event.Arg{{Value: user1, Type: "User"}})
//	app.Push("user.created", []event.Arg{{Value: user2, Type: "User"}})
//
//	// Process all pushed user.created events at once
//	app.Flush("user.created")
func (app *Application) Flush(evt any) {
	app.mu.Lock()
	eventName := app.getEventName(evt)

	payloads, exists := app.pushed[eventName]
	if exists {
		delete(app.pushed, eventName)
	}
	app.mu.Unlock()

	// Dispatch outside of lock to avoid deadlock
	if exists {
		app.Dispatch(eventName, payloads)
	}
}

// Forget removes all listeners for a specific event or wildcard pattern.
// This method also clears related cache entries to maintain performance.
// It distinguishes between direct listeners and wildcard listeners automatically.
//
// Example:
//
//	// Remove all listeners for a specific event
//	app.Forget("user.created")
//
//	// Remove all listeners for a wildcard pattern
//	app.Forget("user.*")
func (app *Application) Forget(evt any) {
	app.mu.Lock()
	defer app.mu.Unlock()

	eventName := app.getEventName(evt)

	if strings.Contains(eventName, "*") {
		delete(app.wildcards, eventName)
	} else {
		delete(app.listeners, eventName)
	}

	// Clear related cache entries for performance
	for key := range app.wildcardsCache {
		if str.Of(app.getEventName(key)).Is(eventName) {
			delete(app.wildcardsCache, key)
		}
	}
}

// ForgetPushed clears all pushed events without processing them.
// This is useful when you want to discard deferred events without executing them.
//
// Example:
//
//	// Push some events
//	app.Push("user.created", []event.Arg{{Value: user, Type: "User"}})
//	app.Push("order.placed", []event.Arg{{Value: order, Type: "Order"}})
//
//	// Discard all pushed events
//	app.ForgetPushed()
func (app *Application) ForgetPushed() {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.pushed = make(map[string][]event.Arg)
}

// GetListeners returns all listeners registered for a specific event.
// This includes both direct listeners and wildcard listeners that match the event.
// The method uses caching to optimize wildcard matching performance.
//
// Example:
//
//	listeners := app.GetListeners("user.created")
//	fmt.Printf("Found %d listeners for user.created\n", len(listeners))
func (app *Application) GetListeners(evt any) []any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	return app.prepareListeners(app.getEventName(evt))
}

// HasListeners checks if there are any listeners registered for a specific event.
// It performs an optimized check by first looking at direct listeners (fastest),
// then wildcard listeners, and finally wildcard pattern matches.
//
// Example:
//
//	if app.HasListeners("user.created") {
//		fmt.Println("User creation listeners are registered")
//	}
func (app *Application) HasListeners(evt any) bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

	eventName := app.getEventName(evt)

	// Check direct listeners first (fastest)
	if listeners, exists := app.listeners[eventName]; exists && len(listeners) > 0 {
		return true
	}

	// Check wildcard listeners
	if listeners, exists := app.wildcards[eventName]; exists && len(listeners) > 0 {
		return true
	}

	// Check wildcard pattern matches
	return app.HasWildcardListeners(eventName)
}

// HasWildcardListeners checks if there are any wildcard listeners that match the given event.
// It iterates through all registered wildcard patterns to find matches.
//
// Example:
//
//	app.Listen("user.*", listener)
//	hasListeners := app.HasWildcardListeners("user.created") // Returns true
func (app *Application) HasWildcardListeners(eventName any) bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

	eventNameStr := app.getEventName(eventName)
	for event := range app.wildcards {
		if str.Of(eventNameStr).Is(app.getEventName(event)) {
			return true
		}
	}

	return false
}

// Listen registers event listeners for one or more events.
// It supports various event types and listener formats:
//   - Events: string, []string, event.Event, []event.Event, or any custom type
//   - Listeners: functions, structs implementing listener interfaces, or closures
//
// The method automatically handles queue registration for listeners that implement ShouldQueue.
// Wildcard patterns (containing "*") are supported for flexible event matching.
//
// Examples:
//
//	// Listen to a single string event
//	app.Listen("user.created", func(args ...any) error {
//		return nil
//	})
//
//	// Listen to multiple events
//	app.Listen([]string{"user.created", "user.updated"}, listener)
//
//	// Listen to wildcard events
//	app.Listen("user.*", func(args ...any) error {
//		return nil
//	})
//
//	// Listen with event structs
//	app.Listen(UserCreatedEvent{}, &UserCreatedListener{})
func (app *Application) Listen(events any, listener ...any) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	if len(listener) == 0 {
		// it may be a closure
		if t := reflect.TypeOf(events); t.Kind() == reflect.Func {
			listenerValue := reflect.ValueOf(events)
			listenerType := listenerValue.Type()
			// check if listenerType is func(event.Event) error
			if listenerType.NumIn() == 1 && app.isEventType(listenerType.In(0)) {
				// get first parameter and pass to setupEvents
				ptr := reflect.New(listenerType.In(0).Elem()).Interface()
				app.setupEvents(ptr, events)
				return nil
			}
		}

		return errors.New("listener is required")
	}

	switch e := events.(type) {
	case string:
		app.setupEvents(e, listener[0])
	case []string:
		for _, eventName := range e {
			app.setupEvents(eventName, listener[0])
		}
	case []event.Event:
		for _, evt := range e {
			app.setupEvents(evt, listener[0])
		}
	case event.Event:
		app.setupEvents(e, listener[0])
	default:
		if eventName := app.getEventName(events); eventName != "" {
			app.setupEvents(eventName, listener[0])
		} else {
			return errors.New("invalid event type")
		}
	}

	return nil
}

// Push adds an event to the deferred events queue without immediately dispatching it.
// This is useful for batch processing events at a later time using Flush().
// Events are stored by their name and can accumulate multiple payloads.
//
// Example:
//
//	// Push events for later processing
//	app.Push("user.created", []event.Arg{{Value: user1, Type: "User"}})
//	app.Push("user.created", []event.Arg{{Value: user2, Type: "User"}})
//
//	// Process all pushed events later
//	app.Flush("user.created")
func (app *Application) Push(evt any, payload []event.Arg) {
	app.mu.Lock()
	defer app.mu.Unlock()

	eventName := app.getEventName(evt)

	app.pushed[eventName] = append(app.pushed[eventName], payload...)
}

// Until fires an event and returns the first non-null response from listeners.
// This is useful for validation or filtering scenarios where you want to stop
// at the first listener that returns a meaningful result.
//
// Example:
//
//	// Register validation listeners
//	app.Listen("user.validate", func(args ...any) any {
//		if user.Email == "" {
//			return "Email is required"
//		}
//		return nil // Continue to next listener
//	})
//
//	// Get first validation error
//	if err := app.Until("user.validate", []event.Arg{{Value: user, Type: "User"}}); err != nil {
//		fmt.Println("Validation failed:", err)
//	}
func (app *Application) Until(eventName any, payload []event.Arg) any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	responses := app.invokeListeners(eventName, payload, true)
	if len(responses) > 0 {
		return responses[0]
	}
	return nil
}

func (app *Application) callListener(listener any, evt any, args []event.Arg) any {
	// Convert event.Arg to []any for most listeners
	payload := make([]any, len(args))
	for i, arg := range args {
		payload[i] = arg.Value
	}

	switch l := listener.(type) {
	case event.QueueListener:
		if l.ShouldQueue() {
			return app.queueListener(l, evt, args)
		}
		return l.Handle(payload...)
	case event.EventQueueListener:
		if l.ShouldQueue() {
			return app.queueListener(eventToQueueListener(l, evt), evt, args)
		}
		return l.Handle(evt, payload...)
	case event.Listener:
		return l.Handle(payload...)
	case event.EventListener:
		return l.Handle(evt, payload...)
	case func(string, ...any) error:
		return l(app.getEventName(evt), payload...)
	case func(any) error:
		return l(evt)
	case func(any, ...any) error:
		return l(evt, payload...)
	case func(...any) error:
		return l(payload...)
	default:
		return app.callReflectListener(listener, evt, payload)
	}
}

func (app *Application) isEventType(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Implements(reflect.TypeOf((*event.Event)(nil)).Elem())
}

func (app *Application) callReflectListener(listener any, evt any, payload []any) any {
	listenerValue := reflect.ValueOf(listener)
	listenerType := listenerValue.Type()

	if listenerType.Kind() != reflect.Func {
		return nil
	}

	numIn := listenerType.NumIn()

	args := make([]reflect.Value, 0, numIn)

	if numIn > 0 && app.isEventType(listenerType.In(0)) {
		args = append(args, reflect.ValueOf(evt))
	}

	// Fill arguments from payload
	for i := 0; i < numIn && i < len(payload); i++ {
		args = append(args, reflect.ValueOf(payload[i]))
	}

	// Fill remaining arguments with zero values
	for len(args) < numIn {
		args = append(args, reflect.Zero(listenerType.In(len(args))))
	}

	for len(args) > numIn {
		// Remove extra arguments if args is longer than numIn
		args = args[:numIn]
	}

	results := listenerValue.Call(args)
	if len(results) > 0 {
		return results[0].Interface()
	}

	return nil
}

func (app *Application) getEventName(evt any) string {
	switch e := evt.(type) {
	case string:
		return e
	case event.Event:
		// Check if event has custom signature method
		if signature, ok := e.(event.Signature); ok {
			return signature.Signature()
		}

		// Use reflection to get struct type name
		eventType := reflect.TypeOf(e)
		if eventType.Kind() == reflect.Ptr {
			eventType = eventType.Elem()
		}

		return eventType.Name()
	default:
		// Use reflection for any other type
		eventType := reflect.TypeOf(evt)
		if eventType != nil {
			if eventType.Kind() == reflect.Ptr {
				eventType = eventType.Elem()
			}
			// For unnamed structs, use the full type string
			name := eventType.Name()
			if name == "" {
				name = eventType.String()
			}
			return name
		}
		return ""
	}
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

func (app *Application) queueListener(listener event.QueueListener, evt any, payload []event.Arg) any {
	task := app.queue.Job(listener, eventArgsToQueueArgs(payload))

	// Use reflection to call optional configuration methods
	ref := reflect.ValueOf(listener)

	// Configure connection if ViaConnection method exists
	if method := ref.MethodByName("ViaConnection"); method.IsValid() {
		result := method.Call(nil)
		if len(result) > 0 {
			viaConnection := result[0].String()
			task.OnConnection(viaConnection)
		}
	}

	// Configure queue if ViaQueue method exists
	if method := ref.MethodByName("ViaQueue"); method.IsValid() {
		result := method.Call(nil)
		if len(result) > 0 {
			viaQueue := result[0].String()
			task.OnQueue(viaQueue)
		}
	}

	// Configure delay if WithDelay method exists
	if method := ref.MethodByName("WithDelay"); method.IsValid() {
		result := method.Call(nil)
		if len(result) > 0 {
			delaySeconds := result[0].Int()
			task.Delay(time.Now().Add(time.Duration(delaySeconds) * time.Second))
		}
	}

	_ = task.Dispatch()

	return nil
}

func (app *Application) setupEvents(e any, listener any) {
	eventName := app.getEventName(e)

	// Register queued listeners with the queue system
	if l, ok := listener.(event.EventQueueListener); ok {
		app.queue.Register([]queue.Job{eventToQueueListener(l, eventName)})
	} else if l, ok := listener.(event.QueueListener); ok {
		app.queue.Register([]queue.Job{l})
	}

	// Register the listener based on event type
	if strings.Contains(eventName, "*") {
		app.setupWildcardListen(eventName, listener)
	} else {
		app.listeners[eventName] = append(app.listeners[eventName], listener)
	}
}

func (app *Application) setupWildcardListen(eventName string, listener any) {
	app.wildcards[eventName] = append(app.wildcards[eventName], listener)
	// Clear cache to ensure fresh wildcard matching
	app.wildcardsCache = make(map[string][]any)
}

func eventArgsToQueueArgs(args []event.Arg) []queue.Arg {
	var queueArgs []queue.Arg
	for _, arg := range args {
		queueArgs = append(queueArgs, queue.Arg{
			Type:  arg.Type,
			Value: arg.Value,
		})
	}

	return queueArgs
}
