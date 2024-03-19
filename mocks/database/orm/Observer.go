// Code generated by mockery. DO NOT EDIT.

package orm

import (
	orm "github.com/goravel/framework/contracts/database/orm"
	mock "github.com/stretchr/testify/mock"
)

// Observer is an autogenerated mock type for the Observer type
type Observer struct {
	mock.Mock
}

type Observer_Expecter struct {
	mock *mock.Mock
}

func (_m *Observer) EXPECT() *Observer_Expecter {
	return &Observer_Expecter{mock: &_m.Mock}
}

// Created provides a mock function with given fields: _a0
func (_m *Observer) Created(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Created")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_Created_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Created'
type Observer_Created_Call struct {
	*mock.Call
}

// Created is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) Created(_a0 interface{}) *Observer_Created_Call {
	return &Observer_Created_Call{Call: _e.mock.On("Created", _a0)}
}

func (_c *Observer_Created_Call) Run(run func(_a0 orm.Event)) *Observer_Created_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_Created_Call) Return(_a0 error) *Observer_Created_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_Created_Call) RunAndReturn(run func(orm.Event) error) *Observer_Created_Call {
	_c.Call.Return(run)
	return _c
}

// Creating provides a mock function with given fields: _a0
func (_m *Observer) Creating(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Creating")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_Creating_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Creating'
type Observer_Creating_Call struct {
	*mock.Call
}

// Creating is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) Creating(_a0 interface{}) *Observer_Creating_Call {
	return &Observer_Creating_Call{Call: _e.mock.On("Creating", _a0)}
}

func (_c *Observer_Creating_Call) Run(run func(_a0 orm.Event)) *Observer_Creating_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_Creating_Call) Return(_a0 error) *Observer_Creating_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_Creating_Call) RunAndReturn(run func(orm.Event) error) *Observer_Creating_Call {
	_c.Call.Return(run)
	return _c
}

// Deleted provides a mock function with given fields: _a0
func (_m *Observer) Deleted(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Deleted")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_Deleted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Deleted'
type Observer_Deleted_Call struct {
	*mock.Call
}

// Deleted is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) Deleted(_a0 interface{}) *Observer_Deleted_Call {
	return &Observer_Deleted_Call{Call: _e.mock.On("Deleted", _a0)}
}

func (_c *Observer_Deleted_Call) Run(run func(_a0 orm.Event)) *Observer_Deleted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_Deleted_Call) Return(_a0 error) *Observer_Deleted_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_Deleted_Call) RunAndReturn(run func(orm.Event) error) *Observer_Deleted_Call {
	_c.Call.Return(run)
	return _c
}

// Deleting provides a mock function with given fields: _a0
func (_m *Observer) Deleting(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Deleting")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_Deleting_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Deleting'
type Observer_Deleting_Call struct {
	*mock.Call
}

// Deleting is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) Deleting(_a0 interface{}) *Observer_Deleting_Call {
	return &Observer_Deleting_Call{Call: _e.mock.On("Deleting", _a0)}
}

func (_c *Observer_Deleting_Call) Run(run func(_a0 orm.Event)) *Observer_Deleting_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_Deleting_Call) Return(_a0 error) *Observer_Deleting_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_Deleting_Call) RunAndReturn(run func(orm.Event) error) *Observer_Deleting_Call {
	_c.Call.Return(run)
	return _c
}

// ForceDeleted provides a mock function with given fields: _a0
func (_m *Observer) ForceDeleted(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for ForceDeleted")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_ForceDeleted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ForceDeleted'
type Observer_ForceDeleted_Call struct {
	*mock.Call
}

// ForceDeleted is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) ForceDeleted(_a0 interface{}) *Observer_ForceDeleted_Call {
	return &Observer_ForceDeleted_Call{Call: _e.mock.On("ForceDeleted", _a0)}
}

func (_c *Observer_ForceDeleted_Call) Run(run func(_a0 orm.Event)) *Observer_ForceDeleted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_ForceDeleted_Call) Return(_a0 error) *Observer_ForceDeleted_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_ForceDeleted_Call) RunAndReturn(run func(orm.Event) error) *Observer_ForceDeleted_Call {
	_c.Call.Return(run)
	return _c
}

// ForceDeleting provides a mock function with given fields: _a0
func (_m *Observer) ForceDeleting(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for ForceDeleting")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_ForceDeleting_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ForceDeleting'
type Observer_ForceDeleting_Call struct {
	*mock.Call
}

// ForceDeleting is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) ForceDeleting(_a0 interface{}) *Observer_ForceDeleting_Call {
	return &Observer_ForceDeleting_Call{Call: _e.mock.On("ForceDeleting", _a0)}
}

func (_c *Observer_ForceDeleting_Call) Run(run func(_a0 orm.Event)) *Observer_ForceDeleting_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_ForceDeleting_Call) Return(_a0 error) *Observer_ForceDeleting_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_ForceDeleting_Call) RunAndReturn(run func(orm.Event) error) *Observer_ForceDeleting_Call {
	_c.Call.Return(run)
	return _c
}

// Retrieved provides a mock function with given fields: _a0
func (_m *Observer) Retrieved(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Retrieved")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_Retrieved_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Retrieved'
type Observer_Retrieved_Call struct {
	*mock.Call
}

// Retrieved is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) Retrieved(_a0 interface{}) *Observer_Retrieved_Call {
	return &Observer_Retrieved_Call{Call: _e.mock.On("Retrieved", _a0)}
}

func (_c *Observer_Retrieved_Call) Run(run func(_a0 orm.Event)) *Observer_Retrieved_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_Retrieved_Call) Return(_a0 error) *Observer_Retrieved_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_Retrieved_Call) RunAndReturn(run func(orm.Event) error) *Observer_Retrieved_Call {
	_c.Call.Return(run)
	return _c
}

// Saved provides a mock function with given fields: _a0
func (_m *Observer) Saved(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Saved")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_Saved_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Saved'
type Observer_Saved_Call struct {
	*mock.Call
}

// Saved is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) Saved(_a0 interface{}) *Observer_Saved_Call {
	return &Observer_Saved_Call{Call: _e.mock.On("Saved", _a0)}
}

func (_c *Observer_Saved_Call) Run(run func(_a0 orm.Event)) *Observer_Saved_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_Saved_Call) Return(_a0 error) *Observer_Saved_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_Saved_Call) RunAndReturn(run func(orm.Event) error) *Observer_Saved_Call {
	_c.Call.Return(run)
	return _c
}

// Saving provides a mock function with given fields: _a0
func (_m *Observer) Saving(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Saving")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_Saving_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Saving'
type Observer_Saving_Call struct {
	*mock.Call
}

// Saving is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) Saving(_a0 interface{}) *Observer_Saving_Call {
	return &Observer_Saving_Call{Call: _e.mock.On("Saving", _a0)}
}

func (_c *Observer_Saving_Call) Run(run func(_a0 orm.Event)) *Observer_Saving_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_Saving_Call) Return(_a0 error) *Observer_Saving_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_Saving_Call) RunAndReturn(run func(orm.Event) error) *Observer_Saving_Call {
	_c.Call.Return(run)
	return _c
}

// Updated provides a mock function with given fields: _a0
func (_m *Observer) Updated(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Updated")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_Updated_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Updated'
type Observer_Updated_Call struct {
	*mock.Call
}

// Updated is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) Updated(_a0 interface{}) *Observer_Updated_Call {
	return &Observer_Updated_Call{Call: _e.mock.On("Updated", _a0)}
}

func (_c *Observer_Updated_Call) Run(run func(_a0 orm.Event)) *Observer_Updated_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_Updated_Call) Return(_a0 error) *Observer_Updated_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_Updated_Call) RunAndReturn(run func(orm.Event) error) *Observer_Updated_Call {
	_c.Call.Return(run)
	return _c
}

// Updating provides a mock function with given fields: _a0
func (_m *Observer) Updating(_a0 orm.Event) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Updating")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(orm.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Observer_Updating_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Updating'
type Observer_Updating_Call struct {
	*mock.Call
}

// Updating is a helper method to define mock.On call
//   - _a0 orm.Event
func (_e *Observer_Expecter) Updating(_a0 interface{}) *Observer_Updating_Call {
	return &Observer_Updating_Call{Call: _e.mock.On("Updating", _a0)}
}

func (_c *Observer_Updating_Call) Run(run func(_a0 orm.Event)) *Observer_Updating_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(orm.Event))
	})
	return _c
}

func (_c *Observer_Updating_Call) Return(_a0 error) *Observer_Updating_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Observer_Updating_Call) RunAndReturn(run func(orm.Event) error) *Observer_Updating_Call {
	_c.Call.Return(run)
	return _c
}

// NewObserver creates a new instance of Observer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewObserver(t interface {
	mock.TestingT
	Cleanup(func())
}) *Observer {
	mock := &Observer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
