package event

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

func (receiver *TestQueueListener) Handle(args ...any) error {
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
			name: "dispatch sync success",
			setup: func() {
				listener := &TestListener{}
				mockTask := &queuemock.Task{}

				mockQueue.EXPECT().Job(listener, []queue.Arg{
					{Type: "string", Value: "test"},
				}).Return(mockTask).Once()
				mockTask.EXPECT().DispatchSync().Return(nil).Once()

				task = NewTask(mockQueue, []event.Arg{
					{Type: "string", Value: "test"},
				}, &TestEvent{}, []event.Listener{
					listener,
				})
			},
			expectErr: false,
		},
		{
			name: "dispatch sync error",
			setup: func() {
				listener := &TestListenerHandleError{}
				mockTask := &queuemock.Task{}

				mockQueue.EXPECT().Job(listener, []queue.Arg{
					{Type: "string", Value: "test"},
				}).Return(mockTask).Once()
				mockTask.EXPECT().DispatchSync().Return(errors.New("error")).Once()

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

func TestDispatchWithQueue(t *testing.T) {
	mockQueue := queuemock.NewQueue(t)
	listener := &TestQueueListener{}
	mockTask := queuemock.NewTask(t)

	mockQueue.EXPECT().Job(listener, []queue.Arg{
		{Type: "string", Value: "test"},
	}).Return(mockTask).Once()
	mockTask.EXPECT().OnConnection("redis").Return(mockTask).Once()
	mockTask.EXPECT().OnQueue("emails").Return(mockTask).Once()
	mockTask.EXPECT().Dispatch().Return(nil).Once()

	task := NewTask(mockQueue, []event.Arg{
		{Type: "string", Value: "test"},
	}, &TestEvent{}, []event.Listener{
		listener,
	})
	assert.Nil(t, task.Dispatch())
}

func TestDispatchWithQueueError(t *testing.T) {
	mockQueue := queuemock.NewQueue(t)
	listener := &TestQueueListener{}
	mockTask := queuemock.NewTask(t)

	mockQueue.EXPECT().Job(listener, []queue.Arg{
		{Type: "string", Value: "test"},
	}).Return(mockTask).Once()
	mockTask.EXPECT().OnConnection("redis").Return(mockTask).Once()
	mockTask.EXPECT().OnQueue("emails").Return(mockTask).Once()
	mockTask.EXPECT().Dispatch().Return(errors.New("queue error")).Once()

	task := NewTask(mockQueue, []event.Arg{
		{Type: "string", Value: "test"},
	}, &TestEvent{}, []event.Listener{
		listener,
	})
	assert.EqualError(t, task.Dispatch(), "queue error")
}

func TestTestUtils(t *testing.T) {
	assert.Equal(t, "test_listener", (&TestListener{}).Signature())
	assert.Nil(t, (&TestListener{}).Handle())
	assert.Equal(t, "test_listener", (&TestListenerHandleError{}).Signature())
	assert.EqualError(t, (&TestListenerHandleError{}).Handle(), "error")
}
