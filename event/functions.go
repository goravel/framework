package event

import (
	"github.com/goravel/framework/contracts/event"
)

var defaultDispatcher event.Instance

// Event dispatches an event and return the responses.
func Event(event any, payload []event.Arg) []any {
	dispatcher := GetDefaultDispatcher()
	if dispatcher == nil {
		return nil
	}
	return dispatcher.Dispatch(event, payload)
}

// Forget removes all listeners for an event.
func Forget(eventName string) {
	dispatcher := GetDefaultDispatcher()
	if dispatcher != nil {
		dispatcher.Forget(eventName)
	}
}

// GetDefaultDispatcher returns the default event dispatcher.
func GetDefaultDispatcher() event.Instance {
	return defaultDispatcher
}

// GetListeners gets all listeners for an event.
func GetListeners(eventName string) []any {
	dispatcher := GetDefaultDispatcher()
	if dispatcher == nil {
		return nil
	}
	return dispatcher.GetListeners(eventName)
}

// HasListeners determines if a given event has listeners.
func HasListeners(eventName string) bool {
	dispatcher := GetDefaultDispatcher()
	if dispatcher == nil {
		return false
	}
	return dispatcher.HasListeners(eventName)
}

// Listen registers an event listener with the dispatcher.
func Listen(eventName string, listener any) {
	dispatcher := GetDefaultDispatcher()
	if dispatcher != nil {
		dispatcher.Listen(eventName, listener)
	}
}

// SetDefaultDispatcher sets the default event dispatcher.
func SetDefaultDispatcher(dispatcher event.Instance) {
	defaultDispatcher = dispatcher
}

// Until fires an event until the first non-null response.
func Until(eventName any, payload []event.Arg) any {
	dispatcher := GetDefaultDispatcher()
	if dispatcher == nil {
		return nil
	}
	return dispatcher.Until(eventName, payload)
}
