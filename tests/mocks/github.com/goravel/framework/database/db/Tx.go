// Code generated by mockery. DO NOT EDIT.

package db

import (
	db "github.com/goravel/framework/contracts/database/db"
	mock "github.com/stretchr/testify/mock"
)

// Tx is an autogenerated mock type for the Tx type
type Tx struct {
	mock.Mock
}

type Tx_Expecter struct {
	mock *mock.Mock
}

func (_m *Tx) EXPECT() *Tx_Expecter {
	return &Tx_Expecter{mock: &_m.Mock}
}

// Commit provides a mock function with no fields
func (_m *Tx) Commit() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Commit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Tx_Commit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Commit'
type Tx_Commit_Call struct {
	*mock.Call
}

// Commit is a helper method to define mock.On call
func (_e *Tx_Expecter) Commit() *Tx_Commit_Call {
	return &Tx_Commit_Call{Call: _e.mock.On("Commit")}
}

func (_c *Tx_Commit_Call) Run(run func()) *Tx_Commit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Tx_Commit_Call) Return(_a0 error) *Tx_Commit_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Tx_Commit_Call) RunAndReturn(run func() error) *Tx_Commit_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: sql, args
func (_m *Tx) Delete(sql string, args ...interface{}) (*db.Result, error) {
	var _ca []interface{}
	_ca = append(_ca, sql)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 *db.Result
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...interface{}) (*db.Result, error)); ok {
		return rf(sql, args...)
	}
	if rf, ok := ret.Get(0).(func(string, ...interface{}) *db.Result); ok {
		r0 = rf(sql, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.Result)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...interface{}) error); ok {
		r1 = rf(sql, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Tx_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type Tx_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - sql string
//   - args ...interface{}
func (_e *Tx_Expecter) Delete(sql interface{}, args ...interface{}) *Tx_Delete_Call {
	return &Tx_Delete_Call{Call: _e.mock.On("Delete",
		append([]interface{}{sql}, args...)...)}
}

func (_c *Tx_Delete_Call) Run(run func(sql string, args ...interface{})) *Tx_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Tx_Delete_Call) Return(_a0 *db.Result, _a1 error) *Tx_Delete_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Tx_Delete_Call) RunAndReturn(run func(string, ...interface{}) (*db.Result, error)) *Tx_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Insert provides a mock function with given fields: sql, args
func (_m *Tx) Insert(sql string, args ...interface{}) (*db.Result, error) {
	var _ca []interface{}
	_ca = append(_ca, sql)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Insert")
	}

	var r0 *db.Result
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...interface{}) (*db.Result, error)); ok {
		return rf(sql, args...)
	}
	if rf, ok := ret.Get(0).(func(string, ...interface{}) *db.Result); ok {
		r0 = rf(sql, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.Result)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...interface{}) error); ok {
		r1 = rf(sql, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Tx_Insert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Insert'
type Tx_Insert_Call struct {
	*mock.Call
}

// Insert is a helper method to define mock.On call
//   - sql string
//   - args ...interface{}
func (_e *Tx_Expecter) Insert(sql interface{}, args ...interface{}) *Tx_Insert_Call {
	return &Tx_Insert_Call{Call: _e.mock.On("Insert",
		append([]interface{}{sql}, args...)...)}
}

func (_c *Tx_Insert_Call) Run(run func(sql string, args ...interface{})) *Tx_Insert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Tx_Insert_Call) Return(_a0 *db.Result, _a1 error) *Tx_Insert_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Tx_Insert_Call) RunAndReturn(run func(string, ...interface{}) (*db.Result, error)) *Tx_Insert_Call {
	_c.Call.Return(run)
	return _c
}

// Rollback provides a mock function with no fields
func (_m *Tx) Rollback() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Rollback")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Tx_Rollback_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rollback'
type Tx_Rollback_Call struct {
	*mock.Call
}

// Rollback is a helper method to define mock.On call
func (_e *Tx_Expecter) Rollback() *Tx_Rollback_Call {
	return &Tx_Rollback_Call{Call: _e.mock.On("Rollback")}
}

func (_c *Tx_Rollback_Call) Run(run func()) *Tx_Rollback_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Tx_Rollback_Call) Return(_a0 error) *Tx_Rollback_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Tx_Rollback_Call) RunAndReturn(run func() error) *Tx_Rollback_Call {
	_c.Call.Return(run)
	return _c
}

// Select provides a mock function with given fields: dest, sql, args
func (_m *Tx) Select(dest interface{}, sql string, args ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, dest, sql)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Select")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}, string, ...interface{}) error); ok {
		r0 = rf(dest, sql, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Tx_Select_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Select'
type Tx_Select_Call struct {
	*mock.Call
}

// Select is a helper method to define mock.On call
//   - dest interface{}
//   - sql string
//   - args ...interface{}
func (_e *Tx_Expecter) Select(dest interface{}, sql interface{}, args ...interface{}) *Tx_Select_Call {
	return &Tx_Select_Call{Call: _e.mock.On("Select",
		append([]interface{}{dest, sql}, args...)...)}
}

func (_c *Tx_Select_Call) Run(run func(dest interface{}, sql string, args ...interface{})) *Tx_Select_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(interface{}), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *Tx_Select_Call) Return(_a0 error) *Tx_Select_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Tx_Select_Call) RunAndReturn(run func(interface{}, string, ...interface{}) error) *Tx_Select_Call {
	_c.Call.Return(run)
	return _c
}

// Statement provides a mock function with given fields: sql, args
func (_m *Tx) Statement(sql string, args ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, sql)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Statement")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, ...interface{}) error); ok {
		r0 = rf(sql, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Tx_Statement_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Statement'
type Tx_Statement_Call struct {
	*mock.Call
}

// Statement is a helper method to define mock.On call
//   - sql string
//   - args ...interface{}
func (_e *Tx_Expecter) Statement(sql interface{}, args ...interface{}) *Tx_Statement_Call {
	return &Tx_Statement_Call{Call: _e.mock.On("Statement",
		append([]interface{}{sql}, args...)...)}
}

func (_c *Tx_Statement_Call) Run(run func(sql string, args ...interface{})) *Tx_Statement_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Tx_Statement_Call) Return(_a0 error) *Tx_Statement_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Tx_Statement_Call) RunAndReturn(run func(string, ...interface{}) error) *Tx_Statement_Call {
	_c.Call.Return(run)
	return _c
}

// Table provides a mock function with given fields: name
func (_m *Tx) Table(name string) db.Query {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Table")
	}

	var r0 db.Query
	if rf, ok := ret.Get(0).(func(string) db.Query); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(db.Query)
		}
	}

	return r0
}

// Tx_Table_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Table'
type Tx_Table_Call struct {
	*mock.Call
}

// Table is a helper method to define mock.On call
//   - name string
func (_e *Tx_Expecter) Table(name interface{}) *Tx_Table_Call {
	return &Tx_Table_Call{Call: _e.mock.On("Table", name)}
}

func (_c *Tx_Table_Call) Run(run func(name string)) *Tx_Table_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Tx_Table_Call) Return(_a0 db.Query) *Tx_Table_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Tx_Table_Call) RunAndReturn(run func(string) db.Query) *Tx_Table_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: sql, args
func (_m *Tx) Update(sql string, args ...interface{}) (*db.Result, error) {
	var _ca []interface{}
	_ca = append(_ca, sql)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 *db.Result
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...interface{}) (*db.Result, error)); ok {
		return rf(sql, args...)
	}
	if rf, ok := ret.Get(0).(func(string, ...interface{}) *db.Result); ok {
		r0 = rf(sql, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.Result)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...interface{}) error); ok {
		r1 = rf(sql, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Tx_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type Tx_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - sql string
//   - args ...interface{}
func (_e *Tx_Expecter) Update(sql interface{}, args ...interface{}) *Tx_Update_Call {
	return &Tx_Update_Call{Call: _e.mock.On("Update",
		append([]interface{}{sql}, args...)...)}
}

func (_c *Tx_Update_Call) Run(run func(sql string, args ...interface{})) *Tx_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Tx_Update_Call) Return(_a0 *db.Result, _a1 error) *Tx_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Tx_Update_Call) RunAndReturn(run func(string, ...interface{}) (*db.Result, error)) *Tx_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewTx creates a new instance of Tx. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTx(t interface {
	mock.TestingT
	Cleanup(func())
}) *Tx {
	mock := &Tx{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
