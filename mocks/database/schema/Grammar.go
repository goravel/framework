// Code generated by mockery. DO NOT EDIT.

package schema

import (
	orm "github.com/goravel/framework/contracts/database/orm"
	schema "github.com/goravel/framework/contracts/database/schema"
	mock "github.com/stretchr/testify/mock"
)

// Grammar is an autogenerated mock type for the Grammar type
type Grammar struct {
	mock.Mock
}

type Grammar_Expecter struct {
	mock *mock.Mock
}

func (_m *Grammar) EXPECT() *Grammar_Expecter {
	return &Grammar_Expecter{mock: &_m.Mock}
}

// CompileAdd provides a mock function with given fields: blueprint
func (_m *Grammar) CompileAdd(blueprint schema.Blueprint) string {
	ret := _m.Called(blueprint)

	if len(ret) == 0 {
		panic("no return value specified for CompileAdd")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(schema.Blueprint) string); ok {
		r0 = rf(blueprint)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileAdd_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileAdd'
type Grammar_CompileAdd_Call struct {
	*mock.Call
}

// CompileAdd is a helper method to define mock.On call
//   - blueprint schema.Blueprint
func (_e *Grammar_Expecter) CompileAdd(blueprint interface{}) *Grammar_CompileAdd_Call {
	return &Grammar_CompileAdd_Call{Call: _e.mock.On("CompileAdd", blueprint)}
}

func (_c *Grammar_CompileAdd_Call) Run(run func(blueprint schema.Blueprint)) *Grammar_CompileAdd_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(schema.Blueprint))
	})
	return _c
}

func (_c *Grammar_CompileAdd_Call) Return(_a0 string) *Grammar_CompileAdd_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileAdd_Call) RunAndReturn(run func(schema.Blueprint) string) *Grammar_CompileAdd_Call {
	_c.Call.Return(run)
	return _c
}

// CompileChange provides a mock function with given fields: blueprint
func (_m *Grammar) CompileChange(blueprint schema.Blueprint) string {
	ret := _m.Called(blueprint)

	if len(ret) == 0 {
		panic("no return value specified for CompileChange")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(schema.Blueprint) string); ok {
		r0 = rf(blueprint)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileChange_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileChange'
type Grammar_CompileChange_Call struct {
	*mock.Call
}

// CompileChange is a helper method to define mock.On call
//   - blueprint schema.Blueprint
func (_e *Grammar_Expecter) CompileChange(blueprint interface{}) *Grammar_CompileChange_Call {
	return &Grammar_CompileChange_Call{Call: _e.mock.On("CompileChange", blueprint)}
}

func (_c *Grammar_CompileChange_Call) Run(run func(blueprint schema.Blueprint)) *Grammar_CompileChange_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(schema.Blueprint))
	})
	return _c
}

func (_c *Grammar_CompileChange_Call) Return(_a0 string) *Grammar_CompileChange_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileChange_Call) RunAndReturn(run func(schema.Blueprint) string) *Grammar_CompileChange_Call {
	_c.Call.Return(run)
	return _c
}

// CompileCreate provides a mock function with given fields: blueprint, query
func (_m *Grammar) CompileCreate(blueprint schema.Blueprint, query orm.Query) string {
	ret := _m.Called(blueprint, query)

	if len(ret) == 0 {
		panic("no return value specified for CompileCreate")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(schema.Blueprint, orm.Query) string); ok {
		r0 = rf(blueprint, query)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileCreate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileCreate'
type Grammar_CompileCreate_Call struct {
	*mock.Call
}

// CompileCreate is a helper method to define mock.On call
//   - blueprint schema.Blueprint
//   - query orm.Query
func (_e *Grammar_Expecter) CompileCreate(blueprint interface{}, query interface{}) *Grammar_CompileCreate_Call {
	return &Grammar_CompileCreate_Call{Call: _e.mock.On("CompileCreate", blueprint, query)}
}

func (_c *Grammar_CompileCreate_Call) Run(run func(blueprint schema.Blueprint, query orm.Query)) *Grammar_CompileCreate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(schema.Blueprint), args[1].(orm.Query))
	})
	return _c
}

func (_c *Grammar_CompileCreate_Call) Return(_a0 string) *Grammar_CompileCreate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileCreate_Call) RunAndReturn(run func(schema.Blueprint, orm.Query) string) *Grammar_CompileCreate_Call {
	_c.Call.Return(run)
	return _c
}

// CompileDropAllDomains provides a mock function with given fields: domains
func (_m *Grammar) CompileDropAllDomains(domains []string) string {
	ret := _m.Called(domains)

	if len(ret) == 0 {
		panic("no return value specified for CompileDropAllDomains")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func([]string) string); ok {
		r0 = rf(domains)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileDropAllDomains_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileDropAllDomains'
type Grammar_CompileDropAllDomains_Call struct {
	*mock.Call
}

// CompileDropAllDomains is a helper method to define mock.On call
//   - domains []string
func (_e *Grammar_Expecter) CompileDropAllDomains(domains interface{}) *Grammar_CompileDropAllDomains_Call {
	return &Grammar_CompileDropAllDomains_Call{Call: _e.mock.On("CompileDropAllDomains", domains)}
}

func (_c *Grammar_CompileDropAllDomains_Call) Run(run func(domains []string)) *Grammar_CompileDropAllDomains_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *Grammar_CompileDropAllDomains_Call) Return(_a0 string) *Grammar_CompileDropAllDomains_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileDropAllDomains_Call) RunAndReturn(run func([]string) string) *Grammar_CompileDropAllDomains_Call {
	_c.Call.Return(run)
	return _c
}

// CompileDropAllTables provides a mock function with given fields: tables
func (_m *Grammar) CompileDropAllTables(tables []string) string {
	ret := _m.Called(tables)

	if len(ret) == 0 {
		panic("no return value specified for CompileDropAllTables")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func([]string) string); ok {
		r0 = rf(tables)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileDropAllTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileDropAllTables'
type Grammar_CompileDropAllTables_Call struct {
	*mock.Call
}

// CompileDropAllTables is a helper method to define mock.On call
//   - tables []string
func (_e *Grammar_Expecter) CompileDropAllTables(tables interface{}) *Grammar_CompileDropAllTables_Call {
	return &Grammar_CompileDropAllTables_Call{Call: _e.mock.On("CompileDropAllTables", tables)}
}

func (_c *Grammar_CompileDropAllTables_Call) Run(run func(tables []string)) *Grammar_CompileDropAllTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *Grammar_CompileDropAllTables_Call) Return(_a0 string) *Grammar_CompileDropAllTables_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileDropAllTables_Call) RunAndReturn(run func([]string) string) *Grammar_CompileDropAllTables_Call {
	_c.Call.Return(run)
	return _c
}

// CompileDropAllTypes provides a mock function with given fields: types
func (_m *Grammar) CompileDropAllTypes(types []string) string {
	ret := _m.Called(types)

	if len(ret) == 0 {
		panic("no return value specified for CompileDropAllTypes")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func([]string) string); ok {
		r0 = rf(types)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileDropAllTypes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileDropAllTypes'
type Grammar_CompileDropAllTypes_Call struct {
	*mock.Call
}

// CompileDropAllTypes is a helper method to define mock.On call
//   - types []string
func (_e *Grammar_Expecter) CompileDropAllTypes(types interface{}) *Grammar_CompileDropAllTypes_Call {
	return &Grammar_CompileDropAllTypes_Call{Call: _e.mock.On("CompileDropAllTypes", types)}
}

func (_c *Grammar_CompileDropAllTypes_Call) Run(run func(types []string)) *Grammar_CompileDropAllTypes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *Grammar_CompileDropAllTypes_Call) Return(_a0 string) *Grammar_CompileDropAllTypes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileDropAllTypes_Call) RunAndReturn(run func([]string) string) *Grammar_CompileDropAllTypes_Call {
	_c.Call.Return(run)
	return _c
}

// CompileDropAllViews provides a mock function with given fields: views
func (_m *Grammar) CompileDropAllViews(views []string) string {
	ret := _m.Called(views)

	if len(ret) == 0 {
		panic("no return value specified for CompileDropAllViews")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func([]string) string); ok {
		r0 = rf(views)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileDropAllViews_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileDropAllViews'
type Grammar_CompileDropAllViews_Call struct {
	*mock.Call
}

// CompileDropAllViews is a helper method to define mock.On call
//   - views []string
func (_e *Grammar_Expecter) CompileDropAllViews(views interface{}) *Grammar_CompileDropAllViews_Call {
	return &Grammar_CompileDropAllViews_Call{Call: _e.mock.On("CompileDropAllViews", views)}
}

func (_c *Grammar_CompileDropAllViews_Call) Run(run func(views []string)) *Grammar_CompileDropAllViews_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *Grammar_CompileDropAllViews_Call) Return(_a0 string) *Grammar_CompileDropAllViews_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileDropAllViews_Call) RunAndReturn(run func([]string) string) *Grammar_CompileDropAllViews_Call {
	_c.Call.Return(run)
	return _c
}

// CompileDropIfExists provides a mock function with given fields: blueprint
func (_m *Grammar) CompileDropIfExists(blueprint schema.Blueprint) string {
	ret := _m.Called(blueprint)

	if len(ret) == 0 {
		panic("no return value specified for CompileDropIfExists")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(schema.Blueprint) string); ok {
		r0 = rf(blueprint)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileDropIfExists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileDropIfExists'
type Grammar_CompileDropIfExists_Call struct {
	*mock.Call
}

// CompileDropIfExists is a helper method to define mock.On call
//   - blueprint schema.Blueprint
func (_e *Grammar_Expecter) CompileDropIfExists(blueprint interface{}) *Grammar_CompileDropIfExists_Call {
	return &Grammar_CompileDropIfExists_Call{Call: _e.mock.On("CompileDropIfExists", blueprint)}
}

func (_c *Grammar_CompileDropIfExists_Call) Run(run func(blueprint schema.Blueprint)) *Grammar_CompileDropIfExists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(schema.Blueprint))
	})
	return _c
}

func (_c *Grammar_CompileDropIfExists_Call) Return(_a0 string) *Grammar_CompileDropIfExists_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileDropIfExists_Call) RunAndReturn(run func(schema.Blueprint) string) *Grammar_CompileDropIfExists_Call {
	_c.Call.Return(run)
	return _c
}

// CompileTables provides a mock function with given fields: database
func (_m *Grammar) CompileTables(database string) string {
	ret := _m.Called(database)

	if len(ret) == 0 {
		panic("no return value specified for CompileTables")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(database)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileTables'
type Grammar_CompileTables_Call struct {
	*mock.Call
}

// CompileTables is a helper method to define mock.On call
//   - database string
func (_e *Grammar_Expecter) CompileTables(database interface{}) *Grammar_CompileTables_Call {
	return &Grammar_CompileTables_Call{Call: _e.mock.On("CompileTables", database)}
}

func (_c *Grammar_CompileTables_Call) Run(run func(database string)) *Grammar_CompileTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Grammar_CompileTables_Call) Return(_a0 string) *Grammar_CompileTables_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileTables_Call) RunAndReturn(run func(string) string) *Grammar_CompileTables_Call {
	_c.Call.Return(run)
	return _c
}

// CompileTypes provides a mock function with given fields:
func (_m *Grammar) CompileTypes() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CompileTypes")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileTypes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileTypes'
type Grammar_CompileTypes_Call struct {
	*mock.Call
}

// CompileTypes is a helper method to define mock.On call
func (_e *Grammar_Expecter) CompileTypes() *Grammar_CompileTypes_Call {
	return &Grammar_CompileTypes_Call{Call: _e.mock.On("CompileTypes")}
}

func (_c *Grammar_CompileTypes_Call) Run(run func()) *Grammar_CompileTypes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Grammar_CompileTypes_Call) Return(_a0 string) *Grammar_CompileTypes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileTypes_Call) RunAndReturn(run func() string) *Grammar_CompileTypes_Call {
	_c.Call.Return(run)
	return _c
}

// CompileViews provides a mock function with given fields:
func (_m *Grammar) CompileViews() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CompileViews")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_CompileViews_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompileViews'
type Grammar_CompileViews_Call struct {
	*mock.Call
}

// CompileViews is a helper method to define mock.On call
func (_e *Grammar_Expecter) CompileViews() *Grammar_CompileViews_Call {
	return &Grammar_CompileViews_Call{Call: _e.mock.On("CompileViews")}
}

func (_c *Grammar_CompileViews_Call) Run(run func()) *Grammar_CompileViews_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Grammar_CompileViews_Call) Return(_a0 string) *Grammar_CompileViews_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_CompileViews_Call) RunAndReturn(run func() string) *Grammar_CompileViews_Call {
	_c.Call.Return(run)
	return _c
}

// GetAttributeCommands provides a mock function with given fields:
func (_m *Grammar) GetAttributeCommands() []string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetAttributeCommands")
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

// Grammar_GetAttributeCommands_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAttributeCommands'
type Grammar_GetAttributeCommands_Call struct {
	*mock.Call
}

// GetAttributeCommands is a helper method to define mock.On call
func (_e *Grammar_Expecter) GetAttributeCommands() *Grammar_GetAttributeCommands_Call {
	return &Grammar_GetAttributeCommands_Call{Call: _e.mock.On("GetAttributeCommands")}
}

func (_c *Grammar_GetAttributeCommands_Call) Run(run func()) *Grammar_GetAttributeCommands_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Grammar_GetAttributeCommands_Call) Return(_a0 []string) *Grammar_GetAttributeCommands_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_GetAttributeCommands_Call) RunAndReturn(run func() []string) *Grammar_GetAttributeCommands_Call {
	_c.Call.Return(run)
	return _c
}

// GetModifiers provides a mock function with given fields:
func (_m *Grammar) GetModifiers() []func(schema.Blueprint, schema.ColumnDefinition) string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetModifiers")
	}

	var r0 []func(schema.Blueprint, schema.ColumnDefinition) string
	if rf, ok := ret.Get(0).(func() []func(schema.Blueprint, schema.ColumnDefinition) string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]func(schema.Blueprint, schema.ColumnDefinition) string)
		}
	}

	return r0
}

// Grammar_GetModifiers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetModifiers'
type Grammar_GetModifiers_Call struct {
	*mock.Call
}

// GetModifiers is a helper method to define mock.On call
func (_e *Grammar_Expecter) GetModifiers() *Grammar_GetModifiers_Call {
	return &Grammar_GetModifiers_Call{Call: _e.mock.On("GetModifiers")}
}

func (_c *Grammar_GetModifiers_Call) Run(run func()) *Grammar_GetModifiers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Grammar_GetModifiers_Call) Return(_a0 []func(schema.Blueprint, schema.ColumnDefinition) string) *Grammar_GetModifiers_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_GetModifiers_Call) RunAndReturn(run func() []func(schema.Blueprint, schema.ColumnDefinition) string) *Grammar_GetModifiers_Call {
	_c.Call.Return(run)
	return _c
}

// TypeBigInteger provides a mock function with given fields: column
func (_m *Grammar) TypeBigInteger(column schema.ColumnDefinition) string {
	ret := _m.Called(column)

	if len(ret) == 0 {
		panic("no return value specified for TypeBigInteger")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(schema.ColumnDefinition) string); ok {
		r0 = rf(column)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_TypeBigInteger_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TypeBigInteger'
type Grammar_TypeBigInteger_Call struct {
	*mock.Call
}

// TypeBigInteger is a helper method to define mock.On call
//   - column schema.ColumnDefinition
func (_e *Grammar_Expecter) TypeBigInteger(column interface{}) *Grammar_TypeBigInteger_Call {
	return &Grammar_TypeBigInteger_Call{Call: _e.mock.On("TypeBigInteger", column)}
}

func (_c *Grammar_TypeBigInteger_Call) Run(run func(column schema.ColumnDefinition)) *Grammar_TypeBigInteger_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(schema.ColumnDefinition))
	})
	return _c
}

func (_c *Grammar_TypeBigInteger_Call) Return(_a0 string) *Grammar_TypeBigInteger_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_TypeBigInteger_Call) RunAndReturn(run func(schema.ColumnDefinition) string) *Grammar_TypeBigInteger_Call {
	_c.Call.Return(run)
	return _c
}

// TypeInteger provides a mock function with given fields: column
func (_m *Grammar) TypeInteger(column schema.ColumnDefinition) string {
	ret := _m.Called(column)

	if len(ret) == 0 {
		panic("no return value specified for TypeInteger")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(schema.ColumnDefinition) string); ok {
		r0 = rf(column)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_TypeInteger_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TypeInteger'
type Grammar_TypeInteger_Call struct {
	*mock.Call
}

// TypeInteger is a helper method to define mock.On call
//   - column schema.ColumnDefinition
func (_e *Grammar_Expecter) TypeInteger(column interface{}) *Grammar_TypeInteger_Call {
	return &Grammar_TypeInteger_Call{Call: _e.mock.On("TypeInteger", column)}
}

func (_c *Grammar_TypeInteger_Call) Run(run func(column schema.ColumnDefinition)) *Grammar_TypeInteger_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(schema.ColumnDefinition))
	})
	return _c
}

func (_c *Grammar_TypeInteger_Call) Return(_a0 string) *Grammar_TypeInteger_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_TypeInteger_Call) RunAndReturn(run func(schema.ColumnDefinition) string) *Grammar_TypeInteger_Call {
	_c.Call.Return(run)
	return _c
}

// TypeString provides a mock function with given fields: column
func (_m *Grammar) TypeString(column schema.ColumnDefinition) string {
	ret := _m.Called(column)

	if len(ret) == 0 {
		panic("no return value specified for TypeString")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(schema.ColumnDefinition) string); ok {
		r0 = rf(column)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Grammar_TypeString_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TypeString'
type Grammar_TypeString_Call struct {
	*mock.Call
}

// TypeString is a helper method to define mock.On call
//   - column schema.ColumnDefinition
func (_e *Grammar_Expecter) TypeString(column interface{}) *Grammar_TypeString_Call {
	return &Grammar_TypeString_Call{Call: _e.mock.On("TypeString", column)}
}

func (_c *Grammar_TypeString_Call) Run(run func(column schema.ColumnDefinition)) *Grammar_TypeString_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(schema.ColumnDefinition))
	})
	return _c
}

func (_c *Grammar_TypeString_Call) Return(_a0 string) *Grammar_TypeString_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Grammar_TypeString_Call) RunAndReturn(run func(schema.ColumnDefinition) string) *Grammar_TypeString_Call {
	_c.Call.Return(run)
	return _c
}

// NewGrammar creates a new instance of Grammar. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGrammar(t interface {
	mock.TestingT
	Cleanup(func())
}) *Grammar {
	mock := &Grammar{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
