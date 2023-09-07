// Code generated by mockery v2.33.1. DO NOT EDIT.

package mocks

import (
	http "github.com/goravel/framework/contracts/http"
	log "github.com/goravel/framework/contracts/log"

	mock "github.com/stretchr/testify/mock"
)

// Writer is an autogenerated mock type for the Writer type
type Writer struct {
	mock.Mock
}

// Code provides a mock function with given fields: code
func (_m *Writer) Code(code string) log.Writer {
	ret := _m.Called(code)

	var r0 log.Writer
	if rf, ok := ret.Get(0).(func(string) log.Writer); ok {
		r0 = rf(code)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Writer)
		}
	}

	return r0
}

// Debug provides a mock function with given fields: args
func (_m *Writer) Debug(args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Debugf provides a mock function with given fields: format, args
func (_m *Writer) Debugf(format string, args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, format)
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Error provides a mock function with given fields: args
func (_m *Writer) Error(args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Errorf provides a mock function with given fields: format, args
func (_m *Writer) Errorf(format string, args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, format)
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Fatal provides a mock function with given fields: args
func (_m *Writer) Fatal(args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Fatalf provides a mock function with given fields: format, args
func (_m *Writer) Fatalf(format string, args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, format)
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Hint provides a mock function with given fields: hint
func (_m *Writer) Hint(hint string) log.Writer {
	ret := _m.Called(hint)

	var r0 log.Writer
	if rf, ok := ret.Get(0).(func(string) log.Writer); ok {
		r0 = rf(hint)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Writer)
		}
	}

	return r0
}

// In provides a mock function with given fields: domain
func (_m *Writer) In(domain string) log.Writer {
	ret := _m.Called(domain)

	var r0 log.Writer
	if rf, ok := ret.Get(0).(func(string) log.Writer); ok {
		r0 = rf(domain)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Writer)
		}
	}

	return r0
}

// Info provides a mock function with given fields: args
func (_m *Writer) Info(args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Infof provides a mock function with given fields: format, args
func (_m *Writer) Infof(format string, args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, format)
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Owner provides a mock function with given fields: owner
func (_m *Writer) Owner(owner interface{}) log.Writer {
	ret := _m.Called(owner)

	var r0 log.Writer
	if rf, ok := ret.Get(0).(func(interface{}) log.Writer); ok {
		r0 = rf(owner)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Writer)
		}
	}

	return r0
}

// Panic provides a mock function with given fields: args
func (_m *Writer) Panic(args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Panicf provides a mock function with given fields: format, args
func (_m *Writer) Panicf(format string, args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, format)
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Request provides a mock function with given fields: req
func (_m *Writer) Request(req http.ContextRequest) log.Writer {
	ret := _m.Called(req)

	var r0 log.Writer
	if rf, ok := ret.Get(0).(func(http.ContextRequest) log.Writer); ok {
		r0 = rf(req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Writer)
		}
	}

	return r0
}

// Response provides a mock function with given fields: res
func (_m *Writer) Response(res http.ContextResponse) log.Writer {
	ret := _m.Called(res)

	var r0 log.Writer
	if rf, ok := ret.Get(0).(func(http.ContextResponse) log.Writer); ok {
		r0 = rf(res)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Writer)
		}
	}

	return r0
}

// Tags provides a mock function with given fields: tags
func (_m *Writer) Tags(tags ...string) log.Writer {
	_va := make([]interface{}, len(tags))
	for _i := range tags {
		_va[_i] = tags[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 log.Writer
	if rf, ok := ret.Get(0).(func(...string) log.Writer); ok {
		r0 = rf(tags...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Writer)
		}
	}

	return r0
}

// User provides a mock function with given fields: user
func (_m *Writer) User(user interface{}) log.Writer {
	ret := _m.Called(user)

	var r0 log.Writer
	if rf, ok := ret.Get(0).(func(interface{}) log.Writer); ok {
		r0 = rf(user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Writer)
		}
	}

	return r0
}

// Warning provides a mock function with given fields: args
func (_m *Writer) Warning(args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Warningf provides a mock function with given fields: format, args
func (_m *Writer) Warningf(format string, args ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, format)
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// With provides a mock function with given fields: data
func (_m *Writer) With(data map[string]interface{}) log.Writer {
	ret := _m.Called(data)

	var r0 log.Writer
	if rf, ok := ret.Get(0).(func(map[string]interface{}) log.Writer); ok {
		r0 = rf(data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Writer)
		}
	}

	return r0
}

// NewWriter creates a new instance of Writer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *Writer {
	mock := &Writer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
