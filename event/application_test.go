package event

import (
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
		evt := &TestEvent{}
		app.listeners = map[string][]any{
			"TestEvent": {
				func(e *TestEvent) any {
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
