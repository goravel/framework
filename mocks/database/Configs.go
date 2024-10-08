// Code generated by mockery. DO NOT EDIT.

package database

import (
	database "github.com/goravel/framework/contracts/database"
	mock "github.com/stretchr/testify/mock"
)

// Configs is an autogenerated mock type for the Configs type
type Configs struct {
	mock.Mock
}

type Configs_Expecter struct {
	mock *mock.Mock
}

func (_m *Configs) EXPECT() *Configs_Expecter {
	return &Configs_Expecter{mock: &_m.Mock}
}

// Reads provides a mock function with given fields:
func (_m *Configs) Reads() []database.FullConfig {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Reads")
	}

	var r0 []database.FullConfig
	if rf, ok := ret.Get(0).(func() []database.FullConfig); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]database.FullConfig)
		}
	}

	return r0
}

// Configs_Reads_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reads'
type Configs_Reads_Call struct {
	*mock.Call
}

// Reads is a helper method to define mock.On call
func (_e *Configs_Expecter) Reads() *Configs_Reads_Call {
	return &Configs_Reads_Call{Call: _e.mock.On("Reads")}
}

func (_c *Configs_Reads_Call) Run(run func()) *Configs_Reads_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Configs_Reads_Call) Return(_a0 []database.FullConfig) *Configs_Reads_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Configs_Reads_Call) RunAndReturn(run func() []database.FullConfig) *Configs_Reads_Call {
	_c.Call.Return(run)
	return _c
}

// Writes provides a mock function with given fields:
func (_m *Configs) Writes() []database.FullConfig {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Writes")
	}

	var r0 []database.FullConfig
	if rf, ok := ret.Get(0).(func() []database.FullConfig); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]database.FullConfig)
		}
	}

	return r0
}

// Configs_Writes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Writes'
type Configs_Writes_Call struct {
	*mock.Call
}

// Writes is a helper method to define mock.On call
func (_e *Configs_Expecter) Writes() *Configs_Writes_Call {
	return &Configs_Writes_Call{Call: _e.mock.On("Writes")}
}

func (_c *Configs_Writes_Call) Run(run func()) *Configs_Writes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Configs_Writes_Call) Return(_a0 []database.FullConfig) *Configs_Writes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Configs_Writes_Call) RunAndReturn(run func() []database.FullConfig) *Configs_Writes_Call {
	_c.Call.Return(run)
	return _c
}

// NewConfigs creates a new instance of Configs. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConfigs(t interface {
	mock.TestingT
	Cleanup(func())
}) *Configs {
	mock := &Configs{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
