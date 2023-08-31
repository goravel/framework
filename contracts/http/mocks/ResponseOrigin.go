// Code generated by mockery v2.33.1. DO NOT EDIT.

package mocks

import (
	bytes "bytes"

	mock "github.com/stretchr/testify/mock"

	nethttp "net/http"
)

// ResponseOrigin is an autogenerated mock type for the ResponseOrigin type
type ResponseOrigin struct {
	mock.Mock
}

// Body provides a mock function with given fields:
func (_m *ResponseOrigin) Body() *bytes.Buffer {
	ret := _m.Called()

	var r0 *bytes.Buffer
	if rf, ok := ret.Get(0).(func() *bytes.Buffer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bytes.Buffer)
		}
	}

	return r0
}

// Header provides a mock function with given fields:
func (_m *ResponseOrigin) Header() nethttp.Header {
	ret := _m.Called()

	var r0 nethttp.Header
	if rf, ok := ret.Get(0).(func() nethttp.Header); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nethttp.Header)
		}
	}

	return r0
}

// Size provides a mock function with given fields:
func (_m *ResponseOrigin) Size() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Status provides a mock function with given fields:
func (_m *ResponseOrigin) Status() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// NewResponseOrigin creates a new instance of ResponseOrigin. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewResponseOrigin(t interface {
	mock.TestingT
	Cleanup(func())
}) *ResponseOrigin {
	mock := &ResponseOrigin{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
