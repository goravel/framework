// Code generated by mockery. DO NOT EDIT.

package schema

import (
	orm "github.com/goravel/framework/contracts/database/orm"
	schema "github.com/goravel/framework/contracts/database/schema"
	mock "github.com/stretchr/testify/mock"
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

// Connection provides a mock function with given fields: name
func (_m *Schema) Connection(name string) schema.Schema {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Connection")
	}

	var r0 schema.Schema
	if rf, ok := ret.Get(0).(func(string) schema.Schema); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(schema.Schema)
		}
	}

	return r0
}

// Schema_Connection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Connection'
type Schema_Connection_Call struct {
	*mock.Call
}

// Connection is a helper method to define mock.On call
//   - name string
func (_e *Schema_Expecter) Connection(name interface{}) *Schema_Connection_Call {
	return &Schema_Connection_Call{Call: _e.mock.On("Connection", name)}
}

func (_c *Schema_Connection_Call) Run(run func(name string)) *Schema_Connection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_Connection_Call) Return(_a0 schema.Schema) *Schema_Connection_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_Connection_Call) RunAndReturn(run func(string) schema.Schema) *Schema_Connection_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: table, callback
func (_m *Schema) Create(table string, callback func(schema.Blueprint)) error {
	ret := _m.Called(table, callback)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, func(schema.Blueprint)) error); ok {
		r0 = rf(table, callback)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type Schema_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - table string
//   - callback func(schema.Blueprint)
func (_e *Schema_Expecter) Create(table interface{}, callback interface{}) *Schema_Create_Call {
	return &Schema_Create_Call{Call: _e.mock.On("Create", table, callback)}
}

func (_c *Schema_Create_Call) Run(run func(table string, callback func(schema.Blueprint))) *Schema_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(func(schema.Blueprint)))
	})
	return _c
}

func (_c *Schema_Create_Call) Return(_a0 error) *Schema_Create_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_Create_Call) RunAndReturn(run func(string, func(schema.Blueprint)) error) *Schema_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Drop provides a mock function with given fields: table
func (_m *Schema) Drop(table string) error {
	ret := _m.Called(table)

	if len(ret) == 0 {
		panic("no return value specified for Drop")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(table)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_Drop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Drop'
type Schema_Drop_Call struct {
	*mock.Call
}

// Drop is a helper method to define mock.On call
//   - table string
func (_e *Schema_Expecter) Drop(table interface{}) *Schema_Drop_Call {
	return &Schema_Drop_Call{Call: _e.mock.On("Drop", table)}
}

func (_c *Schema_Drop_Call) Run(run func(table string)) *Schema_Drop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_Drop_Call) Return(_a0 error) *Schema_Drop_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_Drop_Call) RunAndReturn(run func(string) error) *Schema_Drop_Call {
	_c.Call.Return(run)
	return _c
}

// DropAllTables provides a mock function with given fields:
func (_m *Schema) DropAllTables() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for DropAllTables")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_DropAllTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DropAllTables'
type Schema_DropAllTables_Call struct {
	*mock.Call
}

// DropAllTables is a helper method to define mock.On call
func (_e *Schema_Expecter) DropAllTables() *Schema_DropAllTables_Call {
	return &Schema_DropAllTables_Call{Call: _e.mock.On("DropAllTables")}
}

func (_c *Schema_DropAllTables_Call) Run(run func()) *Schema_DropAllTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_DropAllTables_Call) Return(_a0 error) *Schema_DropAllTables_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_DropAllTables_Call) RunAndReturn(run func() error) *Schema_DropAllTables_Call {
	_c.Call.Return(run)
	return _c
}

// DropAllTypes provides a mock function with given fields:
func (_m *Schema) DropAllTypes() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for DropAllTypes")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_DropAllTypes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DropAllTypes'
type Schema_DropAllTypes_Call struct {
	*mock.Call
}

// DropAllTypes is a helper method to define mock.On call
func (_e *Schema_Expecter) DropAllTypes() *Schema_DropAllTypes_Call {
	return &Schema_DropAllTypes_Call{Call: _e.mock.On("DropAllTypes")}
}

func (_c *Schema_DropAllTypes_Call) Run(run func()) *Schema_DropAllTypes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_DropAllTypes_Call) Return(_a0 error) *Schema_DropAllTypes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_DropAllTypes_Call) RunAndReturn(run func() error) *Schema_DropAllTypes_Call {
	_c.Call.Return(run)
	return _c
}

// DropAllViews provides a mock function with given fields:
func (_m *Schema) DropAllViews() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for DropAllViews")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_DropAllViews_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DropAllViews'
type Schema_DropAllViews_Call struct {
	*mock.Call
}

// DropAllViews is a helper method to define mock.On call
func (_e *Schema_Expecter) DropAllViews() *Schema_DropAllViews_Call {
	return &Schema_DropAllViews_Call{Call: _e.mock.On("DropAllViews")}
}

func (_c *Schema_DropAllViews_Call) Run(run func()) *Schema_DropAllViews_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_DropAllViews_Call) Return(_a0 error) *Schema_DropAllViews_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_DropAllViews_Call) RunAndReturn(run func() error) *Schema_DropAllViews_Call {
	_c.Call.Return(run)
	return _c
}

// DropColumns provides a mock function with given fields: table, columns
func (_m *Schema) DropColumns(table string, columns []string) error {
	ret := _m.Called(table, columns)

	if len(ret) == 0 {
		panic("no return value specified for DropColumns")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []string) error); ok {
		r0 = rf(table, columns)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_DropColumns_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DropColumns'
type Schema_DropColumns_Call struct {
	*mock.Call
}

// DropColumns is a helper method to define mock.On call
//   - table string
//   - columns []string
func (_e *Schema_Expecter) DropColumns(table interface{}, columns interface{}) *Schema_DropColumns_Call {
	return &Schema_DropColumns_Call{Call: _e.mock.On("DropColumns", table, columns)}
}

func (_c *Schema_DropColumns_Call) Run(run func(table string, columns []string)) *Schema_DropColumns_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]string))
	})
	return _c
}

func (_c *Schema_DropColumns_Call) Return(_a0 error) *Schema_DropColumns_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_DropColumns_Call) RunAndReturn(run func(string, []string) error) *Schema_DropColumns_Call {
	_c.Call.Return(run)
	return _c
}

// DropIfExists provides a mock function with given fields: table
func (_m *Schema) DropIfExists(table string) error {
	ret := _m.Called(table)

	if len(ret) == 0 {
		panic("no return value specified for DropIfExists")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(table)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_DropIfExists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DropIfExists'
type Schema_DropIfExists_Call struct {
	*mock.Call
}

// DropIfExists is a helper method to define mock.On call
//   - table string
func (_e *Schema_Expecter) DropIfExists(table interface{}) *Schema_DropIfExists_Call {
	return &Schema_DropIfExists_Call{Call: _e.mock.On("DropIfExists", table)}
}

func (_c *Schema_DropIfExists_Call) Run(run func(table string)) *Schema_DropIfExists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_DropIfExists_Call) Return(_a0 error) *Schema_DropIfExists_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_DropIfExists_Call) RunAndReturn(run func(string) error) *Schema_DropIfExists_Call {
	_c.Call.Return(run)
	return _c
}

// GetColumnListing provides a mock function with given fields: table
func (_m *Schema) GetColumnListing(table string) []string {
	ret := _m.Called(table)

	if len(ret) == 0 {
		panic("no return value specified for GetColumnListing")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(table)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Schema_GetColumnListing_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetColumnListing'
type Schema_GetColumnListing_Call struct {
	*mock.Call
}

// GetColumnListing is a helper method to define mock.On call
//   - table string
func (_e *Schema_Expecter) GetColumnListing(table interface{}) *Schema_GetColumnListing_Call {
	return &Schema_GetColumnListing_Call{Call: _e.mock.On("GetColumnListing", table)}
}

func (_c *Schema_GetColumnListing_Call) Run(run func(table string)) *Schema_GetColumnListing_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_GetColumnListing_Call) Return(_a0 []string) *Schema_GetColumnListing_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_GetColumnListing_Call) RunAndReturn(run func(string) []string) *Schema_GetColumnListing_Call {
	_c.Call.Return(run)
	return _c
}

// GetColumns provides a mock function with given fields: table
func (_m *Schema) GetColumns(table string) ([]schema.Column, error) {
	ret := _m.Called(table)

	if len(ret) == 0 {
		panic("no return value specified for GetColumns")
	}

	var r0 []schema.Column
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]schema.Column, error)); ok {
		return rf(table)
	}
	if rf, ok := ret.Get(0).(func(string) []schema.Column); ok {
		r0 = rf(table)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]schema.Column)
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

func (_c *Schema_GetColumns_Call) Return(_a0 []schema.Column, _a1 error) *Schema_GetColumns_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Schema_GetColumns_Call) RunAndReturn(run func(string) ([]schema.Column, error)) *Schema_GetColumns_Call {
	_c.Call.Return(run)
	return _c
}

// GetConnection provides a mock function with given fields:
func (_m *Schema) GetConnection() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetConnection")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Schema_GetConnection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConnection'
type Schema_GetConnection_Call struct {
	*mock.Call
}

// GetConnection is a helper method to define mock.On call
func (_e *Schema_Expecter) GetConnection() *Schema_GetConnection_Call {
	return &Schema_GetConnection_Call{Call: _e.mock.On("GetConnection")}
}

func (_c *Schema_GetConnection_Call) Run(run func()) *Schema_GetConnection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_GetConnection_Call) Return(_a0 string) *Schema_GetConnection_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_GetConnection_Call) RunAndReturn(run func() string) *Schema_GetConnection_Call {
	_c.Call.Return(run)
	return _c
}

// GetForeignKeys provides a mock function with given fields: table
func (_m *Schema) GetForeignKeys(table string) ([]schema.ForeignKey, error) {
	ret := _m.Called(table)

	if len(ret) == 0 {
		panic("no return value specified for GetForeignKeys")
	}

	var r0 []schema.ForeignKey
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]schema.ForeignKey, error)); ok {
		return rf(table)
	}
	if rf, ok := ret.Get(0).(func(string) []schema.ForeignKey); ok {
		r0 = rf(table)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]schema.ForeignKey)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(table)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Schema_GetForeignKeys_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetForeignKeys'
type Schema_GetForeignKeys_Call struct {
	*mock.Call
}

// GetForeignKeys is a helper method to define mock.On call
//   - table string
func (_e *Schema_Expecter) GetForeignKeys(table interface{}) *Schema_GetForeignKeys_Call {
	return &Schema_GetForeignKeys_Call{Call: _e.mock.On("GetForeignKeys", table)}
}

func (_c *Schema_GetForeignKeys_Call) Run(run func(table string)) *Schema_GetForeignKeys_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_GetForeignKeys_Call) Return(_a0 []schema.ForeignKey, _a1 error) *Schema_GetForeignKeys_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Schema_GetForeignKeys_Call) RunAndReturn(run func(string) ([]schema.ForeignKey, error)) *Schema_GetForeignKeys_Call {
	_c.Call.Return(run)
	return _c
}

// GetIndexListing provides a mock function with given fields: table
func (_m *Schema) GetIndexListing(table string) []string {
	ret := _m.Called(table)

	if len(ret) == 0 {
		panic("no return value specified for GetIndexListing")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(table)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Schema_GetIndexListing_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIndexListing'
type Schema_GetIndexListing_Call struct {
	*mock.Call
}

// GetIndexListing is a helper method to define mock.On call
//   - table string
func (_e *Schema_Expecter) GetIndexListing(table interface{}) *Schema_GetIndexListing_Call {
	return &Schema_GetIndexListing_Call{Call: _e.mock.On("GetIndexListing", table)}
}

func (_c *Schema_GetIndexListing_Call) Run(run func(table string)) *Schema_GetIndexListing_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_GetIndexListing_Call) Return(_a0 []string) *Schema_GetIndexListing_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_GetIndexListing_Call) RunAndReturn(run func(string) []string) *Schema_GetIndexListing_Call {
	_c.Call.Return(run)
	return _c
}

// GetIndexes provides a mock function with given fields: table
func (_m *Schema) GetIndexes(table string) ([]schema.Index, error) {
	ret := _m.Called(table)

	if len(ret) == 0 {
		panic("no return value specified for GetIndexes")
	}

	var r0 []schema.Index
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]schema.Index, error)); ok {
		return rf(table)
	}
	if rf, ok := ret.Get(0).(func(string) []schema.Index); ok {
		r0 = rf(table)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]schema.Index)
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

func (_c *Schema_GetIndexes_Call) Return(_a0 []schema.Index, _a1 error) *Schema_GetIndexes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Schema_GetIndexes_Call) RunAndReturn(run func(string) ([]schema.Index, error)) *Schema_GetIndexes_Call {
	_c.Call.Return(run)
	return _c
}

// GetTableListing provides a mock function with given fields:
func (_m *Schema) GetTableListing() []string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetTableListing")
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

// Schema_GetTableListing_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTableListing'
type Schema_GetTableListing_Call struct {
	*mock.Call
}

// GetTableListing is a helper method to define mock.On call
func (_e *Schema_Expecter) GetTableListing() *Schema_GetTableListing_Call {
	return &Schema_GetTableListing_Call{Call: _e.mock.On("GetTableListing")}
}

func (_c *Schema_GetTableListing_Call) Run(run func()) *Schema_GetTableListing_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_GetTableListing_Call) Return(_a0 []string) *Schema_GetTableListing_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_GetTableListing_Call) RunAndReturn(run func() []string) *Schema_GetTableListing_Call {
	_c.Call.Return(run)
	return _c
}

// GetTables provides a mock function with given fields:
func (_m *Schema) GetTables() ([]schema.Table, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetTables")
	}

	var r0 []schema.Table
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]schema.Table, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []schema.Table); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]schema.Table)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Schema_GetTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTables'
type Schema_GetTables_Call struct {
	*mock.Call
}

// GetTables is a helper method to define mock.On call
func (_e *Schema_Expecter) GetTables() *Schema_GetTables_Call {
	return &Schema_GetTables_Call{Call: _e.mock.On("GetTables")}
}

func (_c *Schema_GetTables_Call) Run(run func()) *Schema_GetTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_GetTables_Call) Return(_a0 []schema.Table, _a1 error) *Schema_GetTables_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Schema_GetTables_Call) RunAndReturn(run func() ([]schema.Table, error)) *Schema_GetTables_Call {
	_c.Call.Return(run)
	return _c
}

// GetTypes provides a mock function with given fields:
func (_m *Schema) GetTypes() ([]schema.Type, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetTypes")
	}

	var r0 []schema.Type
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]schema.Type, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []schema.Type); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]schema.Type)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Schema_GetTypes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTypes'
type Schema_GetTypes_Call struct {
	*mock.Call
}

// GetTypes is a helper method to define mock.On call
func (_e *Schema_Expecter) GetTypes() *Schema_GetTypes_Call {
	return &Schema_GetTypes_Call{Call: _e.mock.On("GetTypes")}
}

func (_c *Schema_GetTypes_Call) Run(run func()) *Schema_GetTypes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_GetTypes_Call) Return(_a0 []schema.Type, _a1 error) *Schema_GetTypes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Schema_GetTypes_Call) RunAndReturn(run func() ([]schema.Type, error)) *Schema_GetTypes_Call {
	_c.Call.Return(run)
	return _c
}

// GetViews provides a mock function with given fields:
func (_m *Schema) GetViews() ([]schema.View, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetViews")
	}

	var r0 []schema.View
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]schema.View, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []schema.View); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]schema.View)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Schema_GetViews_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetViews'
type Schema_GetViews_Call struct {
	*mock.Call
}

// GetViews is a helper method to define mock.On call
func (_e *Schema_Expecter) GetViews() *Schema_GetViews_Call {
	return &Schema_GetViews_Call{Call: _e.mock.On("GetViews")}
}

func (_c *Schema_GetViews_Call) Run(run func()) *Schema_GetViews_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_GetViews_Call) Return(_a0 []schema.View, _a1 error) *Schema_GetViews_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Schema_GetViews_Call) RunAndReturn(run func() ([]schema.View, error)) *Schema_GetViews_Call {
	_c.Call.Return(run)
	return _c
}

// HasColumn provides a mock function with given fields: table, column
func (_m *Schema) HasColumn(table string, column string) bool {
	ret := _m.Called(table, column)

	if len(ret) == 0 {
		panic("no return value specified for HasColumn")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(table, column)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Schema_HasColumn_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasColumn'
type Schema_HasColumn_Call struct {
	*mock.Call
}

// HasColumn is a helper method to define mock.On call
//   - table string
//   - column string
func (_e *Schema_Expecter) HasColumn(table interface{}, column interface{}) *Schema_HasColumn_Call {
	return &Schema_HasColumn_Call{Call: _e.mock.On("HasColumn", table, column)}
}

func (_c *Schema_HasColumn_Call) Run(run func(table string, column string)) *Schema_HasColumn_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Schema_HasColumn_Call) Return(_a0 bool) *Schema_HasColumn_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_HasColumn_Call) RunAndReturn(run func(string, string) bool) *Schema_HasColumn_Call {
	_c.Call.Return(run)
	return _c
}

// HasColumns provides a mock function with given fields: table, columns
func (_m *Schema) HasColumns(table string, columns []string) bool {
	ret := _m.Called(table, columns)

	if len(ret) == 0 {
		panic("no return value specified for HasColumns")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, []string) bool); ok {
		r0 = rf(table, columns)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Schema_HasColumns_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasColumns'
type Schema_HasColumns_Call struct {
	*mock.Call
}

// HasColumns is a helper method to define mock.On call
//   - table string
//   - columns []string
func (_e *Schema_Expecter) HasColumns(table interface{}, columns interface{}) *Schema_HasColumns_Call {
	return &Schema_HasColumns_Call{Call: _e.mock.On("HasColumns", table, columns)}
}

func (_c *Schema_HasColumns_Call) Run(run func(table string, columns []string)) *Schema_HasColumns_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]string))
	})
	return _c
}

func (_c *Schema_HasColumns_Call) Return(_a0 bool) *Schema_HasColumns_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_HasColumns_Call) RunAndReturn(run func(string, []string) bool) *Schema_HasColumns_Call {
	_c.Call.Return(run)
	return _c
}

// HasIndex provides a mock function with given fields: table, index
func (_m *Schema) HasIndex(table string, index string) bool {
	ret := _m.Called(table, index)

	if len(ret) == 0 {
		panic("no return value specified for HasIndex")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(table, index)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Schema_HasIndex_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasIndex'
type Schema_HasIndex_Call struct {
	*mock.Call
}

// HasIndex is a helper method to define mock.On call
//   - table string
//   - index string
func (_e *Schema_Expecter) HasIndex(table interface{}, index interface{}) *Schema_HasIndex_Call {
	return &Schema_HasIndex_Call{Call: _e.mock.On("HasIndex", table, index)}
}

func (_c *Schema_HasIndex_Call) Run(run func(table string, index string)) *Schema_HasIndex_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Schema_HasIndex_Call) Return(_a0 bool) *Schema_HasIndex_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_HasIndex_Call) RunAndReturn(run func(string, string) bool) *Schema_HasIndex_Call {
	_c.Call.Return(run)
	return _c
}

// HasTable provides a mock function with given fields: name
func (_m *Schema) HasTable(name string) bool {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for HasTable")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Schema_HasTable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasTable'
type Schema_HasTable_Call struct {
	*mock.Call
}

// HasTable is a helper method to define mock.On call
//   - name string
func (_e *Schema_Expecter) HasTable(name interface{}) *Schema_HasTable_Call {
	return &Schema_HasTable_Call{Call: _e.mock.On("HasTable", name)}
}

func (_c *Schema_HasTable_Call) Run(run func(name string)) *Schema_HasTable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_HasTable_Call) Return(_a0 bool) *Schema_HasTable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_HasTable_Call) RunAndReturn(run func(string) bool) *Schema_HasTable_Call {
	_c.Call.Return(run)
	return _c
}

// HasType provides a mock function with given fields: name
func (_m *Schema) HasType(name string) bool {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for HasType")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Schema_HasType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasType'
type Schema_HasType_Call struct {
	*mock.Call
}

// HasType is a helper method to define mock.On call
//   - name string
func (_e *Schema_Expecter) HasType(name interface{}) *Schema_HasType_Call {
	return &Schema_HasType_Call{Call: _e.mock.On("HasType", name)}
}

func (_c *Schema_HasType_Call) Run(run func(name string)) *Schema_HasType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_HasType_Call) Return(_a0 bool) *Schema_HasType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_HasType_Call) RunAndReturn(run func(string) bool) *Schema_HasType_Call {
	_c.Call.Return(run)
	return _c
}

// HasView provides a mock function with given fields: name
func (_m *Schema) HasView(name string) bool {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for HasView")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Schema_HasView_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasView'
type Schema_HasView_Call struct {
	*mock.Call
}

// HasView is a helper method to define mock.On call
//   - name string
func (_e *Schema_Expecter) HasView(name interface{}) *Schema_HasView_Call {
	return &Schema_HasView_Call{Call: _e.mock.On("HasView", name)}
}

func (_c *Schema_HasView_Call) Run(run func(name string)) *Schema_HasView_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_HasView_Call) Return(_a0 bool) *Schema_HasView_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_HasView_Call) RunAndReturn(run func(string) bool) *Schema_HasView_Call {
	_c.Call.Return(run)
	return _c
}

// Migrations provides a mock function with given fields:
func (_m *Schema) Migrations() []schema.Migration {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Migrations")
	}

	var r0 []schema.Migration
	if rf, ok := ret.Get(0).(func() []schema.Migration); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]schema.Migration)
		}
	}

	return r0
}

// Schema_Migrations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Migrations'
type Schema_Migrations_Call struct {
	*mock.Call
}

// Migrations is a helper method to define mock.On call
func (_e *Schema_Expecter) Migrations() *Schema_Migrations_Call {
	return &Schema_Migrations_Call{Call: _e.mock.On("Migrations")}
}

func (_c *Schema_Migrations_Call) Run(run func()) *Schema_Migrations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Schema_Migrations_Call) Return(_a0 []schema.Migration) *Schema_Migrations_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_Migrations_Call) RunAndReturn(run func() []schema.Migration) *Schema_Migrations_Call {
	_c.Call.Return(run)
	return _c
}

// Orm provides a mock function with given fields:
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

// Register provides a mock function with given fields: _a0
func (_m *Schema) Register(_a0 []schema.Migration) {
	_m.Called(_a0)
}

// Schema_Register_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Register'
type Schema_Register_Call struct {
	*mock.Call
}

// Register is a helper method to define mock.On call
//   - _a0 []schema.Migration
func (_e *Schema_Expecter) Register(_a0 interface{}) *Schema_Register_Call {
	return &Schema_Register_Call{Call: _e.mock.On("Register", _a0)}
}

func (_c *Schema_Register_Call) Run(run func(_a0 []schema.Migration)) *Schema_Register_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]schema.Migration))
	})
	return _c
}

func (_c *Schema_Register_Call) Return() *Schema_Register_Call {
	_c.Call.Return()
	return _c
}

func (_c *Schema_Register_Call) RunAndReturn(run func([]schema.Migration)) *Schema_Register_Call {
	_c.Call.Return(run)
	return _c
}

// Rename provides a mock function with given fields: from, to
func (_m *Schema) Rename(from string, to string) error {
	ret := _m.Called(from, to)

	if len(ret) == 0 {
		panic("no return value specified for Rename")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(from, to)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_Rename_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rename'
type Schema_Rename_Call struct {
	*mock.Call
}

// Rename is a helper method to define mock.On call
//   - from string
//   - to string
func (_e *Schema_Expecter) Rename(from interface{}, to interface{}) *Schema_Rename_Call {
	return &Schema_Rename_Call{Call: _e.mock.On("Rename", from, to)}
}

func (_c *Schema_Rename_Call) Run(run func(from string, to string)) *Schema_Rename_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Schema_Rename_Call) Return(_a0 error) *Schema_Rename_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_Rename_Call) RunAndReturn(run func(string, string) error) *Schema_Rename_Call {
	_c.Call.Return(run)
	return _c
}

// SetConnection provides a mock function with given fields: name
func (_m *Schema) SetConnection(name string) {
	_m.Called(name)
}

// Schema_SetConnection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetConnection'
type Schema_SetConnection_Call struct {
	*mock.Call
}

// SetConnection is a helper method to define mock.On call
//   - name string
func (_e *Schema_Expecter) SetConnection(name interface{}) *Schema_SetConnection_Call {
	return &Schema_SetConnection_Call{Call: _e.mock.On("SetConnection", name)}
}

func (_c *Schema_SetConnection_Call) Run(run func(name string)) *Schema_SetConnection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_SetConnection_Call) Return() *Schema_SetConnection_Call {
	_c.Call.Return()
	return _c
}

func (_c *Schema_SetConnection_Call) RunAndReturn(run func(string)) *Schema_SetConnection_Call {
	_c.Call.Return(run)
	return _c
}

// Sql provides a mock function with given fields: sql
func (_m *Schema) Sql(sql string) error {
	ret := _m.Called(sql)

	if len(ret) == 0 {
		panic("no return value specified for Sql")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(sql)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_Sql_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Sql'
type Schema_Sql_Call struct {
	*mock.Call
}

// Sql is a helper method to define mock.On call
//   - sql string
func (_e *Schema_Expecter) Sql(sql interface{}) *Schema_Sql_Call {
	return &Schema_Sql_Call{Call: _e.mock.On("Sql", sql)}
}

func (_c *Schema_Sql_Call) Run(run func(sql string)) *Schema_Sql_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Schema_Sql_Call) Return(_a0 error) *Schema_Sql_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_Sql_Call) RunAndReturn(run func(string) error) *Schema_Sql_Call {
	_c.Call.Return(run)
	return _c
}

// Table provides a mock function with given fields: table, callback
func (_m *Schema) Table(table string, callback func(schema.Blueprint)) error {
	ret := _m.Called(table, callback)

	if len(ret) == 0 {
		panic("no return value specified for Table")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, func(schema.Blueprint)) error); ok {
		r0 = rf(table, callback)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Schema_Table_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Table'
type Schema_Table_Call struct {
	*mock.Call
}

// Table is a helper method to define mock.On call
//   - table string
//   - callback func(schema.Blueprint)
func (_e *Schema_Expecter) Table(table interface{}, callback interface{}) *Schema_Table_Call {
	return &Schema_Table_Call{Call: _e.mock.On("Table", table, callback)}
}

func (_c *Schema_Table_Call) Run(run func(table string, callback func(schema.Blueprint))) *Schema_Table_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(func(schema.Blueprint)))
	})
	return _c
}

func (_c *Schema_Table_Call) Return(_a0 error) *Schema_Table_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Schema_Table_Call) RunAndReturn(run func(string, func(schema.Blueprint)) error) *Schema_Table_Call {
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
