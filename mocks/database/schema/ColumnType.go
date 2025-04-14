// Code generated by mockery. DO NOT EDIT.

package schema

import mock "github.com/stretchr/testify/mock"

// ColumnType is an autogenerated mock type for the ColumnType type
type ColumnType[K comparable, V interface{}] struct {
	mock.Mock
}

type ColumnType_Expecter[K comparable, V interface{}] struct {
	mock *mock.Mock
}

func (_m *ColumnType[K, V]) EXPECT() *ColumnType_Expecter[K, V] {
	return &ColumnType_Expecter[K, V]{mock: &_m.Mock}
}

// Key provides a mock function with no fields
func (_m *ColumnType[K, V]) Key() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Key")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ColumnType_Key_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Key'
type ColumnType_Key_Call[K comparable, V interface{}] struct {
	*mock.Call
}

// Key is a helper method to define mock.On call
func (_e *ColumnType_Expecter[K, V]) Key() *ColumnType_Key_Call[K, V] {
	return &ColumnType_Key_Call[K, V]{Call: _e.mock.On("Key")}
}

func (_c *ColumnType_Key_Call[K, V]) Run(run func()) *ColumnType_Key_Call[K, V] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ColumnType_Key_Call[K, V]) Return(_a0 string) *ColumnType_Key_Call[K, V] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ColumnType_Key_Call[K, V]) RunAndReturn(run func() string) *ColumnType_Key_Call[K, V] {
	_c.Call.Return(run)
	return _c
}

// MarshalJSON provides a mock function with no fields
func (_m *ColumnType[K, V]) MarshalJSON() ([]byte, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for MarshalJSON")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]byte, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ColumnType_MarshalJSON_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MarshalJSON'
type ColumnType_MarshalJSON_Call[K comparable, V interface{}] struct {
	*mock.Call
}

// MarshalJSON is a helper method to define mock.On call
func (_e *ColumnType_Expecter[K, V]) MarshalJSON() *ColumnType_MarshalJSON_Call[K, V] {
	return &ColumnType_MarshalJSON_Call[K, V]{Call: _e.mock.On("MarshalJSON")}
}

func (_c *ColumnType_MarshalJSON_Call[K, V]) Run(run func()) *ColumnType_MarshalJSON_Call[K, V] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ColumnType_MarshalJSON_Call[K, V]) Return(_a0 []byte, _a1 error) *ColumnType_MarshalJSON_Call[K, V] {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ColumnType_MarshalJSON_Call[K, V]) RunAndReturn(run func() ([]byte, error)) *ColumnType_MarshalJSON_Call[K, V] {
	_c.Call.Return(run)
	return _c
}

// String provides a mock function with no fields
func (_m *ColumnType[K, V]) String() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for String")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ColumnType_String_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'String'
type ColumnType_String_Call[K comparable, V interface{}] struct {
	*mock.Call
}

// String is a helper method to define mock.On call
func (_e *ColumnType_Expecter[K, V]) String() *ColumnType_String_Call[K, V] {
	return &ColumnType_String_Call[K, V]{Call: _e.mock.On("String")}
}

func (_c *ColumnType_String_Call[K, V]) Run(run func()) *ColumnType_String_Call[K, V] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ColumnType_String_Call[K, V]) Return(_a0 string) *ColumnType_String_Call[K, V] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ColumnType_String_Call[K, V]) RunAndReturn(run func() string) *ColumnType_String_Call[K, V] {
	_c.Call.Return(run)
	return _c
}

// Value provides a mock function with no fields
func (_m *ColumnType[K, V]) Value() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Value")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ColumnType_Value_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Value'
type ColumnType_Value_Call[K comparable, V interface{}] struct {
	*mock.Call
}

// Value is a helper method to define mock.On call
func (_e *ColumnType_Expecter[K, V]) Value() *ColumnType_Value_Call[K, V] {
	return &ColumnType_Value_Call[K, V]{Call: _e.mock.On("Value")}
}

func (_c *ColumnType_Value_Call[K, V]) Run(run func()) *ColumnType_Value_Call[K, V] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ColumnType_Value_Call[K, V]) Return(_a0 string) *ColumnType_Value_Call[K, V] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ColumnType_Value_Call[K, V]) RunAndReturn(run func() string) *ColumnType_Value_Call[K, V] {
	_c.Call.Return(run)
	return _c
}

// NewColumnType creates a new instance of ColumnType. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewColumnType[K comparable, V interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *ColumnType[K, V] {
	mock := &ColumnType[K, V]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
