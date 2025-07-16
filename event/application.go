package event

import (
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/str"
)

// Application implements the event dispatcher with Laravel-style functionality.
//
// This dispatcher provides comprehensive event handling capabilities including:
// - Multi-format event listening (string, struct, slice)
// - Wildcard pattern matching for flexible event handling
// - Synchronous and asynchronous (queued) event processing
// - Push/flush mechanisms for deferred event processing
// - Thread-safe operations with optimized performance
//
// Example usage:
//   queue := getQueueInstance()
//   app := NewApplication(queue)
//
//   // Listen to string events
//   app.Listen("user.created", func(user User) {
//       fmt.Printf("User %s created\n", user.Name)
//   })
//
//   // Listen with wildcard patterns
//   app.Listen("user.*", func(event any, data ...any) {
//       fmt.Printf("User event: %s\n", event)
//   })
//
//   // Dispatch events
//   app.Dispatch("user.created", []event.Arg{{Value: user, Type: "User"}})
type Application struct {
	listeners      map[any][]any       // Direct event listeners
	wildcards      map[any][]any       // Wildcard pattern listeners
	wildcardsCache map[any][]any       // Cache for wildcard matching performance
	pushed         map[any][]event.Arg // Deferred events for batch processing
	mu             sync.RWMutex        // Thread-safe operations
	queue          queue.Queue         // Queue for asynchronous processing
}

// NewApplication creates a new event dispatcher instance.
//
// The queue parameter enables asynchronous event processing for listeners
// that implement QueueListener interface. If nil, all events are processed
// synchronously.
//
// Example:
//   queue := getQueueInstance()
//   dispatcher := NewApplication(queue)
//
// Parameters:
//   - queue: Queue instance for async processing (can be nil)
//
// Returns:
//   - *Application: New dispatcher instance
func NewApplication(queue queue.Queue) *Application {
	return &Application{
		queue:          queue,
		listeners:      make(map[any][]any),
		wildcards:      make(map[any][]any),
		wildcardsCache: make(map[any][]any),
		pushed:         make(map[any][]event.Arg),
	}
}

// =============================================================================
// PUBLIC METHODS (Laravel-style dispatcher interface, alphabetically ordered)
// =============================================================================

// Dispatch fires an event and calls all registered listeners.
//
// This method processes events by finding all matching listeners (direct and
// wildcard) and calling them with the provided payload. Responses are collected
// and returned as a slice.
//
// Example usage:
//   // Dispatch simple event
//   args := []event.Arg{{Value: user, Type: "User"}}
//   responses := app.Dispatch("user.created", args)
//
//   // Dispatch struct event
//   event := UserCreatedEvent{User: user}
//   responses := app.Dispatch(event, args)
//
//   // Process responses
//   for _, response := range responses {
//       if result, ok := response.(ValidationResult); ok {
//           fmt.Printf("Validation: %v\n", result.Valid)
//       }
//   }
//
// Parameters:
//   - event: Event to dispatch (string, struct, or any type)
//   - payload: Data to pass to listeners as event.Arg slice
//
// Returns:
//   - []any: Slice of responses from all listeners
func (app *Application) Dispatch(event any, payload []event.Arg) []any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	return app.invokeListeners(event, payload, false)
}

// Flush processes all deferred events for a specific event name.
//
// Events can be deferred using Push() and processed later in batches using
// Flush(). This is useful for collecting related events and processing them
// together for better performance.
//
// Example usage:
//   // Defer multiple events
//   app.Push("user.created", []event.Arg{{Value: user1, Type: "User"}})
//   app.Push("user.created", []event.Arg{{Value: user2, Type: "User"}})
//   app.Push("user.created", []event.Arg{{Value: user3, Type: "User"}})
//
//   // Process all deferred events at once
//   app.Flush("user.created")
//
// Parameters:
//   - eventName: Name of the event to flush (string)
func (app *Application) Flush(eventName string) {
	app.mu.Lock()
	defer app.mu.Unlock()

	if payloads, exists := app.pushed[eventName]; exists {
		delete(app.pushed, eventName)
		app.mu.Unlock()

		app.Dispatch(eventName, payloads)

		app.mu.Lock()
	}
}

// Forget removes all listeners for a specific event.
//
// This method cleans up listeners and cache entries for the specified event.
// It handles both direct listeners and wildcard patterns, ensuring complete
// cleanup to prevent memory leaks.
//
// Example usage:
//   // Remove all listeners for specific event
//   app.Forget("user.created")
//
//   // Remove wildcard listeners
//   app.Forget("user.*")
//
//   // Verify removal
//   if !app.HasListeners("user.created") {
//       fmt.Println("All listeners removed")
//   }
//
// Parameters:
//   - eventName: Name of the event to forget (string)
func (app *Application) Forget(eventName string) {
	app.mu.Lock()
	defer app.mu.Unlock()

	if strings.Contains(eventName, "*") {
		delete(app.wildcards, eventName)
	} else {
		delete(app.listeners, eventName)
	}

	// Clear related cache entries for performance
	for key, _ := range app.wildcardsCache {
		if str.Of(app.getEventName(key)).Is(eventName) {
			delete(app.wildcardsCache, key)
		}
	}
}

// ForgetPushed clears all deferred events without processing them.
//
// This method removes all events that were pushed for later processing,
// effectively canceling any deferred event processing.
//
// Example usage:
//   // Push some events
//   app.Push("user.created", []event.Arg{{Value: user1, Type: "User"}})
//   app.Push("order.processed", []event.Arg{{Value: order1, Type: "Order"}})
//
//   // Cancel all deferred events
//   app.ForgetPushed()
//
//   // No events will be processed when Flush() is called
func (app *Application) ForgetPushed() {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.pushed = make(map[any][]event.Arg)
}

// GetListeners retrieves all listeners that would handle a specific event.
//
// This method returns all listeners (direct and wildcard matches) that would
// be called if the specified event were dispatched. Useful for debugging
// and introspection.
//
// Example usage:
//   listeners := app.GetListeners("user.created")
//   fmt.Printf("Found %d listeners\n", len(listeners))
//
//   // Inspect listener types
//   for i, listener := range listeners {
//       fmt.Printf("Listener %d: %T\n", i, listener)
//   }
//
// Parameters:
//   - eventName: Event to get listeners for (any type)
//
// Returns:
//   - []any: Slice of all matching listeners
func (app *Application) GetListeners(eventName any) []any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	return app.prepareListeners(eventName)
}

// HasListeners checks if any listeners are registered for an event.
//
// This method efficiently checks for both direct listeners and wildcard
// patterns that would match the specified event. Returns true if any
// listeners would be called.
//
// Example usage:
//   // Check before dispatching to avoid unnecessary work
//   if app.HasListeners("user.created") {
//       app.Dispatch("user.created", payload)
//   }
//
//   // Check for wildcard matches
//   if app.HasListeners("user.updated") {
//       fmt.Println("User update listeners are registered")
//   }
//
// Parameters:
//   - eventName: Event to check for listeners (any type)
//
// Returns:
//   - bool: true if listeners exist, false otherwise
func (app *Application) HasListeners(eventName any) bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

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

// HasWildcardListeners checks if wildcard patterns match an event.
//
// This method specifically checks for wildcard pattern listeners that would
// match the specified event name. More specific than HasListeners() for
// wildcard-only checking.
//
// Example usage:
//   // Register wildcard listener
//   app.Listen("user.*", userHandler)
//
//   // Check if wildcard matches
//   if app.HasWildcardListeners("user.created") {
//       fmt.Println("Wildcard listeners will handle this")
//   }
//
// Parameters:
//   - eventName: Event to check for wildcard matches (any type)
//
// Returns:
//   - bool: true if wildcard listeners match, false otherwise
func (app *Application) HasWildcardListeners(eventName any) bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

	eventNameStr := app.getEventName(eventName)
	for event, _ := range app.wildcards {
		if str.Of(eventNameStr).Is(app.getEventName(event)) {
			return true
		}
	}

	return false
}

// Listen registers event listeners with the dispatcher.
//
// This method supports multiple event formats and automatically handles
// listener registration, queue registration for async listeners, and
// wildcard pattern setup.
//
// Supported event formats:
//   - Single string: "user.created"
//   - String slice: []string{"user.created", "user.updated"}
//   - Single struct: UserCreatedEvent{}
//   - Struct slice: []UserCreatedEvent{event1, event2}
//   - Any other type: Uses reflection to extract name
//
// Example usage:
//   // Listen to string event
//   app.Listen("user.created", func(user User) {
//       fmt.Printf("User %s created\n", user.Name)
//   })
//
//   // Listen to multiple events
//   app.Listen([]string{"user.created", "user.updated"}, func(event any, data ...any) {
//       fmt.Printf("User event: %s\n", event)
//   })
//
//   // Listen to struct event
//   app.Listen(UserCreatedEvent{}, &UserCreatedListener{})
//
//   // Listen with wildcard pattern
//   app.Listen("user.*", func(event any, data ...any) {
//       fmt.Printf("Any user event: %s\n", event)
//   })
//
//   // Listen with queued listener
//   app.Listen("heavy.processing", &HeavyProcessingListener{})
//
// Parameters:
//   - events: Event(s) to listen for (string, []string, struct, []struct, any)
//   - listener: Function or object to handle the event
func (app *Application) Listen(events any, listener any) {
	app.mu.Lock()
	defer app.mu.Unlock()

	switch e := events.(type) {
	case string:
		app.setupEvents(e, listener)
	case []string:
		for _, eventName := range e {
			app.setupEvents(eventName, listener)
		}
	case event.Event:
		app.setupEvents(e, listener)
	case []event.Event:
		for _, evt := range e {
			app.setupEvents(evt, listener)
		}
	default:
		// Handle any other type using reflection
		if eventName := app.getEventName(events); eventName != "" {
			app.setupEvents(eventName, listener)
		}
	}
}

// Push defers an event for later batch processing.
//
// This method stores events with their payloads for later processing via
// Flush(). This is useful for collecting related events and processing
// them together for better performance or transactional consistency.
//
// Example usage:
//   // Collect events during a transaction
//   app.Push("user.created", []event.Arg{{Value: user1, Type: "User"}})
//   app.Push("user.created", []event.Arg{{Value: user2, Type: "User"}})
//   app.Push("user.created", []event.Arg{{Value: user3, Type: "User"}})
//
//   // Process all collected events after transaction commit
//   app.Flush("user.created")
//
// Parameters:
//   - eventName: Event to defer (any type)
//   - payload: Data to store with the event
func (app *Application) Push(eventName any, payload []event.Arg) {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.pushed[eventName] = append(app.pushed[eventName], payload...)
}

// Until fires an event and returns the first non-nil response.
//
// This method processes listeners sequentially and returns the first
// non-nil response, stopping further processing. Useful for validation,
// authorization, and other scenarios where you need a single response.
//
// Example usage:
//   // Validation example
//   result := app.Until("user.validate", []event.Arg{{Value: user, Type: "User"}})
//   if result != nil {
//       if valid, ok := result.(bool); ok && valid {
//           fmt.Println("User is valid")
//       }
//   }
//
//   // Authorization example
//   authorized := app.Until("user.authorize", []event.Arg{
//       {Value: user, Type: "User"},
//       {Value: "edit", Type: "action"},
//   })
//   if authorized == nil {
//       fmt.Println("Authorization failed")
//   }
//
// Parameters:
//   - eventName: Event to dispatch (any type)
//   - payload: Data to pass to listeners
//
// Returns:
//   - any: First non-nil response from a listener, or nil if none
func (app *Application) Until(eventName any, payload []event.Arg) any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	responses := app.invokeListeners(eventName, payload, true)
	if len(responses) > 0 {
		return responses[0]
	}
	return nil
}

// =============================================================================
// PRIVATE METHODS (Internal implementation, alphabetically ordered)
// =============================================================================

// callListener invokes a single listener with the provided event and payload.
//
// This method handles different listener types and formats, converting payloads
// as needed and managing async processing for queued listeners.
//
// Parameters:
//   - listener: The listener to invoke
//   - e: The event being processed
//   - args: Event arguments to pass to the listener
//
// Returns:
//   - any: Response from the listener, or nil if none
func (app *Application) callListener(listener any, e any, args []event.Arg) any {
	// Convert event.Arg to []any for most listeners
	payload := make([]any, len(args))
	for i, arg := range args {
		payload[i] = arg.Value
	}

	switch l := listener.(type) {
	case event.QueueListener:
		if l.ShouldQueue() {
			return app.queueListener(l, e, args)
		}
		return l.Handle(payload...)
	case event.EventQueueListener:
		if l.ShouldQueue() {
			return app.queueListener(eventToQueueListener(l, e), e, args)
		}
		return l.Handle(e, payload...)
	case event.Listener:
		return l.Handle(payload...)
	case event.EventListener:
		return l.Handle(e, payload...)
	case func(string, ...any) any:
		return l(app.getEventName(e), payload...)
	case func(any, ...any) any:
		return l(e, payload...)
	case func(...any) any:
		return l(payload...)
	default:
		return app.callReflectListener(listener, e, payload)
	}
}

// callReflectListener invokes listeners using reflection.
//
// This method handles generic function listeners by using reflection to
// call them with appropriate arguments, providing flexibility for various
// function signatures.
//
// Parameters:
//   - listener: The function listener to invoke
//   - event: The event being processed
//   - payload: Converted payload arguments
//
// Returns:
//   - any: Response from the listener, or nil if none
func (app *Application) callReflectListener(listener any, event any, payload []any) any {
	listenerValue := reflect.ValueOf(listener)
	listenerType := reflect.TypeOf(listener)

	if listenerType.Kind() != reflect.Func {
		return nil
	}

	numIn := listenerType.NumIn()
	args := make([]reflect.Value, 0, numIn)

	// Fill arguments from payload
	for i := 0; i < numIn && i < len(payload); i++ {
		args = append(args, reflect.ValueOf(payload[i]))
	}

	// Fill remaining arguments with zero values
	for len(args) < numIn {
		args = append(args, reflect.Zero(listenerType.In(len(args))))
	}

	results := listenerValue.Call(args)
	if len(results) > 0 {
		return results[0].Interface()
	}

	return nil
}

// getEventName extracts a string name from various event types.
//
// This method handles different event formats and converts them to
// string names for internal processing and matching.
//
// Parameters:
//   - evt: Event to extract name from (any type)
//
// Returns:
//   - string: Event name for internal use
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

// getWildcardListeners finds and caches wildcard listeners matching an event.
//
// This method searches through wildcard patterns and caches results for
// performance optimization on subsequent calls.
//
// Parameters:
//   - eventName: Event to find wildcard matches for (any type)
//
// Returns:
//   - []any: Slice of matching wildcard listeners
func (app *Application) getWildcardListeners(eventName any) []any {
	var wildcardListeners []any
	eventNameStr := app.getEventName(eventName)

	for event, wildcard := range app.wildcards {
		if str.Of(eventNameStr).Is(app.getEventName(event)) {
			wildcardListeners = append(wildcardListeners, wildcard...)
		}
	}

	// Cache results for performance
	app.wildcardsCache[eventName] = wildcardListeners

	return wildcardListeners
}

// invokeListeners processes all listeners for an event.
//
// This method finds all matching listeners and invokes them, collecting
// responses and handling the halt-on-first-response behavior for Until().
//
// Parameters:
//   - event: Event being processed
//   - payload: Data to pass to listeners
//   - halt: Whether to stop after first non-nil response
//
// Returns:
//   - []any: Responses from listeners
func (app *Application) invokeListeners(event any, payload []event.Arg, halt bool) []any {
	var responses []any
	listeners := app.prepareListeners(event)

	for _, listener := range listeners {
		response := app.callListener(listener, event, payload)
		if response != nil {
			responses = append(responses, response)
			if halt {
				break
			}
		}
	}

	return responses
}

// prepareListeners gathers all listeners for an event.
//
// This method combines direct listeners with wildcard matches, using
// cache when available for optimal performance.
//
// Parameters:
//   - event: Event to prepare listeners for
//
// Returns:
//   - []any: All listeners that should handle the event
func (app *Application) prepareListeners(event any) []any {
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

// queueListener handles asynchronous processing of queued listeners.
//
// This method sets up queue jobs with proper configuration (connection, queue,
// delays) and dispatches them for background processing.
//
// Parameters:
//   - listener: Queued listener to process
//   - event: Event being processed
//   - payload: Data to pass to the listener
//
// Returns:
//   - any: nil (async processing doesn't return immediate results)
func (app *Application) queueListener(listener event.QueueListener, event any, payload []event.Arg) any {
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

	task.Dispatch()

	return nil
}

// setupEvents registers a single event-listener pair.
//
// This method handles the registration logic for individual events,
// including queue registration for async listeners and wildcard setup.
//
// Parameters:
//   - e: Event to register
//   - listener: Listener to register for the event
func (app *Application) setupEvents(e any, listener any) {
	// Register queued listeners with the queue system
	if l, ok := listener.(event.EventQueueListener); ok {
		app.queue.Register([]queue.Job{eventToQueueListener(l, e)})
	} else if l, ok := listener.(event.QueueListener); ok {
		app.queue.Register([]queue.Job{l})
	}

	// Register the listener based on event type
	eventName := app.getEventName(e)
	if strings.Contains(eventName, "*") {
		app.setupWildcardListen(eventName, listener)
	} else {
		app.listeners[e] = append(app.listeners[e], listener)
	}
}

// setupWildcardListen registers wildcard pattern listeners.
//
// This method handles wildcard pattern registration and clears the
// wildcard cache to ensure fresh matching on subsequent calls.
//
// Parameters:
//   - eventName: Wildcard pattern to register
//   - listener: Listener to register for the pattern
func (app *Application) setupWildcardListen(eventName string, listener any) {
	app.wildcards[eventName] = append(app.wildcards[eventName], listener)
	// Clear cache to ensure fresh wildcard matching
	app.wildcardsCache = make(map[any][]any)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// eventArgsToQueueArgs converts event arguments to queue arguments.
//
// This function transforms event.Arg slices to queue.Arg slices for
// compatibility with the queue system.
//
// Parameters:
//   - args: Event arguments to convert
//
// Returns:
//   - []queue.Arg: Converted queue arguments
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