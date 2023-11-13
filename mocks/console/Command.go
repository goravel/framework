// Code generated by mockery v2.34.2. DO NOT EDIT.

package mocks

import (
	console "github.com/goravel/framework/contracts/console"
	command "github.com/goravel/framework/contracts/console/command"

	mock "github.com/stretchr/testify/mock"
)

// Command is an autogenerated mock type for the Command type
type Command struct {
	mock.Mock
}

// Description provides a mock function with given fields:
func (_m *Command) Description() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Extend provides a mock function with given fields:
func (_m *Command) Extend() command.Extend {
	ret := _m.Called()

	var r0 command.Extend
	if rf, ok := ret.Get(0).(func() command.Extend); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(command.Extend)
	}

	return r0
}

// Handle provides a mock function with given fields: ctx
func (_m *Command) Handle(ctx console.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(console.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Signature provides a mock function with given fields:
func (_m *Command) Signature() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NewCommand creates a new instance of Command. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *Command {
	mock := &Command{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
