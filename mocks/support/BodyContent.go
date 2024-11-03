// Code generated by mockery. DO NOT EDIT.

package support

import (
	io "io"

	mock "github.com/stretchr/testify/mock"
)

// BodyContent is an autogenerated mock type for the BodyContent type
type BodyContent struct {
	mock.Mock
}

type BodyContent_Expecter struct {
	mock *mock.Mock
}

func (_m *BodyContent) EXPECT() *BodyContent_Expecter {
	return &BodyContent_Expecter{mock: &_m.Mock}
}

// ContentType provides a mock function with given fields:
func (_m *BodyContent) ContentType() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ContentType")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// BodyContent_ContentType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ContentType'
type BodyContent_ContentType_Call struct {
	*mock.Call
}

// ContentType is a helper method to define mock.On call
func (_e *BodyContent_Expecter) ContentType() *BodyContent_ContentType_Call {
	return &BodyContent_ContentType_Call{Call: _e.mock.On("ContentType")}
}

func (_c *BodyContent_ContentType_Call) Run(run func()) *BodyContent_ContentType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BodyContent_ContentType_Call) Return(_a0 string) *BodyContent_ContentType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BodyContent_ContentType_Call) RunAndReturn(run func() string) *BodyContent_ContentType_Call {
	_c.Call.Return(run)
	return _c
}

// Reader provides a mock function with given fields:
func (_m *BodyContent) Reader() io.Reader {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Reader")
	}

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func() io.Reader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	return r0
}

// BodyContent_Reader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reader'
type BodyContent_Reader_Call struct {
	*mock.Call
}

// Reader is a helper method to define mock.On call
func (_e *BodyContent_Expecter) Reader() *BodyContent_Reader_Call {
	return &BodyContent_Reader_Call{Call: _e.mock.On("Reader")}
}

func (_c *BodyContent_Reader_Call) Run(run func()) *BodyContent_Reader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BodyContent_Reader_Call) Return(_a0 io.Reader) *BodyContent_Reader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BodyContent_Reader_Call) RunAndReturn(run func() io.Reader) *BodyContent_Reader_Call {
	_c.Call.Return(run)
	return _c
}

// NewBodyContent creates a new instance of BodyContent. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBodyContent(t interface {
	mock.TestingT
	Cleanup(func())
}) *BodyContent {
	mock := &BodyContent{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
