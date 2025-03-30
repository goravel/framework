// Code generated by mockery. DO NOT EDIT.

package driver

import (
	driver "github.com/goravel/framework/contracts/database/driver"
	mock "github.com/stretchr/testify/mock"
)

// Processor is an autogenerated mock type for the Processor type
type Processor struct {
	mock.Mock
}

type Processor_Expecter struct {
	mock *mock.Mock
}

func (_m *Processor) EXPECT() *Processor_Expecter {
	return &Processor_Expecter{mock: &_m.Mock}
}

// ProcessColumns provides a mock function with given fields: dbColumns
func (_m *Processor) ProcessColumns(dbColumns []driver.DBColumn) []driver.Column {
	ret := _m.Called(dbColumns)

	if len(ret) == 0 {
		panic("no return value specified for ProcessColumns")
	}

	var r0 []driver.Column
	if rf, ok := ret.Get(0).(func([]driver.DBColumn) []driver.Column); ok {
		r0 = rf(dbColumns)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]driver.Column)
		}
	}

	return r0
}

// Processor_ProcessColumns_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProcessColumns'
type Processor_ProcessColumns_Call struct {
	*mock.Call
}

// ProcessColumns is a helper method to define mock.On call
//   - dbColumns []driver.DBColumn
func (_e *Processor_Expecter) ProcessColumns(dbColumns interface{}) *Processor_ProcessColumns_Call {
	return &Processor_ProcessColumns_Call{Call: _e.mock.On("ProcessColumns", dbColumns)}
}

func (_c *Processor_ProcessColumns_Call) Run(run func(dbColumns []driver.DBColumn)) *Processor_ProcessColumns_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]driver.DBColumn))
	})
	return _c
}

func (_c *Processor_ProcessColumns_Call) Return(_a0 []driver.Column) *Processor_ProcessColumns_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Processor_ProcessColumns_Call) RunAndReturn(run func([]driver.DBColumn) []driver.Column) *Processor_ProcessColumns_Call {
	_c.Call.Return(run)
	return _c
}

// ProcessForeignKeys provides a mock function with given fields: dbIndexes
func (_m *Processor) ProcessForeignKeys(dbIndexes []driver.DBForeignKey) []driver.ForeignKey {
	ret := _m.Called(dbIndexes)

	if len(ret) == 0 {
		panic("no return value specified for ProcessForeignKeys")
	}

	var r0 []driver.ForeignKey
	if rf, ok := ret.Get(0).(func([]driver.DBForeignKey) []driver.ForeignKey); ok {
		r0 = rf(dbIndexes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]driver.ForeignKey)
		}
	}

	return r0
}

// Processor_ProcessForeignKeys_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProcessForeignKeys'
type Processor_ProcessForeignKeys_Call struct {
	*mock.Call
}

// ProcessForeignKeys is a helper method to define mock.On call
//   - dbIndexes []driver.DBForeignKey
func (_e *Processor_Expecter) ProcessForeignKeys(dbIndexes interface{}) *Processor_ProcessForeignKeys_Call {
	return &Processor_ProcessForeignKeys_Call{Call: _e.mock.On("ProcessForeignKeys", dbIndexes)}
}

func (_c *Processor_ProcessForeignKeys_Call) Run(run func(dbIndexes []driver.DBForeignKey)) *Processor_ProcessForeignKeys_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]driver.DBForeignKey))
	})
	return _c
}

func (_c *Processor_ProcessForeignKeys_Call) Return(_a0 []driver.ForeignKey) *Processor_ProcessForeignKeys_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Processor_ProcessForeignKeys_Call) RunAndReturn(run func([]driver.DBForeignKey) []driver.ForeignKey) *Processor_ProcessForeignKeys_Call {
	_c.Call.Return(run)
	return _c
}

// ProcessIndexes provides a mock function with given fields: dbIndexes
func (_m *Processor) ProcessIndexes(dbIndexes []driver.DBIndex) []driver.Index {
	ret := _m.Called(dbIndexes)

	if len(ret) == 0 {
		panic("no return value specified for ProcessIndexes")
	}

	var r0 []driver.Index
	if rf, ok := ret.Get(0).(func([]driver.DBIndex) []driver.Index); ok {
		r0 = rf(dbIndexes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]driver.Index)
		}
	}

	return r0
}

// Processor_ProcessIndexes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProcessIndexes'
type Processor_ProcessIndexes_Call struct {
	*mock.Call
}

// ProcessIndexes is a helper method to define mock.On call
//   - dbIndexes []driver.DBIndex
func (_e *Processor_Expecter) ProcessIndexes(dbIndexes interface{}) *Processor_ProcessIndexes_Call {
	return &Processor_ProcessIndexes_Call{Call: _e.mock.On("ProcessIndexes", dbIndexes)}
}

func (_c *Processor_ProcessIndexes_Call) Run(run func(dbIndexes []driver.DBIndex)) *Processor_ProcessIndexes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]driver.DBIndex))
	})
	return _c
}

func (_c *Processor_ProcessIndexes_Call) Return(_a0 []driver.Index) *Processor_ProcessIndexes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Processor_ProcessIndexes_Call) RunAndReturn(run func([]driver.DBIndex) []driver.Index) *Processor_ProcessIndexes_Call {
	_c.Call.Return(run)
	return _c
}

// ProcessTypes provides a mock function with given fields: types
func (_m *Processor) ProcessTypes(types []driver.Type) []driver.Type {
	ret := _m.Called(types)

	if len(ret) == 0 {
		panic("no return value specified for ProcessTypes")
	}

	var r0 []driver.Type
	if rf, ok := ret.Get(0).(func([]driver.Type) []driver.Type); ok {
		r0 = rf(types)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]driver.Type)
		}
	}

	return r0
}

// Processor_ProcessTypes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProcessTypes'
type Processor_ProcessTypes_Call struct {
	*mock.Call
}

// ProcessTypes is a helper method to define mock.On call
//   - types []driver.Type
func (_e *Processor_Expecter) ProcessTypes(types interface{}) *Processor_ProcessTypes_Call {
	return &Processor_ProcessTypes_Call{Call: _e.mock.On("ProcessTypes", types)}
}

func (_c *Processor_ProcessTypes_Call) Run(run func(types []driver.Type)) *Processor_ProcessTypes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]driver.Type))
	})
	return _c
}

func (_c *Processor_ProcessTypes_Call) Return(_a0 []driver.Type) *Processor_ProcessTypes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Processor_ProcessTypes_Call) RunAndReturn(run func([]driver.Type) []driver.Type) *Processor_ProcessTypes_Call {
	_c.Call.Return(run)
	return _c
}

// NewProcessor creates a new instance of Processor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProcessor(t interface {
	mock.TestingT
	Cleanup(func())
}) *Processor {
	mock := &Processor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
