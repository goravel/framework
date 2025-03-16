// Code generated by mockery. DO NOT EDIT.

package auth

import (
	auth "github.com/goravel/framework/contracts/auth"
	mock "github.com/stretchr/testify/mock"
)

// Auth is an autogenerated mock type for the Auth type
type Auth struct {
	mock.Mock
}

type Auth_Expecter struct {
	mock *mock.Mock
}

func (_m *Auth) EXPECT() *Auth_Expecter {
	return &Auth_Expecter{mock: &_m.Mock}
}

// Check provides a mock function with no fields
func (_m *Auth) Check() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Check")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Auth_Check_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Check'
type Auth_Check_Call struct {
	*mock.Call
}

// Check is a helper method to define mock.On call
func (_e *Auth_Expecter) Check() *Auth_Check_Call {
	return &Auth_Check_Call{Call: _e.mock.On("Check")}
}

func (_c *Auth_Check_Call) Run(run func()) *Auth_Check_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Auth_Check_Call) Return(_a0 bool) *Auth_Check_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Auth_Check_Call) RunAndReturn(run func() bool) *Auth_Check_Call {
	_c.Call.Return(run)
	return _c
}

// Extend provides a mock function with given fields: name, fn
func (_m *Auth) Extend(name string, fn auth.AuthGuardFunc) {
	_m.Called(name, fn)
}

// Auth_Extend_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Extend'
type Auth_Extend_Call struct {
	*mock.Call
}

// Extend is a helper method to define mock.On call
//   - name string
//   - fn auth.AuthGuardFunc
func (_e *Auth_Expecter) Extend(name interface{}, fn interface{}) *Auth_Extend_Call {
	return &Auth_Extend_Call{Call: _e.mock.On("Extend", name, fn)}
}

func (_c *Auth_Extend_Call) Run(run func(name string, fn auth.AuthGuardFunc)) *Auth_Extend_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(auth.AuthGuardFunc))
	})
	return _c
}

func (_c *Auth_Extend_Call) Return() *Auth_Extend_Call {
	_c.Call.Return()
	return _c
}

func (_c *Auth_Extend_Call) RunAndReturn(run func(string, auth.AuthGuardFunc)) *Auth_Extend_Call {
	_c.Run(run)
	return _c
}

// GetGuard provides a mock function with given fields: name
func (_m *Auth) GetGuard(name string) (auth.Guard, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for GetGuard")
	}

	var r0 auth.Guard
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (auth.Guard, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) auth.Guard); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(auth.Guard)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Auth_GetGuard_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGuard'
type Auth_GetGuard_Call struct {
	*mock.Call
}

// GetGuard is a helper method to define mock.On call
//   - name string
func (_e *Auth_Expecter) GetGuard(name interface{}) *Auth_GetGuard_Call {
	return &Auth_GetGuard_Call{Call: _e.mock.On("GetGuard", name)}
}

func (_c *Auth_GetGuard_Call) Run(run func(name string)) *Auth_GetGuard_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Auth_GetGuard_Call) Return(_a0 auth.Guard, _a1 error) *Auth_GetGuard_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Auth_GetGuard_Call) RunAndReturn(run func(string) (auth.Guard, error)) *Auth_GetGuard_Call {
	_c.Call.Return(run)
	return _c
}

// Guest provides a mock function with no fields
func (_m *Auth) Guest() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Guest")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Auth_Guest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Guest'
type Auth_Guest_Call struct {
	*mock.Call
}

// Guest is a helper method to define mock.On call
func (_e *Auth_Expecter) Guest() *Auth_Guest_Call {
	return &Auth_Guest_Call{Call: _e.mock.On("Guest")}
}

func (_c *Auth_Guest_Call) Run(run func()) *Auth_Guest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Auth_Guest_Call) Return(_a0 bool) *Auth_Guest_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Auth_Guest_Call) RunAndReturn(run func() bool) *Auth_Guest_Call {
	_c.Call.Return(run)
	return _c
}

// ID provides a mock function with no fields
func (_m *Auth) ID() (string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ID")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Auth_ID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ID'
type Auth_ID_Call struct {
	*mock.Call
}

// ID is a helper method to define mock.On call
func (_e *Auth_Expecter) ID() *Auth_ID_Call {
	return &Auth_ID_Call{Call: _e.mock.On("ID")}
}

func (_c *Auth_ID_Call) Run(run func()) *Auth_ID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Auth_ID_Call) Return(_a0 string, _a1 error) *Auth_ID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Auth_ID_Call) RunAndReturn(run func() (string, error)) *Auth_ID_Call {
	_c.Call.Return(run)
	return _c
}

// Login provides a mock function with given fields: user
func (_m *Auth) Login(user interface{}) error {
	ret := _m.Called(user)

	if len(ret) == 0 {
		panic("no return value specified for Login")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(user)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Auth_Login_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Login'
type Auth_Login_Call struct {
	*mock.Call
}

// Login is a helper method to define mock.On call
//   - user interface{}
func (_e *Auth_Expecter) Login(user interface{}) *Auth_Login_Call {
	return &Auth_Login_Call{Call: _e.mock.On("Login", user)}
}

func (_c *Auth_Login_Call) Run(run func(user interface{})) *Auth_Login_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *Auth_Login_Call) Return(err error) *Auth_Login_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *Auth_Login_Call) RunAndReturn(run func(interface{}) error) *Auth_Login_Call {
	_c.Call.Return(run)
	return _c
}

// LoginUsingID provides a mock function with given fields: id
func (_m *Auth) LoginUsingID(id interface{}) error {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for LoginUsingID")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Auth_LoginUsingID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LoginUsingID'
type Auth_LoginUsingID_Call struct {
	*mock.Call
}

// LoginUsingID is a helper method to define mock.On call
//   - id interface{}
func (_e *Auth_Expecter) LoginUsingID(id interface{}) *Auth_LoginUsingID_Call {
	return &Auth_LoginUsingID_Call{Call: _e.mock.On("LoginUsingID", id)}
}

func (_c *Auth_LoginUsingID_Call) Run(run func(id interface{})) *Auth_LoginUsingID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *Auth_LoginUsingID_Call) Return(err error) *Auth_LoginUsingID_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *Auth_LoginUsingID_Call) RunAndReturn(run func(interface{}) error) *Auth_LoginUsingID_Call {
	_c.Call.Return(run)
	return _c
}

// Logout provides a mock function with no fields
func (_m *Auth) Logout() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Logout")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Auth_Logout_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Logout'
type Auth_Logout_Call struct {
	*mock.Call
}

// Logout is a helper method to define mock.On call
func (_e *Auth_Expecter) Logout() *Auth_Logout_Call {
	return &Auth_Logout_Call{Call: _e.mock.On("Logout")}
}

func (_c *Auth_Logout_Call) Run(run func()) *Auth_Logout_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Auth_Logout_Call) Return(_a0 error) *Auth_Logout_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Auth_Logout_Call) RunAndReturn(run func() error) *Auth_Logout_Call {
	_c.Call.Return(run)
	return _c
}

// Provider provides a mock function with given fields: name, fn
func (_m *Auth) Provider(name string, fn auth.UserProviderFunc) {
	_m.Called(name, fn)
}

// Auth_Provider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Provider'
type Auth_Provider_Call struct {
	*mock.Call
}

// Provider is a helper method to define mock.On call
//   - name string
//   - fn auth.UserProviderFunc
func (_e *Auth_Expecter) Provider(name interface{}, fn interface{}) *Auth_Provider_Call {
	return &Auth_Provider_Call{Call: _e.mock.On("Provider", name, fn)}
}

func (_c *Auth_Provider_Call) Run(run func(name string, fn auth.UserProviderFunc)) *Auth_Provider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(auth.UserProviderFunc))
	})
	return _c
}

func (_c *Auth_Provider_Call) Return() *Auth_Provider_Call {
	_c.Call.Return()
	return _c
}

func (_c *Auth_Provider_Call) RunAndReturn(run func(string, auth.UserProviderFunc)) *Auth_Provider_Call {
	_c.Run(run)
	return _c
}

// User provides a mock function with given fields: user
func (_m *Auth) User(user interface{}) error {
	ret := _m.Called(user)

	if len(ret) == 0 {
		panic("no return value specified for User")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(user)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Auth_User_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'User'
type Auth_User_Call struct {
	*mock.Call
}

// User is a helper method to define mock.On call
//   - user interface{}
func (_e *Auth_Expecter) User(user interface{}) *Auth_User_Call {
	return &Auth_User_Call{Call: _e.mock.On("User", user)}
}

func (_c *Auth_User_Call) Run(run func(user interface{})) *Auth_User_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *Auth_User_Call) Return(_a0 error) *Auth_User_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Auth_User_Call) RunAndReturn(run func(interface{}) error) *Auth_User_Call {
	_c.Call.Return(run)
	return _c
}

// NewAuth creates a new instance of Auth. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAuth(t interface {
	mock.TestingT
	Cleanup(func())
}) *Auth {
	mock := &Auth{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
