package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	queuemock "github.com/goravel/framework/mocks/queue"
)

type TestQueueListener struct{}

func (receiver *TestQueueListener) Signature() string {
	return "test_queue_listener"
}

func (receiver *TestQueueListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     true,
		Connection: "redis",
		Queue:      "emails",
	}
}

func (receiver *TestQueueListener) Handle(eventName string, args ...any) error {
	return nil
}

func TestDispatch(t *testing.T) {
	var (
		mockQueue *queuemock.Queue
		task      *Task
	)

	beforeEach := func() {
		mockQueue = &queuemock.Queue{}
	}

	tests := []struct {
		name      string
		setup     func()
		expectErr bool
	}{
		{
			// A listener that doesn't enable queueing now runs in process, the
			// queue is not involved at all.
			name: "dispatch sync success",
			setup: func() {
				task = NewTask(mockQueue, []event.Arg{
					{Type: "string", Value: "test"},
				}, &TestEvent{}, []event.Listener{
					&TestListener{},
				})
			},
			expectErr: false,
		},
		{
			name: "dispatch sync error",
			setup: func() {
				task = NewTask(mockQueue, []event.Arg{
					{Type: "string", Value: "test"},
				}, &TestEvent{}, []event.Listener{
					&TestListenerHandleError{},
				})
			},
			expectErr: true,
		},
		{
			name: "no listeners",
			setup: func() {
				task = NewTask(mockQueue, []event.Arg{
					{Type: "string", Value: "test"},
				}, &TestEventNoRegister{}, nil)
			},
			expectErr: true,
		},
		{
			name: "event handle return error",
			setup: func() {
				task = NewTask(mockQueue, []event.Arg{
					{Type: "string", Value: "test"},
				}, &TestEventHandleError{}, []event.Listener{
					&TestListener{},
				})
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			err := task.Dispatch()
			assert.Equal(t, test.expectErr, err != nil, test.name)
			mockQueue.AssertExpectations(t)
		})
	}
}

func TestDispatchStopsAtTheFirstError(t *testing.T) {
	mockQueue := queuemock.NewQueue(t)

	// The failing listener must stop the one behind it, the deprecated Task has
	// always been fail fast.
	task := NewTask(mockQueue, nil, &TestEvent{}, []event.Listener{
		&TestListenerHandleError{},
		&TestQueueListener{},
	})

	// No queue expectation: TestQueueListener would have been queued, it is never
	// reached.
	assert.EqualError(t, task.Dispatch(), "error")
}

func TestDispatchWithQueue(t *testing.T) {
	mockQueue := queuemock.NewQueue(t)
	mockPendingJob := queuemock.NewPendingJob(t)

	mockQueue.EXPECT().Register(mock.Anything).Maybe()
	mockQueue.EXPECT().Job(mock.Anything, []queue.Arg{
		{Type: "string", Value: "event.TestEvent"},
		{Type: "string", Value: "test"},
	}).Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().OnConnection("redis").Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().OnQueue("emails").Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().Dispatch().Return(nil).Once()

	task := NewTask(mockQueue, []event.Arg{
		{Type: "string", Value: "test"},
	}, &TestEvent{}, []event.Listener{
		&TestQueueListener{},
	})
	assert.Nil(t, task.Dispatch())
}

func TestDispatchWithQueueError(t *testing.T) {
	mockQueue := queuemock.NewQueue(t)
	mockPendingJob := queuemock.NewPendingJob(t)

	mockQueue.EXPECT().Job(mock.Anything, mock.Anything).Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().OnConnection("redis").Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().OnQueue("emails").Return(mockPendingJob).Once()
	mockPendingJob.EXPECT().Dispatch().Return(errors.New("queue error")).Once()

	task := NewTask(mockQueue, []event.Arg{
		{Type: "string", Value: "test"},
	}, &TestEvent{}, []event.Listener{
		&TestQueueListener{},
	})
	assert.EqualError(t, task.Dispatch(), "queue error")
}

func TestTestUtils(t *testing.T) {
	assert.Equal(t, "test_listener", (&TestListener{}).Signature())
	assert.Nil(t, (&TestListener{}).Handle("event.TestEvent"))
	assert.Equal(t, "test_listener", (&TestListenerHandleError{}).Signature())
	assert.EqualError(t, (&TestListenerHandleError{}).Handle("event.TestEvent"), "error")
}
