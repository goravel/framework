// Code generated by mockery. DO NOT EDIT.

package http

import (
	http "github.com/goravel/framework/contracts/testing/http"
	mock "github.com/stretchr/testify/mock"
)

// AssertableJSON is an autogenerated mock type for the AssertableJSON type
type AssertableJSON struct {
	mock.Mock
}

type AssertableJSON_Expecter struct {
	mock *mock.Mock
}

func (_m *AssertableJSON) EXPECT() *AssertableJSON_Expecter {
	return &AssertableJSON_Expecter{mock: &_m.Mock}
}

// Count provides a mock function with given fields: key, value
func (_m *AssertableJSON) Count(key string, value int) http.AssertableJSON {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for Count")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func(string, int) http.AssertableJSON); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_Count_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Count'
type AssertableJSON_Count_Call struct {
	*mock.Call
}

// Count is a helper method to define mock.On call
//   - key string
//   - value int
func (_e *AssertableJSON_Expecter) Count(key interface{}, value interface{}) *AssertableJSON_Count_Call {
	return &AssertableJSON_Count_Call{Call: _e.mock.On("Count", key, value)}
}

func (_c *AssertableJSON_Count_Call) Run(run func(key string, value int)) *AssertableJSON_Count_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(int))
	})
	return _c
}

func (_c *AssertableJSON_Count_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_Count_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_Count_Call) RunAndReturn(run func(string, int) http.AssertableJSON) *AssertableJSON_Count_Call {
	_c.Call.Return(run)
	return _c
}

// Each provides a mock function with given fields: key, callback
func (_m *AssertableJSON) Each(key string, callback func(http.AssertableJSON)) http.AssertableJSON {
	ret := _m.Called(key, callback)

	if len(ret) == 0 {
		panic("no return value specified for Each")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func(string, func(http.AssertableJSON)) http.AssertableJSON); ok {
		r0 = rf(key, callback)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_Each_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Each'
type AssertableJSON_Each_Call struct {
	*mock.Call
}

// Each is a helper method to define mock.On call
//   - key string
//   - callback func(http.AssertableJSON)
func (_e *AssertableJSON_Expecter) Each(key interface{}, callback interface{}) *AssertableJSON_Each_Call {
	return &AssertableJSON_Each_Call{Call: _e.mock.On("Each", key, callback)}
}

func (_c *AssertableJSON_Each_Call) Run(run func(key string, callback func(http.AssertableJSON))) *AssertableJSON_Each_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(func(http.AssertableJSON)))
	})
	return _c
}

func (_c *AssertableJSON_Each_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_Each_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_Each_Call) RunAndReturn(run func(string, func(http.AssertableJSON)) http.AssertableJSON) *AssertableJSON_Each_Call {
	_c.Call.Return(run)
	return _c
}

// First provides a mock function with given fields: key, callback
func (_m *AssertableJSON) First(key string, callback func(http.AssertableJSON)) http.AssertableJSON {
	ret := _m.Called(key, callback)

	if len(ret) == 0 {
		panic("no return value specified for First")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func(string, func(http.AssertableJSON)) http.AssertableJSON); ok {
		r0 = rf(key, callback)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_First_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'First'
type AssertableJSON_First_Call struct {
	*mock.Call
}

// First is a helper method to define mock.On call
//   - key string
//   - callback func(http.AssertableJSON)
func (_e *AssertableJSON_Expecter) First(key interface{}, callback interface{}) *AssertableJSON_First_Call {
	return &AssertableJSON_First_Call{Call: _e.mock.On("First", key, callback)}
}

func (_c *AssertableJSON_First_Call) Run(run func(key string, callback func(http.AssertableJSON))) *AssertableJSON_First_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(func(http.AssertableJSON)))
	})
	return _c
}

func (_c *AssertableJSON_First_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_First_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_First_Call) RunAndReturn(run func(string, func(http.AssertableJSON)) http.AssertableJSON) *AssertableJSON_First_Call {
	_c.Call.Return(run)
	return _c
}

// Has provides a mock function with given fields: key
func (_m *AssertableJSON) Has(key string) http.AssertableJSON {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Has")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func(string) http.AssertableJSON); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_Has_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Has'
type AssertableJSON_Has_Call struct {
	*mock.Call
}

// Has is a helper method to define mock.On call
//   - key string
func (_e *AssertableJSON_Expecter) Has(key interface{}) *AssertableJSON_Has_Call {
	return &AssertableJSON_Has_Call{Call: _e.mock.On("Has", key)}
}

func (_c *AssertableJSON_Has_Call) Run(run func(key string)) *AssertableJSON_Has_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *AssertableJSON_Has_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_Has_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_Has_Call) RunAndReturn(run func(string) http.AssertableJSON) *AssertableJSON_Has_Call {
	_c.Call.Return(run)
	return _c
}

// HasAll provides a mock function with given fields: keys
func (_m *AssertableJSON) HasAll(keys []string) http.AssertableJSON {
	ret := _m.Called(keys)

	if len(ret) == 0 {
		panic("no return value specified for HasAll")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func([]string) http.AssertableJSON); ok {
		r0 = rf(keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_HasAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasAll'
type AssertableJSON_HasAll_Call struct {
	*mock.Call
}

// HasAll is a helper method to define mock.On call
//   - keys []string
func (_e *AssertableJSON_Expecter) HasAll(keys interface{}) *AssertableJSON_HasAll_Call {
	return &AssertableJSON_HasAll_Call{Call: _e.mock.On("HasAll", keys)}
}

func (_c *AssertableJSON_HasAll_Call) Run(run func(keys []string)) *AssertableJSON_HasAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *AssertableJSON_HasAll_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_HasAll_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_HasAll_Call) RunAndReturn(run func([]string) http.AssertableJSON) *AssertableJSON_HasAll_Call {
	_c.Call.Return(run)
	return _c
}

// HasAny provides a mock function with given fields: keys
func (_m *AssertableJSON) HasAny(keys []string) http.AssertableJSON {
	ret := _m.Called(keys)

	if len(ret) == 0 {
		panic("no return value specified for HasAny")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func([]string) http.AssertableJSON); ok {
		r0 = rf(keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_HasAny_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasAny'
type AssertableJSON_HasAny_Call struct {
	*mock.Call
}

// HasAny is a helper method to define mock.On call
//   - keys []string
func (_e *AssertableJSON_Expecter) HasAny(keys interface{}) *AssertableJSON_HasAny_Call {
	return &AssertableJSON_HasAny_Call{Call: _e.mock.On("HasAny", keys)}
}

func (_c *AssertableJSON_HasAny_Call) Run(run func(keys []string)) *AssertableJSON_HasAny_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *AssertableJSON_HasAny_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_HasAny_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_HasAny_Call) RunAndReturn(run func([]string) http.AssertableJSON) *AssertableJSON_HasAny_Call {
	_c.Call.Return(run)
	return _c
}

// HasWithScope provides a mock function with given fields: key, length, callback
func (_m *AssertableJSON) HasWithScope(key string, length int, callback func(http.AssertableJSON)) http.AssertableJSON {
	ret := _m.Called(key, length, callback)

	if len(ret) == 0 {
		panic("no return value specified for HasWithScope")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func(string, int, func(http.AssertableJSON)) http.AssertableJSON); ok {
		r0 = rf(key, length, callback)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_HasWithScope_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasWithScope'
type AssertableJSON_HasWithScope_Call struct {
	*mock.Call
}

// HasWithScope is a helper method to define mock.On call
//   - key string
//   - length int
//   - callback func(http.AssertableJSON)
func (_e *AssertableJSON_Expecter) HasWithScope(key interface{}, length interface{}, callback interface{}) *AssertableJSON_HasWithScope_Call {
	return &AssertableJSON_HasWithScope_Call{Call: _e.mock.On("HasWithScope", key, length, callback)}
}

func (_c *AssertableJSON_HasWithScope_Call) Run(run func(key string, length int, callback func(http.AssertableJSON))) *AssertableJSON_HasWithScope_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(int), args[2].(func(http.AssertableJSON)))
	})
	return _c
}

func (_c *AssertableJSON_HasWithScope_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_HasWithScope_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_HasWithScope_Call) RunAndReturn(run func(string, int, func(http.AssertableJSON)) http.AssertableJSON) *AssertableJSON_HasWithScope_Call {
	_c.Call.Return(run)
	return _c
}

// Json provides a mock function with no fields
func (_m *AssertableJSON) Json() map[string]interface{} {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Json")
	}

	var r0 map[string]interface{}
	if rf, ok := ret.Get(0).(func() map[string]interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	return r0
}

// AssertableJSON_Json_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Json'
type AssertableJSON_Json_Call struct {
	*mock.Call
}

// Json is a helper method to define mock.On call
func (_e *AssertableJSON_Expecter) Json() *AssertableJSON_Json_Call {
	return &AssertableJSON_Json_Call{Call: _e.mock.On("Json")}
}

func (_c *AssertableJSON_Json_Call) Run(run func()) *AssertableJSON_Json_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AssertableJSON_Json_Call) Return(_a0 map[string]interface{}) *AssertableJSON_Json_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_Json_Call) RunAndReturn(run func() map[string]interface{}) *AssertableJSON_Json_Call {
	_c.Call.Return(run)
	return _c
}

// Missing provides a mock function with given fields: key
func (_m *AssertableJSON) Missing(key string) http.AssertableJSON {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Missing")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func(string) http.AssertableJSON); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_Missing_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Missing'
type AssertableJSON_Missing_Call struct {
	*mock.Call
}

// Missing is a helper method to define mock.On call
//   - key string
func (_e *AssertableJSON_Expecter) Missing(key interface{}) *AssertableJSON_Missing_Call {
	return &AssertableJSON_Missing_Call{Call: _e.mock.On("Missing", key)}
}

func (_c *AssertableJSON_Missing_Call) Run(run func(key string)) *AssertableJSON_Missing_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *AssertableJSON_Missing_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_Missing_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_Missing_Call) RunAndReturn(run func(string) http.AssertableJSON) *AssertableJSON_Missing_Call {
	_c.Call.Return(run)
	return _c
}

// MissingAll provides a mock function with given fields: keys
func (_m *AssertableJSON) MissingAll(keys []string) http.AssertableJSON {
	ret := _m.Called(keys)

	if len(ret) == 0 {
		panic("no return value specified for MissingAll")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func([]string) http.AssertableJSON); ok {
		r0 = rf(keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_MissingAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MissingAll'
type AssertableJSON_MissingAll_Call struct {
	*mock.Call
}

// MissingAll is a helper method to define mock.On call
//   - keys []string
func (_e *AssertableJSON_Expecter) MissingAll(keys interface{}) *AssertableJSON_MissingAll_Call {
	return &AssertableJSON_MissingAll_Call{Call: _e.mock.On("MissingAll", keys)}
}

func (_c *AssertableJSON_MissingAll_Call) Run(run func(keys []string)) *AssertableJSON_MissingAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *AssertableJSON_MissingAll_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_MissingAll_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_MissingAll_Call) RunAndReturn(run func([]string) http.AssertableJSON) *AssertableJSON_MissingAll_Call {
	_c.Call.Return(run)
	return _c
}

// Where provides a mock function with given fields: key, value
func (_m *AssertableJSON) Where(key string, value interface{}) http.AssertableJSON {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for Where")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func(string, interface{}) http.AssertableJSON); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_Where_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Where'
type AssertableJSON_Where_Call struct {
	*mock.Call
}

// Where is a helper method to define mock.On call
//   - key string
//   - value interface{}
func (_e *AssertableJSON_Expecter) Where(key interface{}, value interface{}) *AssertableJSON_Where_Call {
	return &AssertableJSON_Where_Call{Call: _e.mock.On("Where", key, value)}
}

func (_c *AssertableJSON_Where_Call) Run(run func(key string, value interface{})) *AssertableJSON_Where_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(interface{}))
	})
	return _c
}

func (_c *AssertableJSON_Where_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_Where_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_Where_Call) RunAndReturn(run func(string, interface{}) http.AssertableJSON) *AssertableJSON_Where_Call {
	_c.Call.Return(run)
	return _c
}

// WhereNot provides a mock function with given fields: key, value
func (_m *AssertableJSON) WhereNot(key string, value interface{}) http.AssertableJSON {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for WhereNot")
	}

	var r0 http.AssertableJSON
	if rf, ok := ret.Get(0).(func(string, interface{}) http.AssertableJSON); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.AssertableJSON)
		}
	}

	return r0
}

// AssertableJSON_WhereNot_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WhereNot'
type AssertableJSON_WhereNot_Call struct {
	*mock.Call
}

// WhereNot is a helper method to define mock.On call
//   - key string
//   - value interface{}
func (_e *AssertableJSON_Expecter) WhereNot(key interface{}, value interface{}) *AssertableJSON_WhereNot_Call {
	return &AssertableJSON_WhereNot_Call{Call: _e.mock.On("WhereNot", key, value)}
}

func (_c *AssertableJSON_WhereNot_Call) Run(run func(key string, value interface{})) *AssertableJSON_WhereNot_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(interface{}))
	})
	return _c
}

func (_c *AssertableJSON_WhereNot_Call) Return(_a0 http.AssertableJSON) *AssertableJSON_WhereNot_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssertableJSON_WhereNot_Call) RunAndReturn(run func(string, interface{}) http.AssertableJSON) *AssertableJSON_WhereNot_Call {
	_c.Call.Return(run)
	return _c
}

// NewAssertableJSON creates a new instance of AssertableJSON. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAssertableJSON(t interface {
	mock.TestingT
	Cleanup(func())
}) *AssertableJSON {
	mock := &AssertableJSON{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
