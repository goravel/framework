// Code generated by mockery v2.34.2. DO NOT EDIT.

package mocks

import (
	http "github.com/goravel/framework/contracts/http"
	mock "github.com/stretchr/testify/mock"
)

// ResponseSuccess is an autogenerated mock type for the ResponseSuccess type
type ResponseSuccess struct {
	mock.Mock
}

// Data provides a mock function with given fields: contentType, data
func (_m *ResponseSuccess) Data(contentType string, data []byte) http.Response {
	ret := _m.Called(contentType, data)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(string, []byte) http.Response); ok {
		r0 = rf(contentType, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// Json provides a mock function with given fields: obj
func (_m *ResponseSuccess) Json(obj interface{}) http.Response {
	ret := _m.Called(obj)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(interface{}) http.Response); ok {
		r0 = rf(obj)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// String provides a mock function with given fields: format, values
func (_m *ResponseSuccess) String(format string, values ...interface{}) http.Response {
	var _ca []interface{}
	_ca = append(_ca, format)
	_ca = append(_ca, values...)
	ret := _m.Called(_ca...)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(string, ...interface{}) http.Response); ok {
		r0 = rf(format, values...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// NewResponseSuccess creates a new instance of ResponseSuccess. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewResponseSuccess(t interface {
	mock.TestingT
	Cleanup(func())
}) *ResponseSuccess {
	mock := &ResponseSuccess{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
