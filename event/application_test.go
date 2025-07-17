package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

// =============================================================================
// TEST DATA AND HELPERS
// =============================================================================

// TestEvent is a simple test event implementation
type TestEvent struct {
	Name string
	ID   int
}

// TestListener is a simple test listener implementation
type TestListener struct {
	Called        bool
	ReceivedEvent any
	ReceivedArgs  []any
	Response      any
}

func (l *TestListener) Handle(args ...any) error {
	l.Called = true
	l.ReceivedArgs = args
	return nil
}

// TestEventListener is a test event listener implementation
type TestEventListener struct {
	Called        bool
	ReceivedEvent any
	ReceivedArgs  []any
	Response      any
}

func (l *TestEventListener) Handle(event any, args ...any) error {
	l.Called = true
	l.ReceivedEvent = event
	l.ReceivedArgs = args
	return nil
}

// TestQueueListener is a test queue listener implementation
type TestQueueListener struct {
	Called           bool
	ReceivedArgs     []any
	ShouldQueueValue bool
	SignatureValue   string
}

func (l *TestQueueListener) Handle(args ...any) error {
	l.Called = true
	l.ReceivedArgs = args
	return nil
}

func (l *TestQueueListener) ShouldQueue() bool {
	return l.ShouldQueueValue
}

func (l *TestQueueListener) Signature() string {
	return l.SignatureValue
}

// =============================================================================
// CORE FUNCTIONALITY TESTS
// =============================================================================

func TestNewApplication(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	assert.NotNil(t, app)
	assert.Equal(t, mockQueue, app.queue)
	assert.NotNil(t, app.listeners)
	assert.NotNil(t, app.wildcards)
	assert.NotNil(t, app.wildcardsCache)
	assert.NotNil(t, app.pushed)
}

func TestApplication_Listen_StringEvent(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	err := app.Listen("user.created", listener)
	assert.NoError(t, err)

	assert.True(t, app.HasListeners("user.created"))
	listeners := app.GetListeners("user.created")
	assert.Len(t, listeners, 1)
	assert.Equal(t, listener, listeners[0])
}

func TestApplication_Listen_MultipleStringEvents(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	err := app.Listen([]string{"user.created", "user.updated"}, listener)
	assert.NoError(t, err)

	assert.True(t, app.HasListeners("user.created"))
	assert.True(t, app.HasListeners("user.updated"))
	assert.Equal(t, listener, app.GetListeners("user.created")[0])
	assert.Equal(t, listener, app.GetListeners("user.updated")[0])
}

func TestApplication_Listen_StructEvent(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}
	event := TestEvent{Name: "test", ID: 1}

	err := app.Listen(event, listener)
	assert.NoError(t, err)

	assert.True(t, app.HasListeners(event))
	listeners := app.GetListeners(event)
	assert.Len(t, listeners, 1)
	assert.Equal(t, listener, listeners[0])
}

func TestApplication_Listen_WildcardEvent(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	err := app.Listen("user.*", listener)
	assert.NoError(t, err)

	assert.True(t, app.HasWildcardListeners("user.created"))
	assert.True(t, app.HasWildcardListeners("user.updated"))
	assert.True(t, app.HasWildcardListeners("user.deleted"))
	assert.False(t, app.HasWildcardListeners("order.created"))
}

func TestApplication_Dispatch_StringEvent(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	err := app.Listen("user.created", func(args ...any) any {
		return "test response"
	})
	assert.NoError(t, err)

	responses := app.Dispatch("user.created", []event.Arg{{Value: "test", Type: "string"}})

	assert.Len(t, responses, 1)
	assert.Equal(t, "test response", responses[0])
}

func TestApplication_Dispatch_WithWildcardListeners(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	directListener := &TestListener{}
	wildcardListener := &TestListener{}

	err := app.Listen("user.created", directListener)
	assert.NoError(t, err)
	err = app.Listen("user.*", wildcardListener)
	assert.NoError(t, err)

	app.Dispatch("user.created", []event.Arg{{Value: "test", Type: "string"}})

	// Both listeners should be called
	assert.True(t, directListener.Called)
	assert.True(t, wildcardListener.Called)
}

func TestApplication_Until_ReturnsFirstNonNilResponse(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	err := app.Listen("user.validate", func(args ...any) any {
		return nil // First listener returns nil
	})
	assert.NoError(t, err)
	err = app.Listen("user.validate", func(args ...any) any {
		return "valid" // Second listener returns non-nil
	})
	assert.NoError(t, err)
	err = app.Listen("user.validate", func(args ...any) any {
		return "should not be called" // Third listener should not be called
	})
	assert.NoError(t, err)

	result := app.Until("user.validate", []event.Arg{{Value: "test", Type: "string"}})

	assert.Equal(t, "valid", result)
}

func TestApplication_Push_And_Flush(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	err := app.Listen("user.created", listener)
	assert.NoError(t, err)

	// Push events
	app.Push("user.created", []event.Arg{{Value: "user1", Type: "string"}})
	app.Push("user.created", []event.Arg{{Value: "user2", Type: "string"}})

	// Listener should not be called yet
	assert.False(t, listener.Called)

	// Flush events
	app.Flush("user.created")

	// Listener should be called now
	assert.True(t, listener.Called)
}

func TestApplication_Forget_RemovesListeners(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	err := app.Listen("user.created", listener)
	assert.NoError(t, err)
	assert.True(t, app.HasListeners("user.created"))

	app.Forget("user.created")
	assert.False(t, app.HasListeners("user.created"))
}

func TestApplication_Forget_RemovesWildcardListeners(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	err := app.Listen("user.*", listener)
	assert.NoError(t, err)
	assert.True(t, app.HasWildcardListeners("user.created"))

	app.Forget("user.*")
	assert.False(t, app.HasWildcardListeners("user.created"))
}

func TestApplication_ForgetPushed_ClearsPushedEvents(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	err := app.Listen("user.created", listener)
	assert.NoError(t, err)
	app.Push("user.created", []event.Arg{{Value: "user1", Type: "string"}})

	app.ForgetPushed()
	app.Flush("user.created")

	// Listener should not be called since pushed events were cleared
	assert.False(t, listener.Called)
}

func TestApplication_QueueListener_ShouldQueue(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	queueListener := &TestQueueListener{
		ShouldQueueValue: true,
		SignatureValue:   "test.queue.listener",
	}

	// Mock queue expectations for registration
	mockQueue.EXPECT().Register(mock.MatchedBy(func(jobs []queue.Job) bool {
		return len(jobs) == 1 && jobs[0] == queueListener
	})).Once()

	err := app.Listen("user.created", queueListener)
	assert.NoError(t, err)

	// Verify listener was registered
	assert.True(t, app.HasListeners("user.created"))
}

func TestApplication_QueueListener_ShouldNotQueue(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	queueListener := &TestQueueListener{
		ShouldQueueValue: false,
		SignatureValue:   "test.queue.listener",
	}

	// Mock queue expectations for registration
	mockQueue.EXPECT().Register(mock.MatchedBy(func(jobs []queue.Job) bool {
		return len(jobs) == 1 && jobs[0] == queueListener
	})).Once()

	err := app.Listen("user.created", queueListener)
	assert.NoError(t, err)

	// Dispatch event - should call listener directly
	app.Dispatch("user.created", []event.Arg{{Value: "test", Type: "string"}})

	// Verify queue listener was called directly (not queued)
	assert.True(t, queueListener.Called)
}

func TestApplication_FunctionListener_Variants(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	var stringEventCalled, anyEventCalled, varArgsCalled bool

	// Test string event function (func(string, ...any) error pattern)
	e := app.Listen("test.string", func(eventName string, args ...any) error {
		stringEventCalled = true
		assert.Equal(t, "test.string", eventName)
		assert.Len(t, args, 1)
		assert.Equal(t, "test", args[0])
		return nil
	})
	assert.NoError(t, e)

	// Test any event function (func(any, ...any) error pattern)
	e = app.Listen("test.any", func(event any, args ...any) error {
		anyEventCalled = true
		assert.Equal(t, "test.any", event)
		assert.Len(t, args, 1)
		assert.Equal(t, "test", args[0])
		return nil
	})
	assert.NoError(t, e)

	// Test variadic args function (func(...any) error pattern)
	e = app.Listen("test.varargs", func(args ...any) error {
		varArgsCalled = true
		assert.Len(t, args, 1)
		assert.Equal(t, "test", args[0])
		return nil
	})
	assert.NoError(t, e)
	// Dispatch events
	app.Dispatch("test.string", []event.Arg{{Value: "test", Type: "string"}})
	app.Dispatch("test.any", []event.Arg{{Value: "test", Type: "string"}})
	app.Dispatch("test.varargs", []event.Arg{{Value: "test", Type: "string"}})

	assert.True(t, stringEventCalled)
	assert.True(t, anyEventCalled)
	assert.True(t, varArgsCalled)
}

// =============================================================================
// EDGE CASES AND ERROR HANDLING TESTS
// =============================================================================

func TestApplication_Dispatch_NoListeners(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	responses := app.Dispatch("nonexistent.event", []event.Arg{{Value: "test", Type: "string"}})

	assert.Len(t, responses, 0)
}

func TestApplication_Until_NoListeners(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	result := app.Until("nonexistent.event", []event.Arg{{Value: "test", Type: "string"}})

	assert.Nil(t, result)
}

func TestApplication_Until_AllListenersReturnNil(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	e := app.Listen("test.event", func(args ...any) any {
		return nil
	})
	assert.NoError(t, e)

	e = app.Listen("test.event", func(args ...any) any {
		return nil
	})
	assert.NoError(t, e)
	result := app.Until("test.event", []event.Arg{{Value: "test", Type: "string"}})

	assert.Nil(t, result)
}

func TestApplication_WildcardMatching_EdgeCases(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	// Test multiple asterisks
	err := app.Listen("user.*.action.*", listener)
	assert.NoError(t, err)

	err = app.Listen("user.*.action.*", listener)
	assert.NoError(t, err)
	assert.True(t, app.HasWildcardListeners("user.123.action.create"))
	assert.True(t, app.HasWildcardListeners("user.456.action.update"))
	assert.False(t, app.HasWildcardListeners("user.123.create"))
}

func TestApplication_ConcurrentAccess(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	// Test concurrent listen and dispatch
	go func() {
		for i := 0; i < 100; i++ {
			e := app.Listen("concurrent.test", listener)
			assert.NoError(t, e)
		}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			app.Dispatch("concurrent.test", []event.Arg{{Value: i, Type: "int"}})
		}
	}()

	time.Sleep(100 * time.Millisecond)
	assert.True(t, app.HasListeners("concurrent.test"))
}

// =============================================================================
// PERFORMANCE AND CACHING TESTS
// =============================================================================

func TestApplication_WildcardCaching(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	err := app.Listen("user.*", listener)
	assert.NoError(t, err)

	// First call should populate cache
	listeners1 := app.GetListeners("user.created")
	assert.Len(t, listeners1, 1)

	// Second call should use cache
	listeners2 := app.GetListeners("user.created")
	assert.Equal(t, listeners1, listeners2)

	// Verify cache was populated
	assert.NotEmpty(t, app.wildcardsCache)
}

func TestApplication_CacheClearedOnWildcardRegistration(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener1 := &TestListener{}
	listener2 := &TestListener{}

	err := app.Listen("user.*", listener1)
	assert.NoError(t, err)
	app.GetListeners("user.created") // Populate cache

	assert.NotEmpty(t, app.wildcardsCache)

	// Adding new wildcard should clear cache
	err = app.Listen("order.*", listener2)
	assert.NoError(t, err)
	assert.Empty(t, app.wildcardsCache)
}

func BenchmarkApplication_Dispatch_DirectListeners(b *testing.B) {
	mockQueue := mocksqueue.NewQueue(b)
	app := NewApplication(mockQueue)

	// Register 100 listeners
	for i := 0; i < 100; i++ {
		e := app.Listen("benchmark.test", func(args ...any) any {
			return "response"
		})
		assert.NoError(b, e)
	}

	args := []event.Arg{{Value: "test", Type: "string"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Dispatch("benchmark.test", args)
	}
}

func BenchmarkApplication_Dispatch_WildcardListeners(b *testing.B) {
	mockQueue := mocksqueue.NewQueue(b)
	app := NewApplication(mockQueue)

	// Register 100 wildcard listeners
	for i := 0; i < 100; i++ {
		e := app.Listen("benchmark.*", func(args ...any) any {
			return "response"
		})
		assert.NoError(b, e)
	}

	args := []event.Arg{{Value: "test", Type: "string"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Dispatch("benchmark.test", args)
	}
}

// =============================================================================
// ADDITIONAL COVERAGE TESTS
// =============================================================================

func TestApplication_GetEventName_AllTypes(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	// Test string event
	assert.Equal(t, "test.event", app.getEventName("test.event"))

	// Test event with signature
	sigEvent := &TestSignatureEvent{}
	assert.Equal(t, "custom.signature", app.getEventName(sigEvent))

	// Test regular event struct
	testEvent := TestEvent{Name: "test", ID: 1}
	assert.Equal(t, "TestEvent", app.getEventName(testEvent))

	// Test pointer to event struct
	assert.Equal(t, "TestEvent", app.getEventName(&testEvent))

	// Test nil event
	assert.Equal(t, "", app.getEventName(nil))
}

func TestApplication_EventQueueListener_Integration(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	eventQueueListener := &TestEventQueueListener{
		ShouldQueueValue: true,
		SignatureValue:   "test.event.queue",
	}

	// Mock queue expectations for registration
	mockQueue.EXPECT().Register(mock.MatchedBy(func(jobs []queue.Job) bool {
		return len(jobs) == 1
	})).Once()

	// Register event queue listener
	err := app.Listen("test.event", eventQueueListener)
	assert.NoError(t, err)

	// Test that the listener was registered
	assert.True(t, app.HasListeners("test.event"))
}

func TestApplication_Listen_ClosureEventListener(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	var called bool
	var receivedEvent any

	// Test closure that takes event.Event parameter
	e := app.Listen(func(evt *TestEvent) error {
		called = true
		receivedEvent = evt
		return nil
	})
	assert.NoError(t, e)
	testEvent := &TestEvent{Name: "test", ID: 1}
	app.Dispatch(testEvent)

	assert.True(t, called)
	assert.Equal(t, testEvent, receivedEvent)
}

func TestApplication_Listen_InvalidEventType(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	// Test with nil event (should use reflection and get empty string)
	err := app.Listen(nil, func(args ...any) error { return nil })
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event type")
}

func TestApplication_Listen_NoListener(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	// Test with no listener provided
	err := app.Listen("test.event")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "listener is required")

	// Test with closure but no event parameter
	err = app.Listen(func() error { return nil })
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "listener is required")
}

func TestApplication_CallListener_AllTypes(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	args := []event.Arg{{Value: "test", Type: "string"}}

	// Test event.Listener
	listener := &TestListener{}
	app.callListener(listener, "test.event", args)
	assert.True(t, listener.Called)

	// Test event.EventListener
	eventListener := &TestEventListener{}
	app.callListener(eventListener, "test.event", args)
	assert.True(t, eventListener.Called)

	// Test func(any) error
	var anyEventCalled bool
	app.callListener(func(evt any) error {
		anyEventCalled = true
		assert.Equal(t, "test.event", evt)
		return nil
	}, "test.event", args)
	assert.True(t, anyEventCalled)

	// Test invalid listener type
	result := app.callListener("invalid", "test.event", args)
	assert.Nil(t, result)
}

func TestApplication_CallReflectListener_EdgeCases(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	// Test non-function listener
	result := app.callReflectListener("not a function", "test.event", []any{"test"})
	assert.Nil(t, result)

	// Test function with more parameters than provided
	var called bool
	app.callReflectListener(func(a, b, c string) error {
		called = true
		assert.Equal(t, "test", a)
		assert.Equal(t, "", b) // Should be zero value
		assert.Equal(t, "", c) // Should be zero value
		return nil
	}, "test.event", []any{"test"})
	assert.True(t, called)

	// Test function with return value
	result = app.callReflectListener(func() string {
		return "test result"
	}, "test.event", []any{})
	assert.Equal(t, "test result", result)

	// Test function with no return value
	result = app.callReflectListener(func() {}, "test.event", []any{})
	assert.Nil(t, result)
}

func TestApplication_HasListeners_DirectOnly(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	// Test with only direct listeners
	app.listeners["test.event"] = []any{listener}
	assert.True(t, app.HasListeners("test.event"))

	// Test with empty listeners array
	app.listeners["empty.event"] = []any{}
	assert.False(t, app.HasListeners("empty.event"))
}

func TestApplication_Forget_CacheClearing(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	// Set up wildcard cache
	app.wildcardsCache["user.created"] = []any{listener}
	e := app.Listen("user.*", listener)
	assert.NoError(t, e)

	// Forget should clear related cache
	app.Forget("user.*")
	assert.Empty(t, app.wildcardsCache)
}

// =============================================================================
// HELPER STRUCTS FOR ADDITIONAL TESTS
// =============================================================================

type TestSignatureEvent struct{}

func (e *TestSignatureEvent) Signature() string {
	return "custom.signature"
}

type TestEventQueueListener struct {
	Called           bool
	ReceivedEvent    any
	ReceivedArgs     []any
	ShouldQueueValue bool
	SignatureValue   string
}

func (l *TestEventQueueListener) Handle(event any, args ...any) error {
	l.Called = true
	l.ReceivedEvent = event
	l.ReceivedArgs = args
	return nil
}

func (l *TestEventQueueListener) ShouldQueue() bool {
	return l.ShouldQueueValue
}

func (l *TestEventQueueListener) Signature() string {
	return l.SignatureValue
}

func TestEventArgsToQueueArgs(t *testing.T) {
	eventArgs := []event.Arg{
		{Value: "string_value", Type: "string"},
		{Value: 42, Type: "int"},
		{Value: true, Type: "bool"},
	}

	queueArgs := eventArgsToQueueArgs(eventArgs)

	assert.Len(t, queueArgs, 3)
	assert.Equal(t, "string_value", queueArgs[0].Value)
	assert.Equal(t, "string", queueArgs[0].Type)
	assert.Equal(t, 42, queueArgs[1].Value)
	assert.Equal(t, "int", queueArgs[1].Type)
	assert.Equal(t, true, queueArgs[2].Value)
	assert.Equal(t, "bool", queueArgs[2].Type)
}

func TestEventArgsToQueueArgs_Empty(t *testing.T) {
	eventArgs := []event.Arg{}
	queueArgs := eventArgsToQueueArgs(eventArgs)
	assert.Empty(t, queueArgs)
}
