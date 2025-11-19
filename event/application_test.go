package event

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	mocksevent "github.com/goravel/framework/mocks/event"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

func TestApplication_Register(t *testing.T) {
	var (
		mockQueue *mocksqueue.Queue
	)

	tests := []struct {
		name   string
		events func() map[event.Event][]event.Listener
	}{
		{
			name: "MultipleEvents",
			events: func() map[event.Event][]event.Listener {
				event1 := mocksevent.NewEvent(t)
				event2 := mocksevent.NewEvent(t)
				listener1 := mocksevent.NewListener(t)
				listener1.EXPECT().Signature().Return("listener1").Twice()
				listener2 := mocksevent.NewListener(t)
				listener2.EXPECT().Signature().Return("listener2").Times(3)

				mockQueue.EXPECT().Register(mock.MatchedBy(func(listeners []queue.Job) bool {
					return assert.ElementsMatch(t, []queue.Job{
						listener1,
						listener2,
					}, listeners)
				})).Once()

				return map[event.Event][]event.Listener{
					event1: {
						listener1,
						listener2,
					},
					event2: {
						listener2,
					},
				}
			},
		},
		{
			name: "NoEvents",
			events: func() map[event.Event][]event.Listener {
				mockQueue.EXPECT().Register([]queue.Job(nil)).Once()

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQueue = mocksqueue.NewQueue(t)
			app := NewApplication(mockQueue)

			events := tt.events()
			app.Register(events)

			assert.Equal(t, len(events), len(app.GetEvents()))
			for e, listeners := range events {
				assert.ElementsMatch(t, listeners, app.GetEvents()[e])
			}
		})
	}
}

func TestApplication_Dispatch(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	// Test dispatching a string event
	t.Run("DispatchStringEvent", func(t *testing.T) {
		eventName := "test.event"
		app.listeners = map[string][]any{
			eventName: {
				func(name string) any { 
					assert.Equal(t, eventName, name)
					return "response"
				},
			},
		}
		
		responses := app.Dispatch(eventName)
		require.Len(t, responses, 1)
		assert.Equal(t, "response", responses[0])
	})

	// Test dispatching an event with payload
	t.Run("DispatchWithPayload", func(t *testing.T) {
		eventName := "user.created"
		payload := []event.Arg{{Value: "testuser", Type: "string"}}
		
		app.listeners = map[string][]any{
			eventName: {
				func(name string, username any) any {
					assert.Equal(t, eventName, name)
					assert.Equal(t, "testuser", username)
					return "processed"
				},
			},
		}
		
		responses := app.Dispatch(eventName, payload)
		require.Len(t, responses, 1)
		assert.Equal(t, "processed", responses[0])
	})

	// Test dispatching to wildcard listeners
	t.Run("DispatchWithWildcardListeners", func(t *testing.T) {
		eventName := "user.registered"
		app.listeners = map[string][]any{}
		app.wildcards = map[string][]any{
			"user.*": {
				func(name string) any {
					assert.Equal(t, eventName, name)
					return "wildcard"
				},
			},
		}
		app.wildcardsCache = make(map[string][]any)
		
		responses := app.Dispatch(eventName)
		require.Len(t, responses, 1)
		assert.Equal(t, "wildcard", responses[0])
	})

	// Test dispatching a struct event
	t.Run("DispatchStructEvent", func(t *testing.T) {
		evt := &TestEventCustom{}
		app.listeners = map[string][]any{
			"TestEventCustom": {
				func(e *TestEventCustom) any {
					assert.Equal(t, evt, e)
					return "received"
				},
			},
		}
		
		responses := app.Dispatch(evt)
		require.Len(t, responses, 1)
		assert.Equal(t, "received", responses[0])
	})
}

func TestApplication_Listen(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	// Test listening to a string event
	t.Run("ListenStringEvent", func(t *testing.T) {
		eventName := "test.event"
		handler := func() error { return nil }
		
		err := app.Listen(eventName, handler)
		require.NoError(t, err)
		
		listeners, exists := app.listeners[eventName]
		require.True(t, exists)
		require.Len(t, listeners, 1)
	})
	
	// Test listening to multiple string events
	t.Run("ListenMultipleEvents", func(t *testing.T) {
		events := []string{"event1", "event2"}
		handler := func() error { return nil }
		
		err := app.Listen(events, handler)
		require.NoError(t, err)
		
		for _, e := range events {
			listeners, exists := app.listeners[e]
			require.True(t, exists)
			require.Len(t, listeners, 1)
		}
	})
	
	// Test listening to a wildcard event
	t.Run("ListenWildcardEvent", func(t *testing.T) {
		eventName := "user.*"
		handler := func() error { return nil }
		
		err := app.Listen(eventName, handler)
		require.NoError(t, err)
		
		listeners, exists := app.wildcards[eventName]
		require.True(t, exists)
		require.Len(t, listeners, 1)
		assert.Empty(t, app.wildcardsCache) // Cache should be cleared
	})
	
	// Test listening to an event without a listener
	t.Run("ListenNoListener", func(t *testing.T) {
		err := app.Listen("event.name")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "listener is required")
	})
	
	// Test listening with closure - this should fail since closure doesn't match expected pattern
	t.Run("ListenClosure", func(t *testing.T) {
		closure := func(evt *TestEventCustom) error {
			return nil
		}
		
		err := app.Listen(closure)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "closure must accept exactly one event parameter")
	})
	
	// Test listening with invalid closure
	t.Run("ListenInvalidClosure", func(t *testing.T) {
		invalidClosure := func(a, b string) error {
			return nil
		}
		
		err := app.Listen(invalidClosure)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "closure must accept exactly one event parameter")
	})
	
	// Test listening to event interface
	t.Run("ListenEventInterface", func(t *testing.T) {
		// Create fresh app for this test
		mockQueueLocal := mocksqueue.NewQueue(t)
		appLocal := NewApplication(mockQueueLocal)
		
		evt := &TestEventCustom{}
		handler := func() error { return nil }
		
		err := appLocal.Listen(evt, handler)
		require.NoError(t, err)
		
		listeners, exists := appLocal.listeners["TestEventCustom"]
		require.True(t, exists)
		require.Len(t, listeners, 1)
	})
	
	// Test listening to slice of events
	t.Run("ListenEventSlice", func(t *testing.T) {
		// Create fresh app for this test to avoid interference
		mockQueueLocal := mocksqueue.NewQueue(t)
		appLocal := NewApplication(mockQueueLocal)
		
		events := []event.Event{&TestEventCustom{}, &TestEvent{}}
		handler := func() error { return nil }
		
		err := appLocal.Listen(events, handler)
		require.NoError(t, err)
		
		// Check both events were registered
		listeners1, exists1 := appLocal.listeners["TestEventCustom"]
		require.True(t, exists1)
		require.Len(t, listeners1, 1)
		
		listeners2, exists2 := appLocal.listeners["TestEvent"]
		require.True(t, exists2)
		require.Len(t, listeners2, 1)
	})
	
	// Test listening with valid event type that has getEventName 
	t.Run("ListenValidEventType", func(t *testing.T) {
		type ValidEvent struct{ Name string }
		validEvent := ValidEvent{Name: "valid"}
		handler := func() error { return nil }
		
		err := app.Listen(validEvent, handler)
		require.NoError(t, err)
		
		// Should register under the struct name
		listeners, exists := app.listeners["ValidEvent"]
		require.True(t, exists)
		require.Len(t, listeners, 1)
	})
}

func TestApplication_Job(t *testing.T) {
	// Setup test data
	mockEvent := mocksevent.NewEvent(t)
	mockListener := mocksevent.NewListener(t)
	args := []event.Arg{{Value: "test", Type: "string"}}
	
	// Test with registered event
	t.Run("RegisteredEvent", func(t *testing.T) {
		mockQueue := mocksqueue.NewQueue(t)
		app := NewApplication(mockQueue)
		app.events = map[event.Event][]event.Listener{
			mockEvent: {mockListener},
		}
		
		// We only test that Job returns a non-nil task
		// The actual dispatching happens in Task.Dispatch which is tested separately
		task := app.Job(mockEvent, args)
		require.NotNil(t, task)
		
		// Verify the task has the correct properties
		taskImpl, ok := task.(*Task)
		require.True(t, ok, "Task should be of type *Task")
		assert.Equal(t, mockEvent, taskImpl.event)
		assert.Equal(t, mockQueue, taskImpl.queue)
		assert.Equal(t, args, taskImpl.args)
		assert.Equal(t, []event.Listener{mockListener}, taskImpl.listeners)
	})
	
	// Test with unregistered event
	t.Run("UnregisteredEvent", func(t *testing.T) {
		mockQueue := mocksqueue.NewQueue(t)
		app := NewApplication(mockQueue)
		unregisteredEvent := mocksevent.NewEvent(t)
		
		// We only test that Job returns a non-nil task with empty listeners
		task := app.Job(unregisteredEvent, args)
		require.NotNil(t, task)
		
		// Verify the task has the correct properties
		taskImpl, ok := task.(*Task)
		require.True(t, ok, "Task should be of type *Task")
		assert.Equal(t, unregisteredEvent, taskImpl.event)
		assert.Equal(t, mockQueue, taskImpl.queue)
		assert.Equal(t, args, taskImpl.args)
		assert.Empty(t, taskImpl.listeners)
	})
}

// Test event with custom signature
type TestEventCustom struct{}

func (t *TestEventCustom) Signature() string {
	return "TestEventCustom"
}

func (t *TestEventCustom) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

// Test queue listener
type TestQueueListener struct {
	HandleFunc      func(args ...any) error
	ShouldQueueFunc bool
}

func (l *TestQueueListener) Handle(args ...any) error {
	if l.HandleFunc != nil {
		return l.HandleFunc(args...)
	}
	return nil
}

func (l *TestQueueListener) ShouldQueue() bool {
	return l.ShouldQueueFunc
}

func (l *TestQueueListener) Signature() string {
	return "TestQueueListener"
}

// Test event queue listener
type TestEventQueueListener struct {
	HandleFunc      func(evt any, args ...any) error
	ShouldQueueFunc bool
}

func (l *TestEventQueueListener) Handle(evt any, args ...any) error {
	if l.HandleFunc != nil {
		return l.HandleFunc(evt, args...)
	}
	return nil
}

func (l *TestEventQueueListener) ShouldQueue() bool {
	return l.ShouldQueueFunc
}

func (l *TestEventQueueListener) Signature() string {
	return "TestEventQueueListener"
}

func TestApplication_callListener(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	t.Run("FunctionListener", func(t *testing.T) {
		called := false
		listener := func(name string) any {
			called = true
			assert.Equal(t, "test", name)
			return "result"
		}
		
		result := app.callListener(listener, "test", nil)
		assert.True(t, called)
		assert.Equal(t, "result", result)
	})

	t.Run("StructWithHandleMethod", func(t *testing.T) {
		listener := &TestListener{}
		
		result := app.callListener(listener, "test", nil)
		assert.Nil(t, result)
	})

	t.Run("QueueListenerShouldNotQueue", func(t *testing.T) {
		listener := &TestQueueListener{
			ShouldQueueFunc: false,
		}
		
		// Should call the listener directly when not queued
		result := app.callListener(listener, "test", nil)
		assert.Nil(t, result)
	})

	t.Run("EventQueueListenerShouldNotQueue", func(t *testing.T) {
		listener := &TestEventQueueListener{
			ShouldQueueFunc: false,
		}
		
		// Should call the listener directly when not queued
		result := app.callListener(listener, "test", nil)
		assert.Nil(t, result)
	})

	t.Run("InvalidListener", func(t *testing.T) {
		type InvalidListener struct{}
		listener := &InvalidListener{}
		
		result := app.callListener(listener, "test", nil)
		assert.Nil(t, result)
	})
}

func TestApplication_callListenerHandle(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	t.Run("NoArguments", func(t *testing.T) {
		called := false
		fn := reflect.ValueOf(func() any {
			called = true
			return "no-args"
		})
		
		result := app.callListenerHandle(fn, "test", nil)
		assert.True(t, called)
		assert.Equal(t, "no-args", result)
	})

	t.Run("EventAsFirstArgument", func(t *testing.T) {
		called := false
		fn := reflect.ValueOf(func(name string) any {
			called = true
			assert.Equal(t, "test", name)
			return "with-event"
		})
		
		result := app.callListenerHandle(fn, "test", nil)
		assert.True(t, called)
		assert.Equal(t, "with-event", result)
	})

	t.Run("WithPayloadArgs", func(t *testing.T) {
		called := false
		args := []event.Arg{{Value: "payload1", Type: "string"}, {Value: 42, Type: "int"}}
		fn := reflect.ValueOf(func(name string, payload string, num int) any {
			called = true
			assert.Equal(t, "test", name)
			assert.Equal(t, "payload1", payload)
			assert.Equal(t, 42, num)
			return "with-payload"
		})
		
		result := app.callListenerHandle(fn, "test", args)
		assert.True(t, called)
		assert.Equal(t, "with-payload", result)
	})

	t.Run("VariadicSliceArgument", func(t *testing.T) {
		called := false
		args := []event.Arg{{Value: "arg1", Type: "string"}, {Value: "arg2", Type: "string"}}
		fn := reflect.ValueOf(func(name string, remaining []any) any {
			called = true
			assert.Equal(t, "test", name)
			assert.Len(t, remaining, 2)
			assert.Equal(t, "arg1", remaining[0])
			assert.Equal(t, "arg2", remaining[1])
			return "variadic"
		})
		
		result := app.callListenerHandle(fn, "test", args)
		assert.True(t, called)
		assert.Equal(t, "variadic", result)
	})
}


func TestApplication_getWildcardListeners(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	t.Run("MatchingWildcard", func(t *testing.T) {
		listener1 := func() {}
		listener2 := func() {}
		
		app.wildcards = map[string][]any{
			"user.*": {listener1, listener2},
			"order.*": {func() {}},
		}
		app.wildcardsCache = make(map[string][]any)
		
		listeners := app.getWildcardListeners("user.created")
		assert.Len(t, listeners, 2)
		
		// Check cache was updated
		cached, exists := app.wildcardsCache["user.created"]
		assert.True(t, exists)
		assert.Len(t, cached, 2)
	})

	t.Run("NoMatchingWildcard", func(t *testing.T) {
		app.wildcards = map[string][]any{
			"order.*": {func() {}},
		}
		app.wildcardsCache = make(map[string][]any)
		
		listeners := app.getWildcardListeners("user.created")
		assert.Empty(t, listeners)
	})
}

func TestApplication_prepareListeners(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	t.Run("DirectAndWildcardListeners", func(t *testing.T) {
		directListener := func() {}
		wildcardListener := func() {}
		
		app.listeners = map[string][]any{
			"user.created": {directListener},
		}
		app.wildcardsCache = map[string][]any{
			"user.created": {wildcardListener},
		}
		
		listeners := app.prepareListeners("user.created")
		assert.Len(t, listeners, 2)
	})

	t.Run("OnlyDirectListeners", func(t *testing.T) {
		listener := func() {}
		
		app.listeners = map[string][]any{
			"user.created": {listener},
		}
		app.wildcardsCache = make(map[string][]any)
		app.wildcards = make(map[string][]any)
		
		listeners := app.prepareListeners("user.created")
		assert.Len(t, listeners, 1)
	})

	t.Run("NoListeners", func(t *testing.T) {
		app.listeners = make(map[string][]any)
		app.wildcardsCache = make(map[string][]any)
		app.wildcards = make(map[string][]any)
		
		listeners := app.prepareListeners("nonexistent.event")
		assert.Empty(t, listeners)
	})
}
