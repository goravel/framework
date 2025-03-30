// Code generated by mockery. DO NOT EDIT.

package event

import mock "github.com/stretchr/testify/mock"

// Task is an autogenerated mock type for the Task type
type Task struct {
	mock.Mock
}

type Task_Expecter struct {
	mock *mock.Mock
}

func (_m *Task) EXPECT() *Task_Expecter {
	return &Task_Expecter{mock: &_m.Mock}
}

// Dispatch provides a mock function with no fields
func (_m *Task) Dispatch() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Dispatch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Task_Dispatch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Dispatch'
type Task_Dispatch_Call struct {
	*mock.Call
}

// Dispatch is a helper method to define mock.On call
func (_e *Task_Expecter) Dispatch() *Task_Dispatch_Call {
	return &Task_Dispatch_Call{Call: _e.mock.On("Dispatch")}
}

func (_c *Task_Dispatch_Call) Run(run func()) *Task_Dispatch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Task_Dispatch_Call) Return(_a0 error) *Task_Dispatch_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Task_Dispatch_Call) RunAndReturn(run func() error) *Task_Dispatch_Call {
	_c.Call.Return(run)
	return _c
}

// NewTask creates a new instance of Task. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTask(t interface {
	mock.TestingT
	Cleanup(func())
}) *Task {
	mock := &Task{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
