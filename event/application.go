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

type Application struct {
	listeners      map[string][]any       // Direct event listeners
	wildcards      map[string][]any       // Wildcard pattern listeners
	wildcardsCache map[string][]any       // Cache for wildcard matching performance
	pushed         map[string][]event.Arg // Deferred events for batch processing
	mu             sync.RWMutex           // Thread-safe operations
	queue          queue.Queue            // Queue for asynchronous processing
}

func NewApplication(queue queue.Queue) *Application {
	return &Application{
		queue:          queue,
		listeners:      make(map[string][]any),
		wildcards:      make(map[string][]any),
		wildcardsCache: make(map[string][]any),
		pushed:         make(map[string][]event.Arg),
	}
}

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

func (app *Application) Flush(event any) {
	app.mu.Lock()
	defer app.mu.Unlock()

	eventName := app.getEventName(event)

	if payloads, exists := app.pushed[eventName]; exists {
		delete(app.pushed, eventName)
		app.mu.Unlock()

		app.Dispatch(eventName, payloads)

		app.mu.Lock()
	}
}

func (app *Application) Forget(event any) {
	app.mu.Lock()
	defer app.mu.Unlock()

	eventName := app.getEventName(event)

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

func (app *Application) ForgetPushed() {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.pushed = make(map[string][]event.Arg)
}

func (app *Application) GetListeners(event any) []any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	return app.prepareListeners(app.getEventName(event))
}

func (app *Application) HasListeners(event any) bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

	eventName := app.getEventName(event)

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
	case event.Event:
		app.setupEvents(e, listener[0])
	case []event.Event:
		for _, evt := range e {
			app.setupEvents(evt, listener[0])
		}
	default:
		if eventName := app.getEventName(events); eventName != "" {
			app.setupEvents(eventName, listener[0])
		} else {
			return errors.New("invalid event type")
		}
	}

	return nil
}

func (app *Application) Push(event any, payload []event.Arg) {
	app.mu.Lock()
	defer app.mu.Unlock()

	eventName := app.getEventName(event)

	app.pushed[eventName] = append(app.pushed[eventName], payload...)
}

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

func (app *Application) callReflectListener(listener any, event any, payload []any) any {
	listenerValue := reflect.ValueOf(listener)
	listenerType := listenerValue.Type()

	if listenerType.Kind() != reflect.Func {
		return nil
	}

	numIn := listenerType.NumIn()

	args := make([]reflect.Value, 0, numIn)

	if numIn > 0 && app.isEventType(listenerType.In(0)) {
		args = append(args, reflect.ValueOf(event))
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

func (app *Application) invokeListeners(event any, payload []event.Arg, halt bool) []any {
	var responses []any
	listeners := app.prepareListeners(app.getEventName(event))
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
