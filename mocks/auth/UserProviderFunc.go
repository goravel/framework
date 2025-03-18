// Code generated by mockery. DO NOT EDIT.

package auth

import (
	auth "github.com/goravel/framework/contracts/auth"
	mock "github.com/stretchr/testify/mock"
)

// UserProviderFunc is an autogenerated mock type for the UserProviderFunc type
type UserProviderFunc struct {
	mock.Mock
}

type UserProviderFunc_Expecter struct {
	mock *mock.Mock
}

func (_m *UserProviderFunc) EXPECT() *UserProviderFunc_Expecter {
	return &UserProviderFunc_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: _a0
func (_m *UserProviderFunc) Execute(_a0 auth.Auth) (auth.UserProvider, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 auth.UserProvider
	var r1 error
	if rf, ok := ret.Get(0).(func(auth.Auth) (auth.UserProvider, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(auth.Auth) auth.UserProvider); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(auth.UserProvider)
		}
	}

	if rf, ok := ret.Get(1).(func(auth.Auth) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserProviderFunc_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type UserProviderFunc_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - _a0 auth.Auth
func (_e *UserProviderFunc_Expecter) Execute(_a0 interface{}) *UserProviderFunc_Execute_Call {
	return &UserProviderFunc_Execute_Call{Call: _e.mock.On("Execute", _a0)}
}

func (_c *UserProviderFunc_Execute_Call) Run(run func(_a0 auth.Auth)) *UserProviderFunc_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(auth.Auth))
	})
	return _c
}

func (_c *UserProviderFunc_Execute_Call) Return(_a0 auth.UserProvider, _a1 error) *UserProviderFunc_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *UserProviderFunc_Execute_Call) RunAndReturn(run func(auth.Auth) (auth.UserProvider, error)) *UserProviderFunc_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewUserProviderFunc creates a new instance of UserProviderFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUserProviderFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserProviderFunc {
	mock := &UserProviderFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
