// Code generated by mockery. DO NOT EDIT.

package session

import mock "github.com/stretchr/testify/mock"

// Driver is an autogenerated mock type for the Driver type
type Driver struct {
	mock.Mock
}

type Driver_Expecter struct {
	mock *mock.Mock
}

func (_m *Driver) EXPECT() *Driver_Expecter {
	return &Driver_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with no fields
func (_m *Driver) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Driver_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Driver_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Driver_Expecter) Close() *Driver_Close_Call {
	return &Driver_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Driver_Close_Call) Run(run func()) *Driver_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_Close_Call) Return(_a0 error) *Driver_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Close_Call) RunAndReturn(run func() error) *Driver_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Destroy provides a mock function with given fields: id
func (_m *Driver) Destroy(id string) error {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Destroy")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Driver_Destroy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Destroy'
type Driver_Destroy_Call struct {
	*mock.Call
}

// Destroy is a helper method to define mock.On call
//   - id string
func (_e *Driver_Expecter) Destroy(id interface{}) *Driver_Destroy_Call {
	return &Driver_Destroy_Call{Call: _e.mock.On("Destroy", id)}
}

func (_c *Driver_Destroy_Call) Run(run func(id string)) *Driver_Destroy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Driver_Destroy_Call) Return(_a0 error) *Driver_Destroy_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Destroy_Call) RunAndReturn(run func(string) error) *Driver_Destroy_Call {
	_c.Call.Return(run)
	return _c
}

// Gc provides a mock function with given fields: maxLifetime
func (_m *Driver) Gc(maxLifetime int) error {
	ret := _m.Called(maxLifetime)

	if len(ret) == 0 {
		panic("no return value specified for Gc")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(maxLifetime)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Driver_Gc_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Gc'
type Driver_Gc_Call struct {
	*mock.Call
}

// Gc is a helper method to define mock.On call
//   - maxLifetime int
func (_e *Driver_Expecter) Gc(maxLifetime interface{}) *Driver_Gc_Call {
	return &Driver_Gc_Call{Call: _e.mock.On("Gc", maxLifetime)}
}

func (_c *Driver_Gc_Call) Run(run func(maxLifetime int)) *Driver_Gc_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *Driver_Gc_Call) Return(_a0 error) *Driver_Gc_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Gc_Call) RunAndReturn(run func(int) error) *Driver_Gc_Call {
	_c.Call.Return(run)
	return _c
}

// Open provides a mock function with given fields: path, name
func (_m *Driver) Open(path string, name string) error {
	ret := _m.Called(path, name)

	if len(ret) == 0 {
		panic("no return value specified for Open")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(path, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Driver_Open_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Open'
type Driver_Open_Call struct {
	*mock.Call
}

// Open is a helper method to define mock.On call
//   - path string
//   - name string
func (_e *Driver_Expecter) Open(path interface{}, name interface{}) *Driver_Open_Call {
	return &Driver_Open_Call{Call: _e.mock.On("Open", path, name)}
}

func (_c *Driver_Open_Call) Run(run func(path string, name string)) *Driver_Open_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Driver_Open_Call) Return(_a0 error) *Driver_Open_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Open_Call) RunAndReturn(run func(string, string) error) *Driver_Open_Call {
	_c.Call.Return(run)
	return _c
}

// Read provides a mock function with given fields: id
func (_m *Driver) Read(id string) (string, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Read")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Driver_Read_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Read'
type Driver_Read_Call struct {
	*mock.Call
}

// Read is a helper method to define mock.On call
//   - id string
func (_e *Driver_Expecter) Read(id interface{}) *Driver_Read_Call {
	return &Driver_Read_Call{Call: _e.mock.On("Read", id)}
}

func (_c *Driver_Read_Call) Run(run func(id string)) *Driver_Read_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Driver_Read_Call) Return(_a0 string, _a1 error) *Driver_Read_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Driver_Read_Call) RunAndReturn(run func(string) (string, error)) *Driver_Read_Call {
	_c.Call.Return(run)
	return _c
}

// Write provides a mock function with given fields: id, data
func (_m *Driver) Write(id string, data string) error {
	ret := _m.Called(id, data)

	if len(ret) == 0 {
		panic("no return value specified for Write")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(id, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Driver_Write_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Write'
type Driver_Write_Call struct {
	*mock.Call
}

// Write is a helper method to define mock.On call
//   - id string
//   - data string
func (_e *Driver_Expecter) Write(id interface{}, data interface{}) *Driver_Write_Call {
	return &Driver_Write_Call{Call: _e.mock.On("Write", id, data)}
}

func (_c *Driver_Write_Call) Run(run func(id string, data string)) *Driver_Write_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Driver_Write_Call) Return(_a0 error) *Driver_Write_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Write_Call) RunAndReturn(run func(string, string) error) *Driver_Write_Call {
	_c.Call.Return(run)
	return _c
}

// NewDriver creates a new instance of Driver. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDriver(t interface {
	mock.TestingT
	Cleanup(func())
}) *Driver {
	mock := &Driver{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
