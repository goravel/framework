// Code generated by mockery. DO NOT EDIT.

package http

import mock "github.com/stretchr/testify/mock"

// ResponseWithAbort is an autogenerated mock type for the ResponseWithAbort type
type ResponseWithAbort struct {
	mock.Mock
}

type ResponseWithAbort_Expecter struct {
	mock *mock.Mock
}

func (_m *ResponseWithAbort) EXPECT() *ResponseWithAbort_Expecter {
	return &ResponseWithAbort_Expecter{mock: &_m.Mock}
}

// Abort provides a mock function with no fields
func (_m *ResponseWithAbort) Abort() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Abort")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResponseWithAbort_Abort_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Abort'
type ResponseWithAbort_Abort_Call struct {
	*mock.Call
}

// Abort is a helper method to define mock.On call
func (_e *ResponseWithAbort_Expecter) Abort() *ResponseWithAbort_Abort_Call {
	return &ResponseWithAbort_Abort_Call{Call: _e.mock.On("Abort")}
}

func (_c *ResponseWithAbort_Abort_Call) Run(run func()) *ResponseWithAbort_Abort_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ResponseWithAbort_Abort_Call) Return(_a0 error) *ResponseWithAbort_Abort_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ResponseWithAbort_Abort_Call) RunAndReturn(run func() error) *ResponseWithAbort_Abort_Call {
	_c.Call.Return(run)
	return _c
}

// Render provides a mock function with no fields
func (_m *ResponseWithAbort) Render() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Render")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResponseWithAbort_Render_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Render'
type ResponseWithAbort_Render_Call struct {
	*mock.Call
}

// Render is a helper method to define mock.On call
func (_e *ResponseWithAbort_Expecter) Render() *ResponseWithAbort_Render_Call {
	return &ResponseWithAbort_Render_Call{Call: _e.mock.On("Render")}
}

func (_c *ResponseWithAbort_Render_Call) Run(run func()) *ResponseWithAbort_Render_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ResponseWithAbort_Render_Call) Return(_a0 error) *ResponseWithAbort_Render_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ResponseWithAbort_Render_Call) RunAndReturn(run func() error) *ResponseWithAbort_Render_Call {
	_c.Call.Return(run)
	return _c
}

// NewResponseWithAbort creates a new instance of ResponseWithAbort. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewResponseWithAbort(t interface {
	mock.TestingT
	Cleanup(func())
}) *ResponseWithAbort {
	mock := &ResponseWithAbort{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
