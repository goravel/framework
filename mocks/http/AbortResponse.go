// Code generated by mockery. DO NOT EDIT.

package http

import mock "github.com/stretchr/testify/mock"

// AbortResponse is an autogenerated mock type for the AbortResponse type
type AbortResponse struct {
	mock.Mock
}

type AbortResponse_Expecter struct {
	mock *mock.Mock
}

func (_m *AbortResponse) EXPECT() *AbortResponse_Expecter {
	return &AbortResponse_Expecter{mock: &_m.Mock}
}

// Abort provides a mock function with given fields:
func (_m *AbortResponse) Abort() error {
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

// AbortResponse_Abort_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Abort'
type AbortResponse_Abort_Call struct {
	*mock.Call
}

// Abort is a helper method to define mock.On call
func (_e *AbortResponse_Expecter) Abort() *AbortResponse_Abort_Call {
	return &AbortResponse_Abort_Call{Call: _e.mock.On("Abort")}
}

func (_c *AbortResponse_Abort_Call) Run(run func()) *AbortResponse_Abort_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AbortResponse_Abort_Call) Return(_a0 error) *AbortResponse_Abort_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AbortResponse_Abort_Call) RunAndReturn(run func() error) *AbortResponse_Abort_Call {
	_c.Call.Return(run)
	return _c
}

// Render provides a mock function with given fields:
func (_m *AbortResponse) Render() error {
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

// AbortResponse_Render_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Render'
type AbortResponse_Render_Call struct {
	*mock.Call
}

// Render is a helper method to define mock.On call
func (_e *AbortResponse_Expecter) Render() *AbortResponse_Render_Call {
	return &AbortResponse_Render_Call{Call: _e.mock.On("Render")}
}

func (_c *AbortResponse_Render_Call) Run(run func()) *AbortResponse_Render_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AbortResponse_Render_Call) Return(_a0 error) *AbortResponse_Render_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AbortResponse_Render_Call) RunAndReturn(run func() error) *AbortResponse_Render_Call {
	_c.Call.Return(run)
	return _c
}

// NewAbortResponse creates a new instance of AbortResponse. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAbortResponse(t interface {
	mock.TestingT
	Cleanup(func())
}) *AbortResponse {
	mock := &AbortResponse{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
