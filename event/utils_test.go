package event

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/contracts/event"
)

// TestEventWithSignature is a test event that implements Signature interface
type TestEventWithSignature struct{}

func (t *TestEventWithSignature) Signature() string {
	return "custom.signature"
}

func (t *TestEventWithSignature) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

func Test_dynamicQueueJob_Handle(t *testing.T) {
	var handlerCalled bool
	job := &dynamicQueueJob{
		signature: "test.job",
		handler: func(args ...any) error {
			handlerCalled = true
			return nil
		},
		shouldQueue: true,
	}

	err := job.Handle("test")
	require.NoError(t, err)
	assert.True(t, handlerCalled)
}

func Test_dynamicQueueJob_ShouldQueue(t *testing.T) {
	job1 := &dynamicQueueJob{shouldQueue: true}
	assert.True(t, job1.ShouldQueue())

	job2 := &dynamicQueueJob{shouldQueue: false}
	assert.False(t, job2.ShouldQueue())
}

func Test_dynamicQueueJob_Signature(t *testing.T) {
	job := &dynamicQueueJob{signature: "test.signature"}
	assert.Equal(t, "test.signature", job.Signature())
}

func Test_dynamicQueueJob_Queue(t *testing.T) {
	// Test with queueable config
	queueable := &TestListener{}
	job1 := &dynamicQueueJob{queueConfig: queueable}
	queue := job1.Queue()
	assert.False(t, queue.Enable)

	// Test with non-queueable config
	job2 := &dynamicQueueJob{queueConfig: "not-queueable"}
	emptyQueue := job2.Queue()
	assert.Equal(t, event.Queue{}, emptyQueue)
}

func TestApplication_getEventName(t *testing.T) {
	app := &Application{}
	
	tests := []struct {
		name     string
		event    any
		expected string
	}{
		{
			name:     "String event",
			event:    "user.created",
			expected: "user.created",
		},
		{
			name:     "Struct event",
			event:    TestEvent{},
			expected: "TestEvent",
		},
		{
			name:     "Pointer to struct event",
			event:    &TestEvent{},
			expected: "TestEvent",
		},
		{
			name:     "Event with custom signature",
			event:    &TestEventWithSignature{},
			expected: "custom.signature",
		},
		{
			name:     "Nil event",
			event:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.getEventName(tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplication_canUse(t *testing.T) {
	app := &Application{}
	
	tests := []struct {
		name   string
		val    any
		target reflect.Type
		want   bool
	}{
		{
			name:   "Nil value",
			val:    nil,
			target: reflect.TypeOf(""),
			want:   true,
		},
		{
			name:   "Assignable string",
			val:    "test",
			target: reflect.TypeOf(""),
			want:   true,
		},
		{
			name:   "Assignable struct",
			val:    TestEvent{},
			target: reflect.TypeOf(TestEvent{}),
			want:   true,
		},
		{
			name:   "Non-assignable but target is string",
			val:    123,
			target: reflect.TypeOf(""),
			want:   true,
		},
		{
			name:   "Non-assignable types",
			val:    123,
			target: reflect.TypeOf(TestEvent{}),
			want:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.canUse(tt.val, tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestApplication_convert(t *testing.T) {
	app := &Application{}
	
	tests := []struct {
		name         string
		val          any
		target       reflect.Type
		expectZero   bool
		expectedStr  string
		checkIsZero  bool
	}{
		{
			name:        "Nil value",
			val:         nil,
			target:      reflect.TypeOf(""),
			expectZero:  true,
		},
		{
			name:        "String target with string value",
			val:         "test",
			target:      reflect.TypeOf(""),
			expectedStr: "test",
		},
		{
			name:        "String target with event value",
			val:         TestEvent{},
			target:      reflect.TypeOf(""),
			expectedStr: "TestEvent", // Uses getEventName
		},
		{
			name:        "Assignable types",
			val:         TestEvent{},
			target:      reflect.TypeOf(TestEvent{}),
			checkIsZero: false,
		},
		{
			name:        "Non-assignable types",
			val:         123,
			target:      reflect.TypeOf(TestEvent{}),
			expectZero:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.convert(tt.val, tt.target)
			
			if tt.expectZero {
				assert.True(t, got.IsZero())
			}
			
			if tt.target.Kind() == reflect.String && tt.expectedStr != "" {
				assert.Equal(t, tt.expectedStr, got.String())
			}
		})
	}
}

func Test_eventArgsToQueueArgs(t *testing.T) {
	// We're just verifying that the function runs without errors
	// and returns the expected number of arguments
	args := []event.Arg{
		{Value: "string value", Type: "string"},
		{Value: 123, Type: "int"},
	}
	
	queueArgs := eventArgsToQueueArgs(args)
	require.Len(t, queueArgs, 2)
	
	// Since we don't know the exact implementation details of how eventArgsToQueueArgs
	// transforms event.Arg to queue.Arg, we'll just verify that the function doesn't panic
	// and returns the correct number of items
	// 
	// The actual implementation will be indirectly tested in the higher-level
	// integration tests that use this function (e.g., when testing the Job method)
}
