// Code generated by mockery v2.30.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Loader is an autogenerated mock type for the Loader type
type Loader struct {
	mock.Mock
}

// Load provides a mock function with given fields: folder, locale
func (_m *Loader) Load(folder string, locale string) (map[string]interface{}, error) {
	ret := _m.Called(folder, locale)

	var r0 map[string]interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (map[string]interface{}, error)); ok {
		return rf(folder, locale)
	}
	if rf, ok := ret.Get(0).(func(string, string) map[string]interface{}); ok {
		r0 = rf(folder, locale)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(folder, locale)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewLoader creates a new instance of Loader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLoader(t interface {
	mock.TestingT
	Cleanup(func())
}) *Loader {
	mock := &Loader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
