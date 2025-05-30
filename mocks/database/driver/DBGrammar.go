// Code generated by mockery. DO NOT EDIT.

package driver

import (
	driver "github.com/goravel/framework/contracts/database/driver"
	mock "github.com/stretchr/testify/mock"

	squirrel "github.com/Masterminds/squirrel"
)

// DBGrammar is an autogenerated mock type for the DBGrammar type
type DBGrammar struct {
	mock.Mock
}

type DBGrammar_Expecter struct {
	mock *mock.Mock
}

func (_m *DBGrammar) EXPECT() *DBGrammar_Expecter {
	return &DBGrammar_Expecter{mock: &_m.Mock}
}

// CompileInRandomOrder provides a mock function with given fields: builder, conditions
func (_m *DBGrammar) CompileInRandomOrder(builder squirrel.SelectBuilder, conditions *driver.Conditions) squirrel.SelectBuilder {
	ret := _m.Called(builder, conditions)

	if len(ret) == 0 {
		panic("no return value specified for CompileInRandomOrder")
	}

	var r0 squirrel.SelectBuilder
	if rf, ok := ret.Get(0).(func(squirrel.SelectBuilder, *driver.Conditions) squirrel.SelectBuilder); ok {
		r0 = rf(builder, conditions)
	} else {
		r0 = ret.Get(0).(squirrel.SelectBuilder)
	}

	return r0
}

// DBGrammar_CompileInRandomOrder_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileInRandomOrder'
type DBGrammar_CompileInRandomOrder_Call struct {
	*mock.Call
}

// CompileInRandomOrder is a helper method to define mock.On call
//   - builder squirrel.SelectBuilder
//   - conditions *driver.Conditions
func (_e *DBGrammar_Expecter) CompileInRandomOrder(builder interface{}, conditions interface{}) *DBGrammar_CompileInRandomOrder_Call {
	return &DBGrammar_CompileInRandomOrder_Call{Call: _e.mock.On("CompileInRandomOrder", builder, conditions)}
}

func (_c *DBGrammar_CompileInRandomOrder_Call) Run(run func(builder squirrel.SelectBuilder, conditions *driver.Conditions)) *DBGrammar_CompileInRandomOrder_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(squirrel.SelectBuilder), args[1].(*driver.Conditions))
	})
	return _c
}

func (_c *DBGrammar_CompileInRandomOrder_Call) Return(_a0 squirrel.SelectBuilder) *DBGrammar_CompileInRandomOrder_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DBGrammar_CompileInRandomOrder_Call) RunAndReturn(run func(squirrel.SelectBuilder, *driver.Conditions) squirrel.SelectBuilder) *DBGrammar_CompileInRandomOrder_Call {
	_c.Call.Return(run)
	return _c
}

// CompileLockForUpdate provides a mock function with given fields: builder, conditions
func (_m *DBGrammar) CompileLockForUpdate(builder squirrel.SelectBuilder, conditions *driver.Conditions) squirrel.SelectBuilder {
	ret := _m.Called(builder, conditions)

	if len(ret) == 0 {
		panic("no return value specified for CompileLockForUpdate")
	}

	var r0 squirrel.SelectBuilder
	if rf, ok := ret.Get(0).(func(squirrel.SelectBuilder, *driver.Conditions) squirrel.SelectBuilder); ok {
		r0 = rf(builder, conditions)
	} else {
		r0 = ret.Get(0).(squirrel.SelectBuilder)
	}

	return r0
}

// DBGrammar_CompileLockForUpdate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileLockForUpdate'
type DBGrammar_CompileLockForUpdate_Call struct {
	*mock.Call
}

// CompileLockForUpdate is a helper method to define mock.On call
//   - builder squirrel.SelectBuilder
//   - conditions *driver.Conditions
func (_e *DBGrammar_Expecter) CompileLockForUpdate(builder interface{}, conditions interface{}) *DBGrammar_CompileLockForUpdate_Call {
	return &DBGrammar_CompileLockForUpdate_Call{Call: _e.mock.On("CompileLockForUpdate", builder, conditions)}
}

func (_c *DBGrammar_CompileLockForUpdate_Call) Run(run func(builder squirrel.SelectBuilder, conditions *driver.Conditions)) *DBGrammar_CompileLockForUpdate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(squirrel.SelectBuilder), args[1].(*driver.Conditions))
	})
	return _c
}

func (_c *DBGrammar_CompileLockForUpdate_Call) Return(_a0 squirrel.SelectBuilder) *DBGrammar_CompileLockForUpdate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DBGrammar_CompileLockForUpdate_Call) RunAndReturn(run func(squirrel.SelectBuilder, *driver.Conditions) squirrel.SelectBuilder) *DBGrammar_CompileLockForUpdate_Call {
	_c.Call.Return(run)
	return _c
}

// CompilePlaceholderFormat provides a mock function with no fields
func (_m *DBGrammar) CompilePlaceholderFormat() driver.PlaceholderFormat {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CompilePlaceholderFormat")
	}

	var r0 driver.PlaceholderFormat
	if rf, ok := ret.Get(0).(func() driver.PlaceholderFormat); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(driver.PlaceholderFormat)
		}
	}

	return r0
}

// DBGrammar_CompilePlaceholderFormat_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompilePlaceholderFormat'
type DBGrammar_CompilePlaceholderFormat_Call struct {
	*mock.Call
}

// CompilePlaceholderFormat is a helper method to define mock.On call
func (_e *DBGrammar_Expecter) CompilePlaceholderFormat() *DBGrammar_CompilePlaceholderFormat_Call {
	return &DBGrammar_CompilePlaceholderFormat_Call{Call: _e.mock.On("CompilePlaceholderFormat")}
}

func (_c *DBGrammar_CompilePlaceholderFormat_Call) Run(run func()) *DBGrammar_CompilePlaceholderFormat_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DBGrammar_CompilePlaceholderFormat_Call) Return(_a0 driver.PlaceholderFormat) *DBGrammar_CompilePlaceholderFormat_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DBGrammar_CompilePlaceholderFormat_Call) RunAndReturn(run func() driver.PlaceholderFormat) *DBGrammar_CompilePlaceholderFormat_Call {
	_c.Call.Return(run)
	return _c
}

// CompileSharedLock provides a mock function with given fields: builder, conditions
func (_m *DBGrammar) CompileSharedLock(builder squirrel.SelectBuilder, conditions *driver.Conditions) squirrel.SelectBuilder {
	ret := _m.Called(builder, conditions)

	if len(ret) == 0 {
		panic("no return value specified for CompileSharedLock")
	}

	var r0 squirrel.SelectBuilder
	if rf, ok := ret.Get(0).(func(squirrel.SelectBuilder, *driver.Conditions) squirrel.SelectBuilder); ok {
		r0 = rf(builder, conditions)
	} else {
		r0 = ret.Get(0).(squirrel.SelectBuilder)
	}

	return r0
}

// DBGrammar_CompileSharedLock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileSharedLock'
type DBGrammar_CompileSharedLock_Call struct {
	*mock.Call
}

// CompileSharedLock is a helper method to define mock.On call
//   - builder squirrel.SelectBuilder
//   - conditions *driver.Conditions
func (_e *DBGrammar_Expecter) CompileSharedLock(builder interface{}, conditions interface{}) *DBGrammar_CompileSharedLock_Call {
	return &DBGrammar_CompileSharedLock_Call{Call: _e.mock.On("CompileSharedLock", builder, conditions)}
}

func (_c *DBGrammar_CompileSharedLock_Call) Run(run func(builder squirrel.SelectBuilder, conditions *driver.Conditions)) *DBGrammar_CompileSharedLock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(squirrel.SelectBuilder), args[1].(*driver.Conditions))
	})
	return _c
}

func (_c *DBGrammar_CompileSharedLock_Call) Return(_a0 squirrel.SelectBuilder) *DBGrammar_CompileSharedLock_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DBGrammar_CompileSharedLock_Call) RunAndReturn(run func(squirrel.SelectBuilder, *driver.Conditions) squirrel.SelectBuilder) *DBGrammar_CompileSharedLock_Call {
	_c.Call.Return(run)
	return _c
}

// CompileVersion provides a mock function with no fields
func (_m *DBGrammar) CompileVersion() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CompileVersion")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// DBGrammar_CompileVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileVersion'
type DBGrammar_CompileVersion_Call struct {
	*mock.Call
}

// CompileVersion is a helper method to define mock.On call
func (_e *DBGrammar_Expecter) CompileVersion() *DBGrammar_CompileVersion_Call {
	return &DBGrammar_CompileVersion_Call{Call: _e.mock.On("CompileVersion")}
}

func (_c *DBGrammar_CompileVersion_Call) Run(run func()) *DBGrammar_CompileVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DBGrammar_CompileVersion_Call) Return(_a0 string) *DBGrammar_CompileVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DBGrammar_CompileVersion_Call) RunAndReturn(run func() string) *DBGrammar_CompileVersion_Call {
	_c.Call.Return(run)
	return _c
}

// NewDBGrammar creates a new instance of DBGrammar. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDBGrammar(t interface {
	mock.TestingT
	Cleanup(func())
}) *DBGrammar {
	mock := &DBGrammar{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
