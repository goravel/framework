// Code generated by mockery. DO NOT EDIT.

package console

import mock "github.com/stretchr/testify/mock"

// Context is an autogenerated mock type for the Context type
type Context struct {
	mock.Mock
}

type Context_Expecter struct {
	mock *mock.Mock
}

func (_m *Context) EXPECT() *Context_Expecter {
	return &Context_Expecter{mock: &_m.Mock}
}

// Argument provides a mock function with given fields: index
func (_m *Context) Argument(index int) string {
	ret := _m.Called(index)

	if len(ret) == 0 {
		panic("no return value specified for Argument")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(int) string); ok {
		r0 = rf(index)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Context_Argument_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Argument'
type Context_Argument_Call struct {
	*mock.Call
}

// Argument is a helper method to define mock.On call
//   - index int
func (_e *Context_Expecter) Argument(index interface{}) *Context_Argument_Call {
	return &Context_Argument_Call{Call: _e.mock.On("Argument", index)}
}

func (_c *Context_Argument_Call) Run(run func(index int)) *Context_Argument_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *Context_Argument_Call) Return(_a0 string) *Context_Argument_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_Argument_Call) RunAndReturn(run func(int) string) *Context_Argument_Call {
	_c.Call.Return(run)
	return _c
}

// Arguments provides a mock function with given fields:
func (_m *Context) Arguments() []string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Arguments")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Context_Arguments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Arguments'
type Context_Arguments_Call struct {
	*mock.Call
}

// Arguments is a helper method to define mock.On call
func (_e *Context_Expecter) Arguments() *Context_Arguments_Call {
	return &Context_Arguments_Call{Call: _e.mock.On("Arguments")}
}

func (_c *Context_Arguments_Call) Run(run func()) *Context_Arguments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Context_Arguments_Call) Return(_a0 []string) *Context_Arguments_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_Arguments_Call) RunAndReturn(run func() []string) *Context_Arguments_Call {
	_c.Call.Return(run)
	return _c
}

// Option provides a mock function with given fields: key
func (_m *Context) Option(key string) string {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Option")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Context_Option_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Option'
type Context_Option_Call struct {
	*mock.Call
}

// Option is a helper method to define mock.On call
//   - key string
func (_e *Context_Expecter) Option(key interface{}) *Context_Option_Call {
	return &Context_Option_Call{Call: _e.mock.On("Option", key)}
}

func (_c *Context_Option_Call) Run(run func(key string)) *Context_Option_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Context_Option_Call) Return(_a0 string) *Context_Option_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_Option_Call) RunAndReturn(run func(string) string) *Context_Option_Call {
	_c.Call.Return(run)
	return _c
}

// OptionBool provides a mock function with given fields: key
func (_m *Context) OptionBool(key string) bool {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for OptionBool")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Context_OptionBool_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OptionBool'
type Context_OptionBool_Call struct {
	*mock.Call
}

// OptionBool is a helper method to define mock.On call
//   - key string
func (_e *Context_Expecter) OptionBool(key interface{}) *Context_OptionBool_Call {
	return &Context_OptionBool_Call{Call: _e.mock.On("OptionBool", key)}
}

func (_c *Context_OptionBool_Call) Run(run func(key string)) *Context_OptionBool_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Context_OptionBool_Call) Return(_a0 bool) *Context_OptionBool_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_OptionBool_Call) RunAndReturn(run func(string) bool) *Context_OptionBool_Call {
	_c.Call.Return(run)
	return _c
}

// OptionFloat64 provides a mock function with given fields: key
func (_m *Context) OptionFloat64(key string) float64 {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for OptionFloat64")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func(string) float64); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// Context_OptionFloat64_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OptionFloat64'
type Context_OptionFloat64_Call struct {
	*mock.Call
}

// OptionFloat64 is a helper method to define mock.On call
//   - key string
func (_e *Context_Expecter) OptionFloat64(key interface{}) *Context_OptionFloat64_Call {
	return &Context_OptionFloat64_Call{Call: _e.mock.On("OptionFloat64", key)}
}

func (_c *Context_OptionFloat64_Call) Run(run func(key string)) *Context_OptionFloat64_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Context_OptionFloat64_Call) Return(_a0 float64) *Context_OptionFloat64_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_OptionFloat64_Call) RunAndReturn(run func(string) float64) *Context_OptionFloat64_Call {
	_c.Call.Return(run)
	return _c
}

// OptionFloat64Slice provides a mock function with given fields: key
func (_m *Context) OptionFloat64Slice(key string) []float64 {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for OptionFloat64Slice")
	}

	var r0 []float64
	if rf, ok := ret.Get(0).(func(string) []float64); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]float64)
		}
	}

	return r0
}

// Context_OptionFloat64Slice_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OptionFloat64Slice'
type Context_OptionFloat64Slice_Call struct {
	*mock.Call
}

// OptionFloat64Slice is a helper method to define mock.On call
//   - key string
func (_e *Context_Expecter) OptionFloat64Slice(key interface{}) *Context_OptionFloat64Slice_Call {
	return &Context_OptionFloat64Slice_Call{Call: _e.mock.On("OptionFloat64Slice", key)}
}

func (_c *Context_OptionFloat64Slice_Call) Run(run func(key string)) *Context_OptionFloat64Slice_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Context_OptionFloat64Slice_Call) Return(_a0 []float64) *Context_OptionFloat64Slice_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_OptionFloat64Slice_Call) RunAndReturn(run func(string) []float64) *Context_OptionFloat64Slice_Call {
	_c.Call.Return(run)
	return _c
}

// OptionInt provides a mock function with given fields: key
func (_m *Context) OptionInt(key string) int {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for OptionInt")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Context_OptionInt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OptionInt'
type Context_OptionInt_Call struct {
	*mock.Call
}

// OptionInt is a helper method to define mock.On call
//   - key string
func (_e *Context_Expecter) OptionInt(key interface{}) *Context_OptionInt_Call {
	return &Context_OptionInt_Call{Call: _e.mock.On("OptionInt", key)}
}

func (_c *Context_OptionInt_Call) Run(run func(key string)) *Context_OptionInt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Context_OptionInt_Call) Return(_a0 int) *Context_OptionInt_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_OptionInt_Call) RunAndReturn(run func(string) int) *Context_OptionInt_Call {
	_c.Call.Return(run)
	return _c
}

// OptionInt64 provides a mock function with given fields: key
func (_m *Context) OptionInt64(key string) int64 {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for OptionInt64")
	}

	var r0 int64
	if rf, ok := ret.Get(0).(func(string) int64); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// Context_OptionInt64_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OptionInt64'
type Context_OptionInt64_Call struct {
	*mock.Call
}

// OptionInt64 is a helper method to define mock.On call
//   - key string
func (_e *Context_Expecter) OptionInt64(key interface{}) *Context_OptionInt64_Call {
	return &Context_OptionInt64_Call{Call: _e.mock.On("OptionInt64", key)}
}

func (_c *Context_OptionInt64_Call) Run(run func(key string)) *Context_OptionInt64_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Context_OptionInt64_Call) Return(_a0 int64) *Context_OptionInt64_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_OptionInt64_Call) RunAndReturn(run func(string) int64) *Context_OptionInt64_Call {
	_c.Call.Return(run)
	return _c
}

// OptionInt64Slice provides a mock function with given fields: key
func (_m *Context) OptionInt64Slice(key string) []int64 {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for OptionInt64Slice")
	}

	var r0 []int64
	if rf, ok := ret.Get(0).(func(string) []int64); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int64)
		}
	}

	return r0
}

// Context_OptionInt64Slice_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OptionInt64Slice'
type Context_OptionInt64Slice_Call struct {
	*mock.Call
}

// OptionInt64Slice is a helper method to define mock.On call
//   - key string
func (_e *Context_Expecter) OptionInt64Slice(key interface{}) *Context_OptionInt64Slice_Call {
	return &Context_OptionInt64Slice_Call{Call: _e.mock.On("OptionInt64Slice", key)}
}

func (_c *Context_OptionInt64Slice_Call) Run(run func(key string)) *Context_OptionInt64Slice_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Context_OptionInt64Slice_Call) Return(_a0 []int64) *Context_OptionInt64Slice_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_OptionInt64Slice_Call) RunAndReturn(run func(string) []int64) *Context_OptionInt64Slice_Call {
	_c.Call.Return(run)
	return _c
}

// OptionIntSlice provides a mock function with given fields: key
func (_m *Context) OptionIntSlice(key string) []int {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for OptionIntSlice")
	}

	var r0 []int
	if rf, ok := ret.Get(0).(func(string) []int); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int)
		}
	}

	return r0
}

// Context_OptionIntSlice_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OptionIntSlice'
type Context_OptionIntSlice_Call struct {
	*mock.Call
}

// OptionIntSlice is a helper method to define mock.On call
//   - key string
func (_e *Context_Expecter) OptionIntSlice(key interface{}) *Context_OptionIntSlice_Call {
	return &Context_OptionIntSlice_Call{Call: _e.mock.On("OptionIntSlice", key)}
}

func (_c *Context_OptionIntSlice_Call) Run(run func(key string)) *Context_OptionIntSlice_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Context_OptionIntSlice_Call) Return(_a0 []int) *Context_OptionIntSlice_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_OptionIntSlice_Call) RunAndReturn(run func(string) []int) *Context_OptionIntSlice_Call {
	_c.Call.Return(run)
	return _c
}

// OptionSlice provides a mock function with given fields: key
func (_m *Context) OptionSlice(key string) []string {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for OptionSlice")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Context_OptionSlice_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OptionSlice'
type Context_OptionSlice_Call struct {
	*mock.Call
}

// OptionSlice is a helper method to define mock.On call
//   - key string
func (_e *Context_Expecter) OptionSlice(key interface{}) *Context_OptionSlice_Call {
	return &Context_OptionSlice_Call{Call: _e.mock.On("OptionSlice", key)}
}

func (_c *Context_OptionSlice_Call) Run(run func(key string)) *Context_OptionSlice_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Context_OptionSlice_Call) Return(_a0 []string) *Context_OptionSlice_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Context_OptionSlice_Call) RunAndReturn(run func(string) []string) *Context_OptionSlice_Call {
	_c.Call.Return(run)
	return _c
}

// NewContext creates a new instance of Context. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *Context {
	mock := &Context{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
