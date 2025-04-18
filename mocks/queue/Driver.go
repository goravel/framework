// Code generated by mockery. DO NOT EDIT.

package queue

import (
	queue "github.com/goravel/framework/contracts/queue"
	mock "github.com/stretchr/testify/mock"
)

// Driver is an autogenerated mock type for the Driver type
type Driver struct {
	mock.Mock
}

type Driver_Expecter struct {
	mock *mock.Mock
}

func (_m *Driver) EXPECT() *Driver_Expecter {
	return &Driver_Expecter{mock: &_m.Mock}
}

// Connection provides a mock function with no fields
func (_m *Driver) Connection() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Connection")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Driver_Connection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Connection'
type Driver_Connection_Call struct {
	*mock.Call
}

// Connection is a helper method to define mock.On call
func (_e *Driver_Expecter) Connection() *Driver_Connection_Call {
	return &Driver_Connection_Call{Call: _e.mock.On("Connection")}
}

func (_c *Driver_Connection_Call) Run(run func()) *Driver_Connection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_Connection_Call) Return(_a0 string) *Driver_Connection_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Connection_Call) RunAndReturn(run func() string) *Driver_Connection_Call {
	_c.Call.Return(run)
	return _c
}

// Driver provides a mock function with no fields
func (_m *Driver) Driver() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Driver")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Driver_Driver_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Driver'
type Driver_Driver_Call struct {
	*mock.Call
}

// Driver is a helper method to define mock.On call
func (_e *Driver_Expecter) Driver() *Driver_Driver_Call {
	return &Driver_Driver_Call{Call: _e.mock.On("Driver")}
}

func (_c *Driver_Driver_Call) Run(run func()) *Driver_Driver_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_Driver_Call) Return(_a0 string) *Driver_Driver_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Driver_Call) RunAndReturn(run func() string) *Driver_Driver_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with no fields
func (_m *Driver) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Driver_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type Driver_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *Driver_Expecter) Name() *Driver_Name_Call {
	return &Driver_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *Driver_Name_Call) Run(run func()) *Driver_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_Name_Call) Return(_a0 string) *Driver_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Name_Call) RunAndReturn(run func() string) *Driver_Name_Call {
	_c.Call.Return(run)
	return _c
}

// Pop provides a mock function with given fields: _a0
func (_m *Driver) Pop(_a0 string) (queue.Task, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Pop")
	}

	var r0 queue.Task
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (queue.Task, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(string) queue.Task); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(queue.Task)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Driver_Pop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Pop'
type Driver_Pop_Call struct {
	*mock.Call
}

// Pop is a helper method to define mock.On call
//   - _a0 string
func (_e *Driver_Expecter) Pop(_a0 interface{}) *Driver_Pop_Call {
	return &Driver_Pop_Call{Call: _e.mock.On("Pop", _a0)}
}

func (_c *Driver_Pop_Call) Run(run func(_a0 string)) *Driver_Pop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Driver_Pop_Call) Return(_a0 queue.Task, _a1 error) *Driver_Pop_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Driver_Pop_Call) RunAndReturn(run func(string) (queue.Task, error)) *Driver_Pop_Call {
	_c.Call.Return(run)
	return _c
}

// Push provides a mock function with given fields: task, _a1
func (_m *Driver) Push(task queue.Task, _a1 string) error {
	ret := _m.Called(task, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Push")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(queue.Task, string) error); ok {
		r0 = rf(task, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Driver_Push_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Push'
type Driver_Push_Call struct {
	*mock.Call
}

// Push is a helper method to define mock.On call
//   - task queue.Task
//   - _a1 string
func (_e *Driver_Expecter) Push(task interface{}, _a1 interface{}) *Driver_Push_Call {
	return &Driver_Push_Call{Call: _e.mock.On("Push", task, _a1)}
}

func (_c *Driver_Push_Call) Run(run func(task queue.Task, _a1 string)) *Driver_Push_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(queue.Task), args[1].(string))
	})
	return _c
}

func (_c *Driver_Push_Call) Return(_a0 error) *Driver_Push_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Push_Call) RunAndReturn(run func(queue.Task, string) error) *Driver_Push_Call {
	_c.Call.Return(run)
	return _c
}

// NewDriver creates a new instance of Driver. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDriver(t interface {
	mock.TestingT
	Cleanup(func())
}) *Driver {
	mock := &Driver{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
