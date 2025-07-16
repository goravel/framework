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

type Application struct {
	listeners      map[any][]any
	wildcards      map[any][]any
	wildcardsCache map[any][]any
	pushed         map[any][]event.Arg
	mu             sync.RWMutex
	queue          queue.Queue
}

func NewApplication(queue queue.Queue) *Application {
	return &Application{
		queue:          queue,
		listeners:      make(map[any][]any),
		wildcards:      make(map[any][]any),
		wildcardsCache: make(map[any][]any),
		pushed:         make(map[any][]event.Arg),
	}
}

func (app *Application) Dispatch(event any, payload []event.Arg) []any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	return app.invokeListeners(event, payload, false)
}

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

func (app *Application) Forget(eventName string) {
	app.mu.Lock()
	defer app.mu.Unlock()

	if strings.Contains(eventName, "*") {
		delete(app.wildcards, eventName)
	} else {
		delete(app.listeners, eventName)
	}

	for key, _ := range app.wildcardsCache {
		if str.Of(app.getEventName(key)).Is(eventName) {
			delete(app.wildcardsCache, key)
		}
	}
}

func (app *Application) ForgetPushed() {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.pushed = make(map[any][]event.Arg)
}

func (app *Application) GetListeners(eventName any) []any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	return app.prepareListeners(eventName)
}

func (app *Application) HasListeners(eventName any) bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

	if listeners, exists := app.listeners[eventName]; exists && len(listeners) > 0 {
		return true
	}

	if listeners, exists := app.wildcards[eventName]; exists && len(listeners) > 0 {
		return true
	}

	return app.HasWildcardListeners(eventName)
}

func (app *Application) HasWildcardListeners(eventName any) bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

	for event, _ := range app.wildcards {
		if str.Of(app.getEventName(event)).Is(app.getEventName(eventName)) {
			return true
		}
	}

	return false
}

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
		// Handle slice of struct events
		for _, evt := range e {
			app.setupEvents(evt, listener)
		}
	default:
		// Try to handle any other type by converting to string
		if eventName := app.getEventName(events); eventName != "" {
			app.setupEvents(eventName, listener)
		}
	}
}

func (app *Application) Push(eventName any, payload []event.Arg) {
	app.mu.Lock()
	defer app.mu.Unlock()

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

func (app *Application) getWildcardListeners(eventName any) []any {
	var wildcardListeners []any
	for event, wildcard := range app.wildcards {
		if str.Of(app.getEventName(event)).Is(app.getEventName(eventName)) {
			wildcardListeners = append(wildcardListeners, wildcard...)
		}
	}

	app.wildcardsCache[eventName] = wildcardListeners

	return wildcardListeners
}

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

func (app *Application) callListener(listener any, e any, args []event.Arg) any {

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

func (app *Application) queueListener(listener event.QueueListener, event any, payload []event.Arg) any {
	task := app.queue.Job(listener, eventArgsToQueueArgs(payload))

	ref := reflect.ValueOf(listener)

	if method := ref.MethodByName("ViaConnection"); method.IsValid() {
		result := method.Call(nil)
		if len(result) > 0 {
			viaConnection := result[0].String()
			task.OnConnection(viaConnection)
		}
	}

	if method := ref.MethodByName("ViaQueue"); method.IsValid() {
		result := method.Call(nil)
		if len(result) > 0 {
			viaQueue := result[0].String()
			task.OnQueue(viaQueue)
		}
	}

	if method := ref.MethodByName("WithDelay"); method.IsValid() {
		result := method.Call(nil)
		if len(result) > 0 {
			// If WithDelay returns seconds as int64
			delaySeconds := result[0].Int()
			task.Delay(time.Now().Add(time.Duration(delaySeconds) * time.Second))
		}
	}

	task.Dispatch()

	return nil
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

func (app *Application) callReflectListener(listener any, event any, payload []any) any {
	listenerValue := reflect.ValueOf(listener)
	listenerType := reflect.TypeOf(listener)

	if listenerType.Kind() != reflect.Func {
		return nil
	}

	numIn := listenerType.NumIn()
	args := make([]reflect.Value, 0, numIn)

	for i := 0; i < numIn && i < len(payload); i++ {
		args = append(args, reflect.ValueOf(payload[i]))
	}

	for len(args) < numIn {
		args = append(args, reflect.Zero(listenerType.In(len(args))))
	}

	results := listenerValue.Call(args)
	if len(results) > 0 {
		return results[0].Interface()
	}

	return nil
}

func (app *Application) prepareListeners(event any) []any {
	var allListeners []any

	if listeners, exists := app.listeners[event]; exists {
		allListeners = append(allListeners, listeners...)
	}

	if listeners, exists := app.wildcardsCache[event]; exists {
		allListeners = append(allListeners, listeners...)
	} else {
		wildcardListeners := app.getWildcardListeners(app.getEventName(event))
		allListeners = append(allListeners, wildcardListeners...)
	}

	return allListeners
}

func (app *Application) getEventName(evt any) string {
	switch e := evt.(type) {
	case string:
		return e
	case event.Event:
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

func (app *Application) setupEvents(e any, listener any) {

	if l, ok := listener.(event.EventQueueListener); ok {
		app.queue.Register([]queue.Job{eventToQueueListener(l, e)})
	} else if l, ok := listener.(event.QueueListener); ok {
		app.queue.Register([]queue.Job{l})
	}

	eventName := app.getEventName(e)
	if strings.Contains(eventName, "*") {
		app.setupWildcardListen(eventName, listener)
	} else {
		app.listeners[e] = append(app.listeners[e], listener)
	}
}

func (app *Application) setupWildcardListen(eventName string, listener any) {
	app.wildcards[eventName] = append(app.wildcards[eventName], listener)
	app.wildcardsCache = make(map[any][]any)
}
