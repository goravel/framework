// Code generated by mockery. DO NOT EDIT.

package http

import (
	http "github.com/goravel/framework/contracts/http"
	mock "github.com/stretchr/testify/mock"
)

// ResourceController is an autogenerated mock type for the ResourceController type
type ResourceController struct {
	mock.Mock
}

type ResourceController_Expecter struct {
	mock *mock.Mock
}

func (_m *ResourceController) EXPECT() *ResourceController_Expecter {
	return &ResourceController_Expecter{mock: &_m.Mock}
}

// Destroy provides a mock function with given fields: _a0
func (_m *ResourceController) Destroy(_a0 http.Context) http.Response {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Destroy")
	}

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(http.Context) http.Response); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// ResourceController_Destroy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Destroy'
type ResourceController_Destroy_Call struct {
	*mock.Call
}

// Destroy is a helper method to define mock.On call
//   - _a0 http.Context
func (_e *ResourceController_Expecter) Destroy(_a0 interface{}) *ResourceController_Destroy_Call {
	return &ResourceController_Destroy_Call{Call: _e.mock.On("Destroy", _a0)}
}

func (_c *ResourceController_Destroy_Call) Run(run func(_a0 http.Context)) *ResourceController_Destroy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.Context))
	})
	return _c
}

func (_c *ResourceController_Destroy_Call) Return(_a0 http.Response) *ResourceController_Destroy_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ResourceController_Destroy_Call) RunAndReturn(run func(http.Context) http.Response) *ResourceController_Destroy_Call {
	_c.Call.Return(run)
	return _c
}

// Index provides a mock function with given fields: _a0
func (_m *ResourceController) Index(_a0 http.Context) http.Response {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Index")
	}

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(http.Context) http.Response); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// ResourceController_Index_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Index'
type ResourceController_Index_Call struct {
	*mock.Call
}

// Index is a helper method to define mock.On call
//   - _a0 http.Context
func (_e *ResourceController_Expecter) Index(_a0 interface{}) *ResourceController_Index_Call {
	return &ResourceController_Index_Call{Call: _e.mock.On("Index", _a0)}
}

func (_c *ResourceController_Index_Call) Run(run func(_a0 http.Context)) *ResourceController_Index_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.Context))
	})
	return _c
}

func (_c *ResourceController_Index_Call) Return(_a0 http.Response) *ResourceController_Index_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ResourceController_Index_Call) RunAndReturn(run func(http.Context) http.Response) *ResourceController_Index_Call {
	_c.Call.Return(run)
	return _c
}

// Show provides a mock function with given fields: _a0
func (_m *ResourceController) Show(_a0 http.Context) http.Response {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Show")
	}

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(http.Context) http.Response); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// ResourceController_Show_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Show'
type ResourceController_Show_Call struct {
	*mock.Call
}

// Show is a helper method to define mock.On call
//   - _a0 http.Context
func (_e *ResourceController_Expecter) Show(_a0 interface{}) *ResourceController_Show_Call {
	return &ResourceController_Show_Call{Call: _e.mock.On("Show", _a0)}
}

func (_c *ResourceController_Show_Call) Run(run func(_a0 http.Context)) *ResourceController_Show_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.Context))
	})
	return _c
}

func (_c *ResourceController_Show_Call) Return(_a0 http.Response) *ResourceController_Show_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ResourceController_Show_Call) RunAndReturn(run func(http.Context) http.Response) *ResourceController_Show_Call {
	_c.Call.Return(run)
	return _c
}

// Store provides a mock function with given fields: _a0
func (_m *ResourceController) Store(_a0 http.Context) http.Response {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Store")
	}

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(http.Context) http.Response); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// ResourceController_Store_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Store'
type ResourceController_Store_Call struct {
	*mock.Call
}

// Store is a helper method to define mock.On call
//   - _a0 http.Context
func (_e *ResourceController_Expecter) Store(_a0 interface{}) *ResourceController_Store_Call {
	return &ResourceController_Store_Call{Call: _e.mock.On("Store", _a0)}
}

func (_c *ResourceController_Store_Call) Run(run func(_a0 http.Context)) *ResourceController_Store_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.Context))
	})
	return _c
}

func (_c *ResourceController_Store_Call) Return(_a0 http.Response) *ResourceController_Store_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ResourceController_Store_Call) RunAndReturn(run func(http.Context) http.Response) *ResourceController_Store_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: _a0
func (_m *ResourceController) Update(_a0 http.Context) http.Response {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(http.Context) http.Response); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	return r0
}

// ResourceController_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type ResourceController_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - _a0 http.Context
func (_e *ResourceController_Expecter) Update(_a0 interface{}) *ResourceController_Update_Call {
	return &ResourceController_Update_Call{Call: _e.mock.On("Update", _a0)}
}

func (_c *ResourceController_Update_Call) Run(run func(_a0 http.Context)) *ResourceController_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.Context))
	})
	return _c
}

func (_c *ResourceController_Update_Call) Return(_a0 http.Response) *ResourceController_Update_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ResourceController_Update_Call) RunAndReturn(run func(http.Context) http.Response) *ResourceController_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewResourceController creates a new instance of ResourceController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewResourceController(t interface {
	mock.TestingT
	Cleanup(func())
}) *ResourceController {
	mock := &ResourceController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
