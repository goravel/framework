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

	app.Listen("user.created", listener)

	assert.True(t, app.HasListeners("user.created"))
	listeners := app.GetListeners("user.created")
	assert.Len(t, listeners, 1)
	assert.Equal(t, listener, listeners[0])
}

func TestApplication_Listen_MultipleStringEvents(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	app.Listen([]string{"user.created", "user.updated"}, listener)

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

	app.Listen(event, listener)

	assert.True(t, app.HasListeners(event))
	listeners := app.GetListeners(event)
	assert.Len(t, listeners, 1)
	assert.Equal(t, listener, listeners[0])
}

func TestApplication_Listen_WildcardEvent(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	app.Listen("user.*", listener)

	assert.True(t, app.HasWildcardListeners("user.created"))
	assert.True(t, app.HasWildcardListeners("user.updated"))
	assert.True(t, app.HasWildcardListeners("user.deleted"))
	assert.False(t, app.HasWildcardListeners("order.created"))
}

func TestApplication_Dispatch_StringEvent(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	app.Listen("user.created", func(args ...any) any {
		return "test response"
	})

	responses := app.Dispatch("user.created", []event.Arg{{Value: "test", Type: "string"}})

	assert.Len(t, responses, 1)
	assert.Equal(t, "test response", responses[0])
}

func TestApplication_Dispatch_WithWildcardListeners(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	directListener := &TestListener{}
	wildcardListener := &TestListener{}

	app.Listen("user.created", directListener)
	app.Listen("user.*", wildcardListener)

	app.Dispatch("user.created", []event.Arg{{Value: "test", Type: "string"}})

	// Both listeners should be called
	assert.True(t, directListener.Called)
	assert.True(t, wildcardListener.Called)
}

func TestApplication_Until_ReturnsFirstNonNilResponse(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	app.Listen("user.validate", func(args ...any) any {
		return nil // First listener returns nil
	})
	app.Listen("user.validate", func(args ...any) any {
		return "valid" // Second listener returns non-nil
	})
	app.Listen("user.validate", func(args ...any) any {
		return "should not be called" // Third listener should not be called
	})

	result := app.Until("user.validate", []event.Arg{{Value: "test", Type: "string"}})

	assert.Equal(t, "valid", result)
}

func TestApplication_Push_And_Flush(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	app.Listen("user.created", listener)

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

	app.Listen("user.created", listener)
	assert.True(t, app.HasListeners("user.created"))

	app.Forget("user.created")
	assert.False(t, app.HasListeners("user.created"))
}

func TestApplication_Forget_RemovesWildcardListeners(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	app.Listen("user.*", listener)
	assert.True(t, app.HasWildcardListeners("user.created"))

	app.Forget("user.*")
	assert.False(t, app.HasWildcardListeners("user.created"))
}

func TestApplication_ForgetPushed_ClearsPushedEvents(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	app.Listen("user.created", listener)
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

	app.Listen("user.created", queueListener)

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

	app.Listen("user.created", queueListener)

	// Dispatch event - should call listener directly
	app.Dispatch("user.created", []event.Arg{{Value: "test", Type: "string"}})

	// Verify queue listener was called directly (not queued)
	assert.True(t, queueListener.Called)
}

func TestApplication_FunctionListener_Variants(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)

	var stringEventCalled, anyEventCalled, varArgsCalled bool

	// Test string event function
	app.Listen("test.string", func(eventName string, args ...any) any {
		stringEventCalled = true
		assert.Equal(t, "test.string", eventName)
		return "string event response"
	})

	// Test any event function
	app.Listen("test.any", func(event any, args ...any) any {
		anyEventCalled = true
		assert.Equal(t, "test.any", event)
		return "any event response"
	})

	// Test variadic args function
	app.Listen("test.varargs", func(args ...any) any {
		varArgsCalled = true
		return "varargs response"
	})

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

	app.Listen("test.event", func(args ...any) any {
		return nil
	})
	app.Listen("test.event", func(args ...any) any {
		return nil
	})

	result := app.Until("test.event", []event.Arg{{Value: "test", Type: "string"}})

	assert.Nil(t, result)
}

func TestApplication_WildcardMatching_EdgeCases(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockQueue)
	listener := &TestListener{}

	// Test multiple asterisks
	app.Listen("user.*.action.*", listener)

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
			app.Listen("concurrent.test", listener)
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

	app.Listen("user.*", listener)

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

	app.Listen("user.*", listener1)
	app.GetListeners("user.created") // Populate cache

	assert.NotEmpty(t, app.wildcardsCache)

	// Adding new wildcard should clear cache
	app.Listen("order.*", listener2)
	assert.Empty(t, app.wildcardsCache)
}

func BenchmarkApplication_Dispatch_DirectListeners(b *testing.B) {
	mockQueue := mocksqueue.NewQueue(b)
	app := NewApplication(mockQueue)

	// Register 100 listeners
	for i := 0; i < 100; i++ {
		app.Listen("benchmark.test", func(args ...any) any {
			return "response"
		})
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
		app.Listen("benchmark.*", func(args ...any) any {
			return "response"
		})
	}

	args := []event.Arg{{Value: "test", Type: "string"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Dispatch("benchmark.test", args)
	}
}
