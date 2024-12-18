// Code generated by mockery. DO NOT EDIT.

package testing

import (
	database "github.com/goravel/framework/contracts/database"
	mock "github.com/stretchr/testify/mock"

	testing "github.com/goravel/framework/contracts/testing"
)

// DatabaseDriver is an autogenerated mock type for the DatabaseDriver type
type DatabaseDriver struct {
	mock.Mock
}

type DatabaseDriver_Expecter struct {
	mock *mock.Mock
}

func (_m *DatabaseDriver) EXPECT() *DatabaseDriver_Expecter {
	return &DatabaseDriver_Expecter{mock: &_m.Mock}
}

// Build provides a mock function with given fields:
func (_m *DatabaseDriver) Build() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Build")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DatabaseDriver_Build_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Build'
type DatabaseDriver_Build_Call struct {
	*mock.Call
}

// Build is a helper method to define mock.On call
func (_e *DatabaseDriver_Expecter) Build() *DatabaseDriver_Build_Call {
	return &DatabaseDriver_Build_Call{Call: _e.mock.On("Build")}
}

func (_c *DatabaseDriver_Build_Call) Run(run func()) *DatabaseDriver_Build_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DatabaseDriver_Build_Call) Return(_a0 error) *DatabaseDriver_Build_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DatabaseDriver_Build_Call) RunAndReturn(run func() error) *DatabaseDriver_Build_Call {
	_c.Call.Return(run)
	return _c
}

// Config provides a mock function with given fields:
func (_m *DatabaseDriver) Config() testing.DatabaseConfig {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Config")
	}

	var r0 testing.DatabaseConfig
	if rf, ok := ret.Get(0).(func() testing.DatabaseConfig); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(testing.DatabaseConfig)
	}

	return r0
}

// DatabaseDriver_Config_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Config'
type DatabaseDriver_Config_Call struct {
	*mock.Call
}

// Config is a helper method to define mock.On call
func (_e *DatabaseDriver_Expecter) Config() *DatabaseDriver_Config_Call {
	return &DatabaseDriver_Config_Call{Call: _e.mock.On("Config")}
}

func (_c *DatabaseDriver_Config_Call) Run(run func()) *DatabaseDriver_Config_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DatabaseDriver_Config_Call) Return(_a0 testing.DatabaseConfig) *DatabaseDriver_Config_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DatabaseDriver_Config_Call) RunAndReturn(run func() testing.DatabaseConfig) *DatabaseDriver_Config_Call {
	_c.Call.Return(run)
	return _c
}

// Database provides a mock function with given fields: name
func (_m *DatabaseDriver) Database(name string) (testing.DatabaseDriver, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Database")
	}

	var r0 testing.DatabaseDriver
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (testing.DatabaseDriver, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) testing.DatabaseDriver); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.DatabaseDriver)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DatabaseDriver_Database_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Database'
type DatabaseDriver_Database_Call struct {
	*mock.Call
}

// Database is a helper method to define mock.On call
//   - name string
func (_e *DatabaseDriver_Expecter) Database(name interface{}) *DatabaseDriver_Database_Call {
	return &DatabaseDriver_Database_Call{Call: _e.mock.On("Database", name)}
}

func (_c *DatabaseDriver_Database_Call) Run(run func(name string)) *DatabaseDriver_Database_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *DatabaseDriver_Database_Call) Return(_a0 testing.DatabaseDriver, _a1 error) *DatabaseDriver_Database_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DatabaseDriver_Database_Call) RunAndReturn(run func(string) (testing.DatabaseDriver, error)) *DatabaseDriver_Database_Call {
	_c.Call.Return(run)
	return _c
}

// Driver provides a mock function with given fields:
func (_m *DatabaseDriver) Driver() database.Driver {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Driver")
	}

	var r0 database.Driver
	if rf, ok := ret.Get(0).(func() database.Driver); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(database.Driver)
	}

	return r0
}

// DatabaseDriver_Driver_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Driver'
type DatabaseDriver_Driver_Call struct {
	*mock.Call
}

// Driver is a helper method to define mock.On call
func (_e *DatabaseDriver_Expecter) Driver() *DatabaseDriver_Driver_Call {
	return &DatabaseDriver_Driver_Call{Call: _e.mock.On("Driver")}
}

func (_c *DatabaseDriver_Driver_Call) Run(run func()) *DatabaseDriver_Driver_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DatabaseDriver_Driver_Call) Return(_a0 database.Driver) *DatabaseDriver_Driver_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DatabaseDriver_Driver_Call) RunAndReturn(run func() database.Driver) *DatabaseDriver_Driver_Call {
	_c.Call.Return(run)
	return _c
}

// Fresh provides a mock function with given fields:
func (_m *DatabaseDriver) Fresh() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Fresh")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DatabaseDriver_Fresh_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Fresh'
type DatabaseDriver_Fresh_Call struct {
	*mock.Call
}

// Fresh is a helper method to define mock.On call
func (_e *DatabaseDriver_Expecter) Fresh() *DatabaseDriver_Fresh_Call {
	return &DatabaseDriver_Fresh_Call{Call: _e.mock.On("Fresh")}
}

func (_c *DatabaseDriver_Fresh_Call) Run(run func()) *DatabaseDriver_Fresh_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DatabaseDriver_Fresh_Call) Return(_a0 error) *DatabaseDriver_Fresh_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DatabaseDriver_Fresh_Call) RunAndReturn(run func() error) *DatabaseDriver_Fresh_Call {
	_c.Call.Return(run)
	return _c
}

// Image provides a mock function with given fields: image
func (_m *DatabaseDriver) Image(image testing.Image) {
	_m.Called(image)
}

// DatabaseDriver_Image_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Image'
type DatabaseDriver_Image_Call struct {
	*mock.Call
}

// Image is a helper method to define mock.On call
//   - image testing.Image
func (_e *DatabaseDriver_Expecter) Image(image interface{}) *DatabaseDriver_Image_Call {
	return &DatabaseDriver_Image_Call{Call: _e.mock.On("Image", image)}
}

func (_c *DatabaseDriver_Image_Call) Run(run func(image testing.Image)) *DatabaseDriver_Image_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(testing.Image))
	})
	return _c
}

func (_c *DatabaseDriver_Image_Call) Return() *DatabaseDriver_Image_Call {
	_c.Call.Return()
	return _c
}

func (_c *DatabaseDriver_Image_Call) RunAndReturn(run func(testing.Image)) *DatabaseDriver_Image_Call {
	_c.Call.Return(run)
	return _c
}

// Ready provides a mock function with given fields:
func (_m *DatabaseDriver) Ready() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Ready")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DatabaseDriver_Ready_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Ready'
type DatabaseDriver_Ready_Call struct {
	*mock.Call
}

// Ready is a helper method to define mock.On call
func (_e *DatabaseDriver_Expecter) Ready() *DatabaseDriver_Ready_Call {
	return &DatabaseDriver_Ready_Call{Call: _e.mock.On("Ready")}
}

func (_c *DatabaseDriver_Ready_Call) Run(run func()) *DatabaseDriver_Ready_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DatabaseDriver_Ready_Call) Return(_a0 error) *DatabaseDriver_Ready_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DatabaseDriver_Ready_Call) RunAndReturn(run func() error) *DatabaseDriver_Ready_Call {
	_c.Call.Return(run)
	return _c
}

// Shutdown provides a mock function with given fields:
func (_m *DatabaseDriver) Shutdown() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Shutdown")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DatabaseDriver_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type DatabaseDriver_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
func (_e *DatabaseDriver_Expecter) Shutdown() *DatabaseDriver_Shutdown_Call {
	return &DatabaseDriver_Shutdown_Call{Call: _e.mock.On("Shutdown")}
}

func (_c *DatabaseDriver_Shutdown_Call) Run(run func()) *DatabaseDriver_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DatabaseDriver_Shutdown_Call) Return(_a0 error) *DatabaseDriver_Shutdown_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DatabaseDriver_Shutdown_Call) RunAndReturn(run func() error) *DatabaseDriver_Shutdown_Call {
	_c.Call.Return(run)
	return _c
}

// NewDatabaseDriver creates a new instance of DatabaseDriver. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDatabaseDriver(t interface {
	mock.TestingT
	Cleanup(func())
}) *DatabaseDriver {
	mock := &DatabaseDriver{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
