// Code generated by mockery. DO NOT EDIT.

package db

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	sql "database/sql"

	sqlx "github.com/jmoiron/sqlx"
)

// Builder is an autogenerated mock type for the Builder type
type Builder struct {
	mock.Mock
}

type Builder_Expecter struct {
	mock *mock.Mock
}

func (_m *Builder) EXPECT() *Builder_Expecter {
	return &Builder_Expecter{mock: &_m.Mock}
}

// Beginx provides a mock function with no fields
func (_m *Builder) Beginx() (*sqlx.Tx, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Beginx")
	}

	var r0 *sqlx.Tx
	var r1 error
	if rf, ok := ret.Get(0).(func() (*sqlx.Tx, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *sqlx.Tx); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlx.Tx)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Builder_Beginx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Beginx'
type Builder_Beginx_Call struct {
	*mock.Call
}

// Beginx is a helper method to define mock.On call
func (_e *Builder_Expecter) Beginx() *Builder_Beginx_Call {
	return &Builder_Beginx_Call{Call: _e.mock.On("Beginx")}
}

func (_c *Builder_Beginx_Call) Run(run func()) *Builder_Beginx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Builder_Beginx_Call) Return(_a0 *sqlx.Tx, _a1 error) *Builder_Beginx_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Builder_Beginx_Call) RunAndReturn(run func() (*sqlx.Tx, error)) *Builder_Beginx_Call {
	_c.Call.Return(run)
	return _c
}

// ExecContext provides a mock function with given fields: ctx, query, args
func (_m *Builder) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ExecContext")
	}

	var r0 sql.Result
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) (sql.Result, error)); ok {
		return rf(ctx, query, args...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) sql.Result); ok {
		r0 = rf(ctx, query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sql.Result)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...interface{}) error); ok {
		r1 = rf(ctx, query, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Builder_ExecContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecContext'
type Builder_ExecContext_Call struct {
	*mock.Call
}

// ExecContext is a helper method to define mock.On call
//   - ctx context.Context
//   - query string
//   - args ...interface{}
func (_e *Builder_Expecter) ExecContext(ctx interface{}, query interface{}, args ...interface{}) *Builder_ExecContext_Call {
	return &Builder_ExecContext_Call{Call: _e.mock.On("ExecContext",
		append([]interface{}{ctx, query}, args...)...)}
}

func (_c *Builder_ExecContext_Call) Run(run func(ctx context.Context, query string, args ...interface{})) *Builder_ExecContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *Builder_ExecContext_Call) Return(_a0 sql.Result, _a1 error) *Builder_ExecContext_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Builder_ExecContext_Call) RunAndReturn(run func(context.Context, string, ...interface{}) (sql.Result, error)) *Builder_ExecContext_Call {
	_c.Call.Return(run)
	return _c
}

// Explain provides a mock function with given fields: _a0, args
func (_m *Builder) Explain(_a0 string, args ...interface{}) string {
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Explain")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string, ...interface{}) string); ok {
		r0 = rf(_a0, args...)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Builder_Explain_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Explain'
type Builder_Explain_Call struct {
	*mock.Call
}

// Explain is a helper method to define mock.On call
//   - _a0 string
//   - args ...interface{}
func (_e *Builder_Expecter) Explain(_a0 interface{}, args ...interface{}) *Builder_Explain_Call {
	return &Builder_Explain_Call{Call: _e.mock.On("Explain",
		append([]interface{}{_a0}, args...)...)}
}

func (_c *Builder_Explain_Call) Run(run func(_a0 string, args ...interface{})) *Builder_Explain_Call {
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

func (_c *Builder_Explain_Call) Return(_a0 string) *Builder_Explain_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Builder_Explain_Call) RunAndReturn(run func(string, ...interface{}) string) *Builder_Explain_Call {
	_c.Call.Return(run)
	return _c
}

// GetContext provides a mock function with given fields: ctx, dest, query, args
func (_m *Builder) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, ctx, dest, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetContext")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string, ...interface{}) error); ok {
		r0 = rf(ctx, dest, query, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Builder_GetContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetContext'
type Builder_GetContext_Call struct {
	*mock.Call
}

// GetContext is a helper method to define mock.On call
//   - ctx context.Context
//   - dest interface{}
//   - query string
//   - args ...interface{}
func (_e *Builder_Expecter) GetContext(ctx interface{}, dest interface{}, query interface{}, args ...interface{}) *Builder_GetContext_Call {
	return &Builder_GetContext_Call{Call: _e.mock.On("GetContext",
		append([]interface{}{ctx, dest, query}, args...)...)}
}

func (_c *Builder_GetContext_Call) Run(run func(ctx context.Context, dest interface{}, query string, args ...interface{})) *Builder_GetContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), args[1].(interface{}), args[2].(string), variadicArgs...)
	})
	return _c
}

func (_c *Builder_GetContext_Call) Return(_a0 error) *Builder_GetContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Builder_GetContext_Call) RunAndReturn(run func(context.Context, interface{}, string, ...interface{}) error) *Builder_GetContext_Call {
	_c.Call.Return(run)
	return _c
}

// QueryxContext provides a mock function with given fields: ctx, query, args
func (_m *Builder) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for QueryxContext")
	}

	var r0 *sqlx.Rows
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) (*sqlx.Rows, error)); ok {
		return rf(ctx, query, args...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) *sqlx.Rows); ok {
		r0 = rf(ctx, query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlx.Rows)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...interface{}) error); ok {
		r1 = rf(ctx, query, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Builder_QueryxContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'QueryxContext'
type Builder_QueryxContext_Call struct {
	*mock.Call
}

// QueryxContext is a helper method to define mock.On call
//   - ctx context.Context
//   - query string
//   - args ...interface{}
func (_e *Builder_Expecter) QueryxContext(ctx interface{}, query interface{}, args ...interface{}) *Builder_QueryxContext_Call {
	return &Builder_QueryxContext_Call{Call: _e.mock.On("QueryxContext",
		append([]interface{}{ctx, query}, args...)...)}
}

func (_c *Builder_QueryxContext_Call) Run(run func(ctx context.Context, query string, args ...interface{})) *Builder_QueryxContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *Builder_QueryxContext_Call) Return(_a0 *sqlx.Rows, _a1 error) *Builder_QueryxContext_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Builder_QueryxContext_Call) RunAndReturn(run func(context.Context, string, ...interface{}) (*sqlx.Rows, error)) *Builder_QueryxContext_Call {
	_c.Call.Return(run)
	return _c
}

// SelectContext provides a mock function with given fields: ctx, dest, query, args
func (_m *Builder) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, ctx, dest, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for SelectContext")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string, ...interface{}) error); ok {
		r0 = rf(ctx, dest, query, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Builder_SelectContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SelectContext'
type Builder_SelectContext_Call struct {
	*mock.Call
}

// SelectContext is a helper method to define mock.On call
//   - ctx context.Context
//   - dest interface{}
//   - query string
//   - args ...interface{}
func (_e *Builder_Expecter) SelectContext(ctx interface{}, dest interface{}, query interface{}, args ...interface{}) *Builder_SelectContext_Call {
	return &Builder_SelectContext_Call{Call: _e.mock.On("SelectContext",
		append([]interface{}{ctx, dest, query}, args...)...)}
}

func (_c *Builder_SelectContext_Call) Run(run func(ctx context.Context, dest interface{}, query string, args ...interface{})) *Builder_SelectContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), args[1].(interface{}), args[2].(string), variadicArgs...)
	})
	return _c
}

func (_c *Builder_SelectContext_Call) Return(_a0 error) *Builder_SelectContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Builder_SelectContext_Call) RunAndReturn(run func(context.Context, interface{}, string, ...interface{}) error) *Builder_SelectContext_Call {
	_c.Call.Return(run)
	return _c
}

// NewBuilder creates a new instance of Builder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBuilder(t interface {
	mock.TestingT
	Cleanup(func())
}) *Builder {
	mock := &Builder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
