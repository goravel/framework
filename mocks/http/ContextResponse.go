// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	http "github.com/goravel/framework/contracts/http"
	mock "github.com/stretchr/testify/mock"

	nethttp "net/http"
)

// ContextResponse is an autogenerated mock type for the ContextResponse type
type ContextResponse struct {
	mock.Mock
}

// Data provides a mock function with given fields: code, contentType, data
func (_m *ContextResponse) Data(code int, contentType string, data []byte) http.Response {
	ret := _m.Called(code, contentType, data)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(int, string, []byte) http.Response); ok {
		r0 = rf(code, contentType, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// Download provides a mock function with given fields: filepath, filename
func (_m *ContextResponse) Download(filepath string, filename string) http.Response {
	ret := _m.Called(filepath, filename)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(string, string) http.Response); ok {
		r0 = rf(filepath, filename)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// File provides a mock function with given fields: filepath
func (_m *ContextResponse) File(filepath string) http.Response {
	ret := _m.Called(filepath)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(string) http.Response); ok {
		r0 = rf(filepath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// Flush provides a mock function with given fields:
func (_m *ContextResponse) Flush() {
	_m.Called()
}

// Header provides a mock function with given fields: key, value
func (_m *ContextResponse) Header(key string, value string) http.ContextResponse {
	ret := _m.Called(key, value)

	var r0 http.ContextResponse
	if rf, ok := ret.Get(0).(func(string, string) http.ContextResponse); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.ContextResponse)
		}
	}

	return r0
}

// Json provides a mock function with given fields: code, obj
func (_m *ContextResponse) Json(code int, obj interface{}) http.Response {
	ret := _m.Called(code, obj)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(int, interface{}) http.Response); ok {
		r0 = rf(code, obj)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// Origin provides a mock function with given fields:
func (_m *ContextResponse) Origin() http.ResponseOrigin {
	ret := _m.Called()

	var r0 http.ResponseOrigin
	if rf, ok := ret.Get(0).(func() http.ResponseOrigin); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.ResponseOrigin)
		}
	}

	return r0
}

// Redirect provides a mock function with given fields: code, location
func (_m *ContextResponse) Redirect(code int, location string) http.Response {
	ret := _m.Called(code, location)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(int, string) http.Response); ok {
		r0 = rf(code, location)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// Status provides a mock function with given fields: code
func (_m *ContextResponse) Status(code int) http.ResponseStatus {
	ret := _m.Called(code)

	var r0 http.ResponseStatus
	if rf, ok := ret.Get(0).(func(int) http.ResponseStatus); ok {
		r0 = rf(code)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.ResponseStatus)
		}
	}

	return r0
}

// String provides a mock function with given fields: code, format, values
func (_m *ContextResponse) String(code int, format string, values ...interface{}) http.Response {
	var _ca []interface{}
	_ca = append(_ca, code, format)
	_ca = append(_ca, values...)
	ret := _m.Called(_ca...)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(int, string, ...interface{}) http.Response); ok {
		r0 = rf(code, format, values...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// Success provides a mock function with given fields:
func (_m *ContextResponse) Success() http.ResponseSuccess {
	ret := _m.Called()

	var r0 http.ResponseSuccess
	if rf, ok := ret.Get(0).(func() http.ResponseSuccess); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.ResponseSuccess)
		}
	}

	return r0
}

// View provides a mock function with given fields:
func (_m *ContextResponse) View() http.ResponseView {
	ret := _m.Called()

	var r0 http.ResponseView
	if rf, ok := ret.Get(0).(func() http.ResponseView); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.ResponseView)
		}
	}

	return r0
}

// Writer provides a mock function with given fields:
func (_m *ContextResponse) Writer() nethttp.ResponseWriter {
	ret := _m.Called()

	var r0 nethttp.ResponseWriter
	if rf, ok := ret.Get(0).(func() nethttp.ResponseWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nethttp.ResponseWriter)
		}
	}

	return r0
}
