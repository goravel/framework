// Code generated by mockery. DO NOT EDIT.

package driver

import (
	driver "github.com/goravel/framework/contracts/database/driver"
	mock "github.com/stretchr/testify/mock"

	orm "github.com/goravel/framework/contracts/database/orm"
)

// Schema is an autogenerated mock type for the Schema type
type Schema struct {
	mock.Mock
}

type Schema_Expecter struct {
	mock *mock.Mock
}

func (_m *Schema) EXPECT() *Schema_Expecter {
	return &Schema_Expecter{mock: &_m.Mock}
}

// GetColumns provides a mock function with given fields: table
func (_m *Schema) GetColumns(table string) ([]driver.Column, error) {
	ret := _m.Called(table)

	if len(ret) == 0 {
		panic("no return value specified for GetColumns")
	}

	var r0 []driver.Column
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]driver.Column, error)); ok {
		return rf(table)
	}
	if rf, ok := ret.Get(0).(func(string) []driver.Column); ok {
		r0 = rf(table)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]driver.Column)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(table)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Schema_GetColumns_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetColumns'
type Schema_GetColumns_Call struct {
	*mock.Call
}

// GetColumns is a helper method to define mock.On call
//   - table string
func (_e *Schema_Expecter) GetColumns(table interface{}) *Schema_GetColumns_Call {
	return &Schema_GetColumns_Call{Call: _e.mock.On("GetColumns", table)}
}

func (_c *Schema_GetColumns_Call) Run(run func(table string)) *Schema_GetColumns_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_GetColumns_Call) Return(_a0 []driver.Column, _a1 error) *Schema_GetColumns_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Schema_GetColumns_Call) RunAndReturn(run func(string) ([]driver.Column, error)) *Schema_GetColumns_Call {
	_c.Call.Return(run)
	return _c
}

// GetIndexes provides a mock function with given fields: table
func (_m *Schema) GetIndexes(table string) ([]driver.Index, error) {
	ret := _m.Called(table)

	if len(ret) == 0 {
		panic("no return value specified for GetIndexes")
	}

	var r0 []driver.Index
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]driver.Index, error)); ok {
		return rf(table)
	}
	if rf, ok := ret.Get(0).(func(string) []driver.Index); ok {
		r0 = rf(table)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]driver.Index)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(table)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Schema_GetIndexes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIndexes'
type Schema_GetIndexes_Call struct {
	*mock.Call
}

// GetIndexes is a helper method to define mock.On call
//   - table string
func (_e *Schema_Expecter) GetIndexes(table interface{}) *Schema_GetIndexes_Call {
	return &Schema_GetIndexes_Call{Call: _e.mock.On("GetIndexes", table)}
}

func (_c *Schema_GetIndexes_Call) Run(run func(table string)) *Schema_GetIndexes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_GetIndexes_Call) Return(_a0 []driver.Index, _a1 error) *Schema_GetIndexes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Schema_GetIndexes_Call) RunAndReturn(run func(string) ([]driver.Index, error)) *Schema_GetIndexes_Call {
	_c.Call.Return(run)
	return _c
}

// Orm provides a mock function with no fields
func (_m *Schema) Orm() orm.Orm {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Orm")
	}

	var r0 orm.Orm
	if rf, ok := ret.Get(0).(func() orm.Orm); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(orm.Orm)
		}
	}

	return r0
}

// Schema_Orm_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Orm'
type Schema_Orm_Call struct {
	*mock.Call
}

// Orm is a helper method to define mock.On call
func (_e *Schema_Expecter) Orm() *Schema_Orm_Call {
	return &Schema_Orm_Call{Call: _e.mock.On("Orm")}
}

func (_c *Schema_Orm_Call) Run(run func()) *Schema_Orm_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_Orm_Call) Return(_a0 orm.Orm) *Schema_Orm_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_Orm_Call) RunAndReturn(run func() orm.Orm) *Schema_Orm_Call {
	_c.Call.Return(run)
	return _c
}

// NewSchema creates a new instance of Schema. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSchema(t interface {
	mock.TestingT
	Cleanup(func())
}) *Schema {
	mock := &Schema{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
