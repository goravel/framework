// Code generated by mockery. DO NOT EDIT.

package gorm

import (
	mock "github.com/stretchr/testify/mock"
	gorm "gorm.io/gorm"
)

// Gorm is an autogenerated mock type for the Gorm type
type Gorm struct {
	mock.Mock
}

type Gorm_Expecter struct {
	mock *mock.Mock
}

func (_m *Gorm) EXPECT() *Gorm_Expecter {
	return &Gorm_Expecter{mock: &_m.Mock}
}

// Make provides a mock function with given fields:
func (_m *Gorm) Make() (*gorm.DB, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Make")
	}

	var r0 *gorm.DB
	var r1 error
	if rf, ok := ret.Get(0).(func() (*gorm.DB, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *gorm.DB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gorm.DB)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Gorm_Make_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Make'
type Gorm_Make_Call struct {
	*mock.Call
}

// Make is a helper method to define mock.On call
func (_e *Gorm_Expecter) Make() *Gorm_Make_Call {
	return &Gorm_Make_Call{Call: _e.mock.On("Make")}
}

func (_c *Gorm_Make_Call) Run(run func()) *Gorm_Make_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Gorm_Make_Call) Return(_a0 *gorm.DB, _a1 error) *Gorm_Make_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Gorm_Make_Call) RunAndReturn(run func() (*gorm.DB, error)) *Gorm_Make_Call {
	_c.Call.Return(run)
	return _c
}

// NewGorm creates a new instance of Gorm. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGorm(t interface {
	mock.TestingT
	Cleanup(func())
}) *Gorm {
	mock := &Gorm{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
