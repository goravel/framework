package event

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/event"
)

// TestEventQueueListenerWrapper for testing the wrapper
type TestEventQueueListenerWrapper struct {
	Called           bool
	ReceivedEvent    any
	ReceivedArgs     []any
	ShouldQueueValue bool
	SignatureValue   string
}

func (l *TestEventQueueListenerWrapper) Handle(evt any, args ...any) error {
	l.Called = true
	l.ReceivedEvent = evt
	l.ReceivedArgs = args
	return nil
}

func (l *TestEventQueueListenerWrapper) ShouldQueue() bool {
	return l.ShouldQueueValue
}

func (l *TestEventQueueListenerWrapper) Signature() string {
	return l.SignatureValue
}

func TestEventQueueWrapper_Handle(t *testing.T) {
	listener := &TestEventQueueListenerWrapper{}
	wrapper := &eventQueueWrapper{
		listener: listener,
		event:    "test.event",
	}

	err := wrapper.Handle("arg1", "arg2")

	assert.NoError(t, err)
	assert.True(t, listener.Called)
	assert.Equal(t, "test.event", listener.ReceivedEvent)
	assert.Equal(t, []any{"arg1", "arg2"}, listener.ReceivedArgs)
}

func TestEventQueueWrapper_ShouldQueue(t *testing.T) {
	listener := &TestEventQueueListenerWrapper{ShouldQueueValue: true}
	wrapper := &eventQueueWrapper{
		listener: listener,
		event:    "test.event",
	}

	assert.True(t, wrapper.ShouldQueue())

	listener.ShouldQueueValue = false
	assert.False(t, wrapper.ShouldQueue())
}

func TestEventQueueWrapper_Signature(t *testing.T) {
	listener := &TestEventQueueListenerWrapper{SignatureValue: "test.signature"}
	wrapper := &eventQueueWrapper{
		listener: listener,
		event:    "test.event",
	}

	assert.Equal(t, "test.signature", wrapper.Signature())
}

func TestEventToQueueListener(t *testing.T) {
	listener := &TestEventQueueListenerWrapper{
		ShouldQueueValue: true,
		SignatureValue:   "test.signature",
	}

	queueListener := eventToQueueListener(listener, "test.event")

	// Test that it implements the QueueListener interface
	assert.Implements(t, (*event.QueueListener)(nil), queueListener)

	// Test Handle method
	err := queueListener.Handle("arg1")
	assert.NoError(t, err)
	assert.True(t, listener.Called)
	assert.Equal(t, "test.event", listener.ReceivedEvent)
	assert.Equal(t, []any{"arg1"}, listener.ReceivedArgs)

	// Test ShouldQueue method
	assert.True(t, queueListener.ShouldQueue())

	// Test Signature method
	assert.Equal(t, "test.signature", queueListener.Signature())
}

func TestEventToQueueListener_MultipleArgs(t *testing.T) {
	listener := &TestEventQueueListenerWrapper{}
	queueListener := eventToQueueListener(listener, "complex.event")

	err := queueListener.Handle("arg1", 42, true, struct{ Name string }{Name: "test"})

	assert.NoError(t, err)
	assert.True(t, listener.Called)
	assert.Equal(t, "complex.event", listener.ReceivedEvent)
	assert.Len(t, listener.ReceivedArgs, 4)
	assert.Equal(t, "arg1", listener.ReceivedArgs[0])
	assert.Equal(t, 42, listener.ReceivedArgs[1])
	assert.Equal(t, true, listener.ReceivedArgs[2])
	assert.Equal(t, struct{ Name string }{Name: "test"}, listener.ReceivedArgs[3])
}
