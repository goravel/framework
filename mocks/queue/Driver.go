// Code generated by mockery v2.34.2. DO NOT EDIT.

package mocks

import (
	queue "github.com/goravel/framework/contracts/queue"
	mock "github.com/stretchr/testify/mock"
)

// Driver is an autogenerated mock type for the Driver type
type Driver struct {
	mock.Mock
}

// Bulk provides a mock function with given fields: jobs, _a1
func (_m *Driver) Bulk(jobs []queue.Jobs, _a1 string) error {
	ret := _m.Called(jobs, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func([]queue.Jobs, string) error); ok {
		r0 = rf(jobs, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Clear provides a mock function with given fields: _a0
func (_m *Driver) Clear(_a0 string) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ConnectionName provides a mock function with given fields:
func (_m *Driver) ConnectionName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Delete provides a mock function with given fields: _a0, job
func (_m *Driver) Delete(_a0 string, job queue.Job) error {
	ret := _m.Called(_a0, job)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, queue.Job) error); ok {
		r0 = rf(_a0, job)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Later provides a mock function with given fields: delay, job, args, _a3
func (_m *Driver) Later(delay int, job queue.Job, args []any, _a3 string) error {
	ret := _m.Called(delay, job, args, _a3)

	var r0 error
	if rf, ok := ret.Get(0).(func(int, queue.Job, []any, string) error); ok {
		r0 = rf(delay, job, args, _a3)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Pop provides a mock function with given fields: _a0
func (_m *Driver) Pop(_a0 string) (queue.Job, []any, error) {
	ret := _m.Called(_a0)

	var r0 queue.Job
	var r1 []any
	var r2 error
	if rf, ok := ret.Get(0).(func(string) (queue.Job, []any, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(string) queue.Job); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Job)
		}
	}

	if rf, ok := ret.Get(1).(func(string) []any); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]any)
		}
	}

	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(_a0)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Push provides a mock function with given fields: job, args, _a2
func (_m *Driver) Push(job queue.Job, args []any, _a2 string) error {
	ret := _m.Called(job, args, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(queue.Job, []any, string) error); ok {
		r0 = rf(job, args, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Release provides a mock function with given fields: _a0, job, delay
func (_m *Driver) Release(_a0 string, job queue.Job, delay int) error {
	ret := _m.Called(_a0, job, delay)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, queue.Job, int) error); ok {
		r0 = rf(_a0, job, delay)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Server provides a mock function with given fields: concurrent, _a1
func (_m *Driver) Server(concurrent int, _a1 string) {
	_m.Called(concurrent, _a1)
}

// Size provides a mock function with given fields: _a0
func (_m *Driver) Size(_a0 string) (int64, error) {
	ret := _m.Called(_a0)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (int64, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(string) int64); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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