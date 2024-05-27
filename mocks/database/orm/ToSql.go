// Code generated by mockery. DO NOT EDIT.

package orm

import mock "github.com/stretchr/testify/mock"

// ToSql is an autogenerated mock type for the ToSql type
type ToSql struct {
	mock.Mock
}

type ToSql_Expecter struct {
	mock *mock.Mock
}

func (_m *ToSql) EXPECT() *ToSql_Expecter {
	return &ToSql_Expecter{mock: &_m.Mock}
}

// Count provides a mock function with given fields:
func (_m *ToSql) Count() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Count")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_Count_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Count'
type ToSql_Count_Call struct {
	*mock.Call
}

// Count is a helper method to define mock.On call
func (_e *ToSql_Expecter) Count() *ToSql_Count_Call {
	return &ToSql_Count_Call{Call: _e.mock.On("Count")}
}

func (_c *ToSql_Count_Call) Run(run func()) *ToSql_Count_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ToSql_Count_Call) Return(_a0 string) *ToSql_Count_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_Count_Call) RunAndReturn(run func() string) *ToSql_Count_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: value
func (_m *ToSql) Create(value interface{}) string {
	ret := _m.Called(value)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(interface{}) string); ok {
		r0 = rf(value)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type ToSql_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - value interface{}
func (_e *ToSql_Expecter) Create(value interface{}) *ToSql_Create_Call {
	return &ToSql_Create_Call{Call: _e.mock.On("Create", value)}
}

func (_c *ToSql_Create_Call) Run(run func(value interface{})) *ToSql_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *ToSql_Create_Call) Return(_a0 string) *ToSql_Create_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_Create_Call) RunAndReturn(run func(interface{}) string) *ToSql_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: value, conds
func (_m *ToSql) Delete(value interface{}, conds ...interface{}) string {
	var _ca []interface{}
	_ca = append(_ca, value)
	_ca = append(_ca, conds...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(interface{}, ...interface{}) string); ok {
		r0 = rf(value, conds...)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type ToSql_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - value interface{}
//   - conds ...interface{}
func (_e *ToSql_Expecter) Delete(value interface{}, conds ...interface{}) *ToSql_Delete_Call {
	return &ToSql_Delete_Call{Call: _e.mock.On("Delete",
		append([]interface{}{value}, conds...)...)}
}

func (_c *ToSql_Delete_Call) Run(run func(value interface{}, conds ...interface{})) *ToSql_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *ToSql_Delete_Call) Return(_a0 string) *ToSql_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_Delete_Call) RunAndReturn(run func(interface{}, ...interface{}) string) *ToSql_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Find provides a mock function with given fields: dest, conds
func (_m *ToSql) Find(dest interface{}, conds ...interface{}) string {
	var _ca []interface{}
	_ca = append(_ca, dest)
	_ca = append(_ca, conds...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Find")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(interface{}, ...interface{}) string); ok {
		r0 = rf(dest, conds...)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_Find_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Find'
type ToSql_Find_Call struct {
	*mock.Call
}

// Find is a helper method to define mock.On call
//   - dest interface{}
//   - conds ...interface{}
func (_e *ToSql_Expecter) Find(dest interface{}, conds ...interface{}) *ToSql_Find_Call {
	return &ToSql_Find_Call{Call: _e.mock.On("Find",
		append([]interface{}{dest}, conds...)...)}
}

func (_c *ToSql_Find_Call) Run(run func(dest interface{}, conds ...interface{})) *ToSql_Find_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *ToSql_Find_Call) Return(_a0 string) *ToSql_Find_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_Find_Call) RunAndReturn(run func(interface{}, ...interface{}) string) *ToSql_Find_Call {
	_c.Call.Return(run)
	return _c
}

// First provides a mock function with given fields: dest
func (_m *ToSql) First(dest interface{}) string {
	ret := _m.Called(dest)

	if len(ret) == 0 {
		panic("no return value specified for First")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(interface{}) string); ok {
		r0 = rf(dest)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_First_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'First'
type ToSql_First_Call struct {
	*mock.Call
}

// First is a helper method to define mock.On call
//   - dest interface{}
func (_e *ToSql_Expecter) First(dest interface{}) *ToSql_First_Call {
	return &ToSql_First_Call{Call: _e.mock.On("First", dest)}
}

func (_c *ToSql_First_Call) Run(run func(dest interface{})) *ToSql_First_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *ToSql_First_Call) Return(_a0 string) *ToSql_First_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_First_Call) RunAndReturn(run func(interface{}) string) *ToSql_First_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: dest
func (_m *ToSql) Get(dest interface{}) string {
	ret := _m.Called(dest)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(interface{}) string); ok {
		r0 = rf(dest)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type ToSql_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - dest interface{}
func (_e *ToSql_Expecter) Get(dest interface{}) *ToSql_Get_Call {
	return &ToSql_Get_Call{Call: _e.mock.On("Get", dest)}
}

func (_c *ToSql_Get_Call) Run(run func(dest interface{})) *ToSql_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *ToSql_Get_Call) Return(_a0 string) *ToSql_Get_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_Get_Call) RunAndReturn(run func(interface{}) string) *ToSql_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Pluck provides a mock function with given fields: column, dest
func (_m *ToSql) Pluck(column string, dest interface{}) string {
	ret := _m.Called(column, dest)

	if len(ret) == 0 {
		panic("no return value specified for Pluck")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string, interface{}) string); ok {
		r0 = rf(column, dest)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_Pluck_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Pluck'
type ToSql_Pluck_Call struct {
	*mock.Call
}

// Pluck is a helper method to define mock.On call
//   - column string
//   - dest interface{}
func (_e *ToSql_Expecter) Pluck(column interface{}, dest interface{}) *ToSql_Pluck_Call {
	return &ToSql_Pluck_Call{Call: _e.mock.On("Pluck", column, dest)}
}

func (_c *ToSql_Pluck_Call) Run(run func(column string, dest interface{})) *ToSql_Pluck_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(interface{}))
	})
	return _c
}

func (_c *ToSql_Pluck_Call) Return(_a0 string) *ToSql_Pluck_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_Pluck_Call) RunAndReturn(run func(string, interface{}) string) *ToSql_Pluck_Call {
	_c.Call.Return(run)
	return _c
}

// Save provides a mock function with given fields: value
func (_m *ToSql) Save(value interface{}) string {
	ret := _m.Called(value)

	if len(ret) == 0 {
		panic("no return value specified for Save")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(interface{}) string); ok {
		r0 = rf(value)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_Save_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Save'
type ToSql_Save_Call struct {
	*mock.Call
}

// Save is a helper method to define mock.On call
//   - value interface{}
func (_e *ToSql_Expecter) Save(value interface{}) *ToSql_Save_Call {
	return &ToSql_Save_Call{Call: _e.mock.On("Save", value)}
}

func (_c *ToSql_Save_Call) Run(run func(value interface{})) *ToSql_Save_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *ToSql_Save_Call) Return(_a0 string) *ToSql_Save_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_Save_Call) RunAndReturn(run func(interface{}) string) *ToSql_Save_Call {
	_c.Call.Return(run)
	return _c
}

// Sum provides a mock function with given fields: column, dest
func (_m *ToSql) Sum(column string, dest interface{}) string {
	ret := _m.Called(column, dest)

	if len(ret) == 0 {
		panic("no return value specified for Sum")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string, interface{}) string); ok {
		r0 = rf(column, dest)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_Sum_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Sum'
type ToSql_Sum_Call struct {
	*mock.Call
}

// Sum is a helper method to define mock.On call
//   - column string
//   - dest interface{}
func (_e *ToSql_Expecter) Sum(column interface{}, dest interface{}) *ToSql_Sum_Call {
	return &ToSql_Sum_Call{Call: _e.mock.On("Sum", column, dest)}
}

func (_c *ToSql_Sum_Call) Run(run func(column string, dest interface{})) *ToSql_Sum_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(interface{}))
	})
	return _c
}

func (_c *ToSql_Sum_Call) Return(_a0 string) *ToSql_Sum_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_Sum_Call) RunAndReturn(run func(string, interface{}) string) *ToSql_Sum_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: column, value
func (_m *ToSql) Update(column interface{}, value ...interface{}) string {
	var _ca []interface{}
	_ca = append(_ca, column)
	_ca = append(_ca, value...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(interface{}, ...interface{}) string); ok {
		r0 = rf(column, value...)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ToSql_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type ToSql_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - column interface{}
//   - value ...interface{}
func (_e *ToSql_Expecter) Update(column interface{}, value ...interface{}) *ToSql_Update_Call {
	return &ToSql_Update_Call{Call: _e.mock.On("Update",
		append([]interface{}{column}, value...)...)}
}

func (_c *ToSql_Update_Call) Run(run func(column interface{}, value ...interface{})) *ToSql_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *ToSql_Update_Call) Return(_a0 string) *ToSql_Update_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ToSql_Update_Call) RunAndReturn(run func(interface{}, ...interface{}) string) *ToSql_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewToSql creates a new instance of ToSql. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewToSql(t interface {
	mock.TestingT
	Cleanup(func())
}) *ToSql {
	mock := &ToSql{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}