// Code generated by mockery. DO NOT EDIT.

package driver

import (
	database "github.com/goravel/framework/contracts/database"
	docker "github.com/goravel/framework/contracts/testing/docker"

	driver "github.com/goravel/framework/contracts/database/driver"

	gorm "gorm.io/gorm"

	mock "github.com/stretchr/testify/mock"

	schema "github.com/goravel/framework/contracts/database/schema"

	sqlx "github.com/jmoiron/sqlx"
)

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

// Config provides a mock function with no fields
func (_m *Driver) Config() database.Config {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Config")
	}

	var r0 database.Config
	if rf, ok := ret.Get(0).(func() database.Config); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(database.Config)
	}

	return r0
}

// Driver_Config_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Config'
type Driver_Config_Call struct {
	*mock.Call
}

// Config is a helper method to define mock.On call
func (_e *Driver_Expecter) Config() *Driver_Config_Call {
	return &Driver_Config_Call{Call: _e.mock.On("Config")}
}

func (_c *Driver_Config_Call) Run(run func()) *Driver_Config_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_Config_Call) Return(_a0 database.Config) *Driver_Config_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Config_Call) RunAndReturn(run func() database.Config) *Driver_Config_Call {
	_c.Call.Return(run)
	return _c
}

// DB provides a mock function with no fields
func (_m *Driver) DB() (*sqlx.DB, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for DB")
	}

	var r0 *sqlx.DB
	var r1 error
	if rf, ok := ret.Get(0).(func() (*sqlx.DB, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *sqlx.DB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlx.DB)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Driver_DB_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DB'
type Driver_DB_Call struct {
	*mock.Call
}

// DB is a helper method to define mock.On call
func (_e *Driver_Expecter) DB() *Driver_DB_Call {
	return &Driver_DB_Call{Call: _e.mock.On("DB")}
}

func (_c *Driver_DB_Call) Run(run func()) *Driver_DB_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_DB_Call) Return(_a0 *sqlx.DB, _a1 error) *Driver_DB_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Driver_DB_Call) RunAndReturn(run func() (*sqlx.DB, error)) *Driver_DB_Call {
	_c.Call.Return(run)
	return _c
}

// Docker provides a mock function with no fields
func (_m *Driver) Docker() (docker.DatabaseDriver, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Docker")
	}

	var r0 docker.DatabaseDriver
	var r1 error
	if rf, ok := ret.Get(0).(func() (docker.DatabaseDriver, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() docker.DatabaseDriver); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(docker.DatabaseDriver)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Driver_Docker_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Docker'
type Driver_Docker_Call struct {
	*mock.Call
}

// Docker is a helper method to define mock.On call
func (_e *Driver_Expecter) Docker() *Driver_Docker_Call {
	return &Driver_Docker_Call{Call: _e.mock.On("Docker")}
}

func (_c *Driver_Docker_Call) Run(run func()) *Driver_Docker_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_Docker_Call) Return(_a0 docker.DatabaseDriver, _a1 error) *Driver_Docker_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Driver_Docker_Call) RunAndReturn(run func() (docker.DatabaseDriver, error)) *Driver_Docker_Call {
	_c.Call.Return(run)
	return _c
}

// Gorm provides a mock function with no fields
func (_m *Driver) Gorm() (*gorm.DB, driver.GormQuery, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Gorm")
	}

	var r0 *gorm.DB
	var r1 driver.GormQuery
	var r2 error
	if rf, ok := ret.Get(0).(func() (*gorm.DB, driver.GormQuery, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *gorm.DB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gorm.DB)
		}
	}

	if rf, ok := ret.Get(1).(func() driver.GormQuery); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(driver.GormQuery)
		}
	}

	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Driver_Gorm_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Gorm'
type Driver_Gorm_Call struct {
	*mock.Call
}

// Gorm is a helper method to define mock.On call
func (_e *Driver_Expecter) Gorm() *Driver_Gorm_Call {
	return &Driver_Gorm_Call{Call: _e.mock.On("Gorm")}
}

func (_c *Driver_Gorm_Call) Run(run func()) *Driver_Gorm_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_Gorm_Call) Return(_a0 *gorm.DB, _a1 driver.GormQuery, _a2 error) *Driver_Gorm_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *Driver_Gorm_Call) RunAndReturn(run func() (*gorm.DB, driver.GormQuery, error)) *Driver_Gorm_Call {
	_c.Call.Return(run)
	return _c
}

// Grammar provides a mock function with no fields
func (_m *Driver) Grammar() schema.Grammar {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Grammar")
	}

	var r0 schema.Grammar
	if rf, ok := ret.Get(0).(func() schema.Grammar); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(schema.Grammar)
		}
	}

	return r0
}

// Driver_Grammar_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Grammar'
type Driver_Grammar_Call struct {
	*mock.Call
}

// Grammar is a helper method to define mock.On call
func (_e *Driver_Expecter) Grammar() *Driver_Grammar_Call {
	return &Driver_Grammar_Call{Call: _e.mock.On("Grammar")}
}

func (_c *Driver_Grammar_Call) Run(run func()) *Driver_Grammar_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_Grammar_Call) Return(_a0 schema.Grammar) *Driver_Grammar_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Grammar_Call) RunAndReturn(run func() schema.Grammar) *Driver_Grammar_Call {
	_c.Call.Return(run)
	return _c
}

// Processor provides a mock function with no fields
func (_m *Driver) Processor() schema.Processor {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Processor")
	}

	var r0 schema.Processor
	if rf, ok := ret.Get(0).(func() schema.Processor); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(schema.Processor)
		}
	}

	return r0
}

// Driver_Processor_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Processor'
type Driver_Processor_Call struct {
	*mock.Call
}

// Processor is a helper method to define mock.On call
func (_e *Driver_Expecter) Processor() *Driver_Processor_Call {
	return &Driver_Processor_Call{Call: _e.mock.On("Processor")}
}

func (_c *Driver_Processor_Call) Run(run func()) *Driver_Processor_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Driver_Processor_Call) Return(_a0 schema.Processor) *Driver_Processor_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Driver_Processor_Call) RunAndReturn(run func() schema.Processor) *Driver_Processor_Call {
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
