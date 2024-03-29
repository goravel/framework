// Code generated by mockery. DO NOT EDIT.

package gorm

import (
	context "context"

	config "github.com/goravel/framework/contracts/config"

	gorm "github.com/goravel/framework/contracts/database/gorm"

	mock "github.com/stretchr/testify/mock"

	orm "github.com/goravel/framework/contracts/database/orm"
)

// Initialize is an autogenerated mock type for the Initialize type
type Initialize struct {
	mock.Mock
}

type Initialize_Expecter struct {
	mock *mock.Mock
}

func (_m *Initialize) EXPECT() *Initialize_Expecter {
	return &Initialize_Expecter{mock: &_m.Mock}
}

// InitializeGorm provides a mock function with given fields: _a0, connection
func (_m *Initialize) InitializeGorm(_a0 config.Config, connection string) gorm.Gorm {
	ret := _m.Called(_a0, connection)

	if len(ret) == 0 {
		panic("no return value specified for InitializeGorm")
	}

	var r0 gorm.Gorm
	if rf, ok := ret.Get(0).(func(config.Config, string) gorm.Gorm); ok {
		r0 = rf(_a0, connection)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(gorm.Gorm)
		}
	}

	return r0
}

// Initialize_InitializeGorm_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InitializeGorm'
type Initialize_InitializeGorm_Call struct {
	*mock.Call
}

// InitializeGorm is a helper method to define mock.On call
//   - _a0 config.Config
//   - connection string
func (_e *Initialize_Expecter) InitializeGorm(_a0 interface{}, connection interface{}) *Initialize_InitializeGorm_Call {
	return &Initialize_InitializeGorm_Call{Call: _e.mock.On("InitializeGorm", _a0, connection)}
}

func (_c *Initialize_InitializeGorm_Call) Run(run func(_a0 config.Config, connection string)) *Initialize_InitializeGorm_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(config.Config), args[1].(string))
	})
	return _c
}

func (_c *Initialize_InitializeGorm_Call) Return(_a0 gorm.Gorm) *Initialize_InitializeGorm_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Initialize_InitializeGorm_Call) RunAndReturn(run func(config.Config, string) gorm.Gorm) *Initialize_InitializeGorm_Call {
	_c.Call.Return(run)
	return _c
}

// InitializeQuery provides a mock function with given fields: ctx, _a1, connection
func (_m *Initialize) InitializeQuery(ctx context.Context, _a1 config.Config, connection string) (orm.Query, error) {
	ret := _m.Called(ctx, _a1, connection)

	if len(ret) == 0 {
		panic("no return value specified for InitializeQuery")
	}

	var r0 orm.Query
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, config.Config, string) (orm.Query, error)); ok {
		return rf(ctx, _a1, connection)
	}
	if rf, ok := ret.Get(0).(func(context.Context, config.Config, string) orm.Query); ok {
		r0 = rf(ctx, _a1, connection)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(orm.Query)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, config.Config, string) error); ok {
		r1 = rf(ctx, _a1, connection)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Initialize_InitializeQuery_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InitializeQuery'
type Initialize_InitializeQuery_Call struct {
	*mock.Call
}

// InitializeQuery is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 config.Config
//   - connection string
func (_e *Initialize_Expecter) InitializeQuery(ctx interface{}, _a1 interface{}, connection interface{}) *Initialize_InitializeQuery_Call {
	return &Initialize_InitializeQuery_Call{Call: _e.mock.On("InitializeQuery", ctx, _a1, connection)}
}

func (_c *Initialize_InitializeQuery_Call) Run(run func(ctx context.Context, _a1 config.Config, connection string)) *Initialize_InitializeQuery_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(config.Config), args[2].(string))
	})
	return _c
}

func (_c *Initialize_InitializeQuery_Call) Return(_a0 orm.Query, _a1 error) *Initialize_InitializeQuery_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Initialize_InitializeQuery_Call) RunAndReturn(run func(context.Context, config.Config, string) (orm.Query, error)) *Initialize_InitializeQuery_Call {
	_c.Call.Return(run)
	return _c
}

// NewInitialize creates a new instance of Initialize. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewInitialize(t interface {
	mock.TestingT
	Cleanup(func())
}) *Initialize {
	mock := &Initialize{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
