// Code generated by mockery. DO NOT EDIT.

package auth

import mock "github.com/stretchr/testify/mock"

// Guard is an autogenerated mock type for the Guard type
type Guard struct {
	mock.Mock
}

type Guard_Expecter struct {
	mock *mock.Mock
}

func (_m *Guard) EXPECT() *Guard_Expecter {
	return &Guard_Expecter{mock: &_m.Mock}
}

// Check provides a mock function with no fields
func (_m *Guard) Check() bool {
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

// Guard_Check_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Check'
type Guard_Check_Call struct {
	*mock.Call
}

// Check is a helper method to define mock.On call
func (_e *Guard_Expecter) Check() *Guard_Check_Call {
	return &Guard_Check_Call{Call: _e.mock.On("Check")}
}

func (_c *Guard_Check_Call) Run(run func()) *Guard_Check_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Guard_Check_Call) Return(_a0 bool) *Guard_Check_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Guard_Check_Call) RunAndReturn(run func() bool) *Guard_Check_Call {
	_c.Call.Return(run)
	return _c
}

// Guest provides a mock function with no fields
func (_m *Guard) Guest() bool {
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

// Guard_Guest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Guest'
type Guard_Guest_Call struct {
	*mock.Call
}

// Guest is a helper method to define mock.On call
func (_e *Guard_Expecter) Guest() *Guard_Guest_Call {
	return &Guard_Guest_Call{Call: _e.mock.On("Guest")}
}

func (_c *Guard_Guest_Call) Run(run func()) *Guard_Guest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Guard_Guest_Call) Return(_a0 bool) *Guard_Guest_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Guard_Guest_Call) RunAndReturn(run func() bool) *Guard_Guest_Call {
	_c.Call.Return(run)
	return _c
}

// ID provides a mock function with no fields
func (_m *Guard) ID() (string, error) {
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

// Guard_ID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ID'
type Guard_ID_Call struct {
	*mock.Call
}

// ID is a helper method to define mock.On call
func (_e *Guard_Expecter) ID() *Guard_ID_Call {
	return &Guard_ID_Call{Call: _e.mock.On("ID")}
}

func (_c *Guard_ID_Call) Run(run func()) *Guard_ID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Guard_ID_Call) Return(_a0 string, _a1 error) *Guard_ID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Guard_ID_Call) RunAndReturn(run func() (string, error)) *Guard_ID_Call {
	_c.Call.Return(run)
	return _c
}

// Login provides a mock function with given fields: user
func (_m *Guard) Login(user interface{}) error {
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

// Guard_Login_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Login'
type Guard_Login_Call struct {
	*mock.Call
}

// Login is a helper method to define mock.On call
//   - user interface{}
func (_e *Guard_Expecter) Login(user interface{}) *Guard_Login_Call {
	return &Guard_Login_Call{Call: _e.mock.On("Login", user)}
}

func (_c *Guard_Login_Call) Run(run func(user interface{})) *Guard_Login_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *Guard_Login_Call) Return(err error) *Guard_Login_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *Guard_Login_Call) RunAndReturn(run func(interface{}) error) *Guard_Login_Call {
	_c.Call.Return(run)
	return _c
}

// LoginUsingID provides a mock function with given fields: id
func (_m *Guard) LoginUsingID(id interface{}) error {
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

// Guard_LoginUsingID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LoginUsingID'
type Guard_LoginUsingID_Call struct {
	*mock.Call
}

// LoginUsingID is a helper method to define mock.On call
//   - id interface{}
func (_e *Guard_Expecter) LoginUsingID(id interface{}) *Guard_LoginUsingID_Call {
	return &Guard_LoginUsingID_Call{Call: _e.mock.On("LoginUsingID", id)}
}

func (_c *Guard_LoginUsingID_Call) Run(run func(id interface{})) *Guard_LoginUsingID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *Guard_LoginUsingID_Call) Return(err error) *Guard_LoginUsingID_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *Guard_LoginUsingID_Call) RunAndReturn(run func(interface{}) error) *Guard_LoginUsingID_Call {
	_c.Call.Return(run)
	return _c
}

// Logout provides a mock function with no fields
func (_m *Guard) Logout() error {
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

// Guard_Logout_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Logout'
type Guard_Logout_Call struct {
	*mock.Call
}

// Logout is a helper method to define mock.On call
func (_e *Guard_Expecter) Logout() *Guard_Logout_Call {
	return &Guard_Logout_Call{Call: _e.mock.On("Logout")}
}

func (_c *Guard_Logout_Call) Run(run func()) *Guard_Logout_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Guard_Logout_Call) Return(_a0 error) *Guard_Logout_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Guard_Logout_Call) RunAndReturn(run func() error) *Guard_Logout_Call {
	_c.Call.Return(run)
	return _c
}

// User provides a mock function with given fields: user
func (_m *Guard) User(user interface{}) error {
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

// Guard_User_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'User'
type Guard_User_Call struct {
	*mock.Call
}

// User is a helper method to define mock.On call
//   - user interface{}
func (_e *Guard_Expecter) User(user interface{}) *Guard_User_Call {
	return &Guard_User_Call{Call: _e.mock.On("User", user)}
}

func (_c *Guard_User_Call) Run(run func(user interface{})) *Guard_User_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *Guard_User_Call) Return(_a0 error) *Guard_User_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Guard_User_Call) RunAndReturn(run func(interface{}) error) *Guard_User_Call {
	_c.Call.Return(run)
	return _c
}

// NewGuard creates a new instance of Guard. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGuard(t interface {
	mock.TestingT
	Cleanup(func())
}) *Guard {
	mock := &Guard{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
