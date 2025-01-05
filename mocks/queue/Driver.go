// Code generated by mockery. DO NOT EDIT.

package queue

import (
	queue "github.com/goravel/framework/contracts/queue"
	mock "github.com/stretchr/testify/mock"

	time "time"
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

// Bulk provides a mock function with given fields: jobs, _a1
func (_m *Driver) Bulk(jobs []queue.Jobs, _a1 string) error {
	ret := _m.Called(jobs, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Bulk")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]queue.Jobs, string) error); ok {
		r0 = rf(jobs, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Driver_Bulk_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bulk'
type Driver_Bulk_Call struct {
	*mock.Call
}

// Bulk is a helper method to define mock.On call
//   - jobs []queue.Jobs
//   - _a1 string
func (_e *Driver_Expecter) Bulk(jobs interface{}, _a1 interface{}) *Driver_Bulk_Call {
	return &Driver_Bulk_Call{Call: _e.mock.On("Bulk", jobs, _a1)}
}

func (_c *Driver_Bulk_Call) Run(run func(jobs []queue.Jobs, _a1 string)) *Driver_Bulk_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]queue.Jobs), args[1].(string))
	})
	return _c
}

func (_c *Driver_Bulk_Call) Return(_a0 error) *Driver_Bulk_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Bulk_Call) RunAndReturn(run func([]queue.Jobs, string) error) *Driver_Bulk_Call {
	_c.Call.Return(run)
	return _c
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

// Later provides a mock function with given fields: delay, job, args, _a3
func (_m *Driver) Later(delay time.Duration, job queue.Job, args []interface{}, _a3 string) error {
	ret := _m.Called(delay, job, args, _a3)

	if len(ret) == 0 {
		panic("no return value specified for Later")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(time.Duration, queue.Job, []interface{}, string) error); ok {
		r0 = rf(delay, job, args, _a3)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Driver_Later_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Later'
type Driver_Later_Call struct {
	*mock.Call
}

// Later is a helper method to define mock.On call
//   - delay time.Duration
//   - job queue.Job
//   - args []interface{}
//   - _a3 string
func (_e *Driver_Expecter) Later(delay interface{}, job interface{}, args interface{}, _a3 interface{}) *Driver_Later_Call {
	return &Driver_Later_Call{Call: _e.mock.On("Later", delay, job, args, _a3)}
}

func (_c *Driver_Later_Call) Run(run func(delay time.Duration, job queue.Job, args []interface{}, _a3 string)) *Driver_Later_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Duration), args[1].(queue.Job), args[2].([]interface{}), args[3].(string))
	})
	return _c
}

func (_c *Driver_Later_Call) Return(_a0 error) *Driver_Later_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Later_Call) RunAndReturn(run func(time.Duration, queue.Job, []interface{}, string) error) *Driver_Later_Call {
	_c.Call.Return(run)
	return _c
}

// Pop provides a mock function with given fields: _a0
func (_m *Driver) Pop(_a0 string) (queue.Job, []interface{}, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Pop")
	}

	var r0 queue.Job
	var r1 []interface{}
	var r2 error
	if rf, ok := ret.Get(0).(func(string) (queue.Job, []interface{}, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(string) queue.Job); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Job)
		}
	}

	if rf, ok := ret.Get(1).(func(string) []interface{}); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]interface{})
		}
	}

	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(_a0)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
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

func (_c *Driver_Pop_Call) Return(_a0 queue.Job, _a1 []interface{}, _a2 error) *Driver_Pop_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *Driver_Pop_Call) RunAndReturn(run func(string) (queue.Job, []interface{}, error)) *Driver_Pop_Call {
	_c.Call.Return(run)
	return _c
}

// Push provides a mock function with given fields: job, args, _a2
func (_m *Driver) Push(job queue.Job, args []interface{}, _a2 string) error {
	ret := _m.Called(job, args, _a2)

	if len(ret) == 0 {
		panic("no return value specified for Push")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(queue.Job, []interface{}, string) error); ok {
		r0 = rf(job, args, _a2)
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
//   - job queue.Job
//   - args []interface{}
//   - _a2 string
func (_e *Driver_Expecter) Push(job interface{}, args interface{}, _a2 interface{}) *Driver_Push_Call {
	return &Driver_Push_Call{Call: _e.mock.On("Push", job, args, _a2)}
}

func (_c *Driver_Push_Call) Run(run func(job queue.Job, args []interface{}, _a2 string)) *Driver_Push_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(queue.Job), args[1].([]interface{}), args[2].(string))
	})
	return _c
}

func (_c *Driver_Push_Call) Return(_a0 error) *Driver_Push_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Push_Call) RunAndReturn(run func(queue.Job, []interface{}, string) error) *Driver_Push_Call {
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
