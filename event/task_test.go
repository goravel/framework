package event

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/event"
	queuecontract "github.com/goravel/framework/contracts/queue"
	queuemock "github.com/goravel/framework/contracts/queue/mocks"
)

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

				mockQueue.On("Job", listener, []queuecontract.Arg{
					{Type: "string", Value: "test"},
				}).Return(mockTask).Once()
				mockTask.On("DispatchSync").Return(nil).Once()

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

				mockQueue.On("Job", listener, []queuecontract.Arg{
					{Type: "string", Value: "test"},
				}).Return(mockTask).Once()
				mockTask.On("DispatchSync").Return(errors.New("error")).Once()

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
