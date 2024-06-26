// Code generated by mockery. DO NOT EDIT.

package mail

import (
	mail "github.com/goravel/framework/contracts/mail"
	mock "github.com/stretchr/testify/mock"
)

// ShouldQueue is an autogenerated mock type for the ShouldQueue type
type ShouldQueue struct {
	mock.Mock
}

type ShouldQueue_Expecter struct {
	mock *mock.Mock
}

func (_m *ShouldQueue) EXPECT() *ShouldQueue_Expecter {
	return &ShouldQueue_Expecter{mock: &_m.Mock}
}

// Attachments provides a mock function with given fields:
func (_m *ShouldQueue) Attachments() []string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Attachments")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// ShouldQueue_Attachments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Attachments'
type ShouldQueue_Attachments_Call struct {
	*mock.Call
}

// Attachments is a helper method to define mock.On call
func (_e *ShouldQueue_Expecter) Attachments() *ShouldQueue_Attachments_Call {
	return &ShouldQueue_Attachments_Call{Call: _e.mock.On("Attachments")}
}

func (_c *ShouldQueue_Attachments_Call) Run(run func()) *ShouldQueue_Attachments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ShouldQueue_Attachments_Call) Return(_a0 []string) *ShouldQueue_Attachments_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ShouldQueue_Attachments_Call) RunAndReturn(run func() []string) *ShouldQueue_Attachments_Call {
	_c.Call.Return(run)
	return _c
}

// Content provides a mock function with given fields:
func (_m *ShouldQueue) Content() *mail.Content {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Content")
	}

	var r0 *mail.Content
	if rf, ok := ret.Get(0).(func() *mail.Content); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mail.Content)
		}
	}

	return r0
}

// ShouldQueue_Content_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Content'
type ShouldQueue_Content_Call struct {
	*mock.Call
}

// Content is a helper method to define mock.On call
func (_e *ShouldQueue_Expecter) Content() *ShouldQueue_Content_Call {
	return &ShouldQueue_Content_Call{Call: _e.mock.On("Content")}
}

func (_c *ShouldQueue_Content_Call) Run(run func()) *ShouldQueue_Content_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ShouldQueue_Content_Call) Return(_a0 *mail.Content) *ShouldQueue_Content_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ShouldQueue_Content_Call) RunAndReturn(run func() *mail.Content) *ShouldQueue_Content_Call {
	_c.Call.Return(run)
	return _c
}

// Envelope provides a mock function with given fields:
func (_m *ShouldQueue) Envelope() *mail.Envelope {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Envelope")
	}

	var r0 *mail.Envelope
	if rf, ok := ret.Get(0).(func() *mail.Envelope); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mail.Envelope)
		}
	}

	return r0
}

// ShouldQueue_Envelope_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Envelope'
type ShouldQueue_Envelope_Call struct {
	*mock.Call
}

// Envelope is a helper method to define mock.On call
func (_e *ShouldQueue_Expecter) Envelope() *ShouldQueue_Envelope_Call {
	return &ShouldQueue_Envelope_Call{Call: _e.mock.On("Envelope")}
}

func (_c *ShouldQueue_Envelope_Call) Run(run func()) *ShouldQueue_Envelope_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ShouldQueue_Envelope_Call) Return(_a0 *mail.Envelope) *ShouldQueue_Envelope_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ShouldQueue_Envelope_Call) RunAndReturn(run func() *mail.Envelope) *ShouldQueue_Envelope_Call {
	_c.Call.Return(run)
	return _c
}

// Queue provides a mock function with given fields:
func (_m *ShouldQueue) Queue() *mail.Queue {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Queue")
	}

	var r0 *mail.Queue
	if rf, ok := ret.Get(0).(func() *mail.Queue); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mail.Queue)
		}
	}

	return r0
}

// ShouldQueue_Queue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Queue'
type ShouldQueue_Queue_Call struct {
	*mock.Call
}

// Queue is a helper method to define mock.On call
func (_e *ShouldQueue_Expecter) Queue() *ShouldQueue_Queue_Call {
	return &ShouldQueue_Queue_Call{Call: _e.mock.On("Queue")}
}

func (_c *ShouldQueue_Queue_Call) Run(run func()) *ShouldQueue_Queue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ShouldQueue_Queue_Call) Return(_a0 *mail.Queue) *ShouldQueue_Queue_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ShouldQueue_Queue_Call) RunAndReturn(run func() *mail.Queue) *ShouldQueue_Queue_Call {
	_c.Call.Return(run)
	return _c
}

// NewShouldQueue creates a new instance of ShouldQueue. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewShouldQueue(t interface {
	mock.TestingT
	Cleanup(func())
}) *ShouldQueue {
	mock := &ShouldQueue{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
