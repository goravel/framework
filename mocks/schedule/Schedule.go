// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	schedule "github.com/goravel/framework/contracts/schedule"
	mock "github.com/stretchr/testify/mock"
)

// Schedule is an autogenerated mock type for the Schedule type
type Schedule struct {
	mock.Mock
}

// Call provides a mock function with given fields: callback
func (_m *Schedule) Call(callback func()) schedule.Event {
	ret := _m.Called(callback)

	var r0 schedule.Event
	if rf, ok := ret.Get(0).(func(func()) schedule.Event); ok {
		r0 = rf(callback)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(schedule.Event)
		}
	}

	return r0
}

// Command provides a mock function with given fields: command
func (_m *Schedule) Command(command string) schedule.Event {
	ret := _m.Called(command)

	var r0 schedule.Event
	if rf, ok := ret.Get(0).(func(string) schedule.Event); ok {
		r0 = rf(command)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(schedule.Event)
		}
	}

	return r0
}

// Register provides a mock function with given fields: events
func (_m *Schedule) Register(events []schedule.Event) {
	_m.Called(events)
}

// Run provides a mock function with given fields:
func (_m *Schedule) Run() {
	_m.Called()
}
