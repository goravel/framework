// Code generated by mockery. DO NOT EDIT.

package config

import (
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// Config is an autogenerated mock type for the Config type
type Config struct {
	mock.Mock
}

type Config_Expecter struct {
	mock *mock.Mock
}

func (_m *Config) EXPECT() *Config_Expecter {
	return &Config_Expecter{mock: &_m.Mock}
}

// Add provides a mock function with given fields: name, configuration
func (_m *Config) Add(name string, configuration interface{}) {
	_m.Called(name, configuration)
}

// Config_Add_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Add'
type Config_Add_Call struct {
	*mock.Call
}

// Add is a helper method to define mock.On call
//   - name string
//   - configuration interface{}
func (_e *Config_Expecter) Add(name interface{}, configuration interface{}) *Config_Add_Call {
	return &Config_Add_Call{Call: _e.mock.On("Add", name, configuration)}
}

func (_c *Config_Add_Call) Run(run func(name string, configuration interface{})) *Config_Add_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(interface{}))
	})
	return _c
}

func (_c *Config_Add_Call) Return() *Config_Add_Call {
	_c.Call.Return()
	return _c
}

func (_c *Config_Add_Call) RunAndReturn(run func(string, interface{})) *Config_Add_Call {
	_c.Run(run)
	return _c
}

// Env provides a mock function with given fields: envName, defaultValue
func (_m *Config) Env(envName string, defaultValue ...interface{}) interface{} {
	var _ca []interface{}
	_ca = append(_ca, envName)
	_ca = append(_ca, defaultValue...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Env")
	}

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string, ...interface{}) interface{}); ok {
		r0 = rf(envName, defaultValue...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// Config_Env_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Env'
type Config_Env_Call struct {
	*mock.Call
}

// Env is a helper method to define mock.On call
//   - envName string
//   - defaultValue ...interface{}
func (_e *Config_Expecter) Env(envName interface{}, defaultValue ...interface{}) *Config_Env_Call {
	return &Config_Env_Call{Call: _e.mock.On("Env",
		append([]interface{}{envName}, defaultValue...)...)}
}

func (_c *Config_Env_Call) Run(run func(envName string, defaultValue ...interface{})) *Config_Env_Call {
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

func (_c *Config_Env_Call) Return(_a0 interface{}) *Config_Env_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Config_Env_Call) RunAndReturn(run func(string, ...interface{}) interface{}) *Config_Env_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: path, defaultValue
func (_m *Config) Get(path string, defaultValue ...interface{}) interface{} {
	var _ca []interface{}
	_ca = append(_ca, path)
	_ca = append(_ca, defaultValue...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string, ...interface{}) interface{}); ok {
		r0 = rf(path, defaultValue...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// Config_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Config_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - path string
//   - defaultValue ...interface{}
func (_e *Config_Expecter) Get(path interface{}, defaultValue ...interface{}) *Config_Get_Call {
	return &Config_Get_Call{Call: _e.mock.On("Get",
		append([]interface{}{path}, defaultValue...)...)}
}

func (_c *Config_Get_Call) Run(run func(path string, defaultValue ...interface{})) *Config_Get_Call {
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

func (_c *Config_Get_Call) Return(_a0 interface{}) *Config_Get_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Config_Get_Call) RunAndReturn(run func(string, ...interface{}) interface{}) *Config_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetBool provides a mock function with given fields: path, defaultValue
func (_m *Config) GetBool(path string, defaultValue ...bool) bool {
	_va := make([]interface{}, len(defaultValue))
	for _i := range defaultValue {
		_va[_i] = defaultValue[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, path)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetBool")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, ...bool) bool); ok {
		r0 = rf(path, defaultValue...)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Config_GetBool_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBool'
type Config_GetBool_Call struct {
	*mock.Call
}

// GetBool is a helper method to define mock.On call
//   - path string
//   - defaultValue ...bool
func (_e *Config_Expecter) GetBool(path interface{}, defaultValue ...interface{}) *Config_GetBool_Call {
	return &Config_GetBool_Call{Call: _e.mock.On("GetBool",
		append([]interface{}{path}, defaultValue...)...)}
}

func (_c *Config_GetBool_Call) Run(run func(path string, defaultValue ...bool)) *Config_GetBool_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]bool, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(bool)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Config_GetBool_Call) Return(_a0 bool) *Config_GetBool_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Config_GetBool_Call) RunAndReturn(run func(string, ...bool) bool) *Config_GetBool_Call {
	_c.Call.Return(run)
	return _c
}

// GetDuration provides a mock function with given fields: path, defaultValue
func (_m *Config) GetDuration(path string, defaultValue ...time.Duration) time.Duration {
	_va := make([]interface{}, len(defaultValue))
	for _i := range defaultValue {
		_va[_i] = defaultValue[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, path)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetDuration")
	}

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func(string, ...time.Duration) time.Duration); ok {
		r0 = rf(path, defaultValue...)
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// Config_GetDuration_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDuration'
type Config_GetDuration_Call struct {
	*mock.Call
}

// GetDuration is a helper method to define mock.On call
//   - path string
//   - defaultValue ...time.Duration
func (_e *Config_Expecter) GetDuration(path interface{}, defaultValue ...interface{}) *Config_GetDuration_Call {
	return &Config_GetDuration_Call{Call: _e.mock.On("GetDuration",
		append([]interface{}{path}, defaultValue...)...)}
}

func (_c *Config_GetDuration_Call) Run(run func(path string, defaultValue ...time.Duration)) *Config_GetDuration_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]time.Duration, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(time.Duration)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Config_GetDuration_Call) Return(_a0 time.Duration) *Config_GetDuration_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Config_GetDuration_Call) RunAndReturn(run func(string, ...time.Duration) time.Duration) *Config_GetDuration_Call {
	_c.Call.Return(run)
	return _c
}

// GetInt provides a mock function with given fields: path, defaultValue
func (_m *Config) GetInt(path string, defaultValue ...int) int {
	_va := make([]interface{}, len(defaultValue))
	for _i := range defaultValue {
		_va[_i] = defaultValue[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, path)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetInt")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func(string, ...int) int); ok {
		r0 = rf(path, defaultValue...)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Config_GetInt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetInt'
type Config_GetInt_Call struct {
	*mock.Call
}

// GetInt is a helper method to define mock.On call
//   - path string
//   - defaultValue ...int
func (_e *Config_Expecter) GetInt(path interface{}, defaultValue ...interface{}) *Config_GetInt_Call {
	return &Config_GetInt_Call{Call: _e.mock.On("GetInt",
		append([]interface{}{path}, defaultValue...)...)}
}

func (_c *Config_GetInt_Call) Run(run func(path string, defaultValue ...int)) *Config_GetInt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]int, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(int)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Config_GetInt_Call) Return(_a0 int) *Config_GetInt_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Config_GetInt_Call) RunAndReturn(run func(string, ...int) int) *Config_GetInt_Call {
	_c.Call.Return(run)
	return _c
}

// GetString provides a mock function with given fields: path, defaultValue
func (_m *Config) GetString(path string, defaultValue ...string) string {
	_va := make([]interface{}, len(defaultValue))
	for _i := range defaultValue {
		_va[_i] = defaultValue[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, path)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetString")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string, ...string) string); ok {
		r0 = rf(path, defaultValue...)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Config_GetString_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetString'
type Config_GetString_Call struct {
	*mock.Call
}

// GetString is a helper method to define mock.On call
//   - path string
//   - defaultValue ...string
func (_e *Config_Expecter) GetString(path interface{}, defaultValue ...interface{}) *Config_GetString_Call {
	return &Config_GetString_Call{Call: _e.mock.On("GetString",
		append([]interface{}{path}, defaultValue...)...)}
}

func (_c *Config_GetString_Call) Run(run func(path string, defaultValue ...string)) *Config_GetString_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Config_GetString_Call) Return(_a0 string) *Config_GetString_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Config_GetString_Call) RunAndReturn(run func(string, ...string) string) *Config_GetString_Call {
	_c.Call.Return(run)
	return _c
}

// NewConfig creates a new instance of Config. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConfig(t interface {
	mock.TestingT
	Cleanup(func())
}) *Config {
	mock := &Config{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
