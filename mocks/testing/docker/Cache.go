// Code generated by mockery. DO NOT EDIT.

package docker

import (
	docker "github.com/goravel/framework/contracts/testing/docker"
	mock "github.com/stretchr/testify/mock"
)

// Cache is an autogenerated mock type for the Cache type
type Cache struct {
	mock.Mock
}

type Cache_Expecter struct {
	mock *mock.Mock
}

func (_m *Cache) EXPECT() *Cache_Expecter {
	return &Cache_Expecter{mock: &_m.Mock}
}

// Build provides a mock function with no fields
func (_m *Cache) Build() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Build")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Cache_Build_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Build'
type Cache_Build_Call struct {
	*mock.Call
}

// Build is a helper method to define mock.On call
func (_e *Cache_Expecter) Build() *Cache_Build_Call {
	return &Cache_Build_Call{Call: _e.mock.On("Build")}
}

func (_c *Cache_Build_Call) Run(run func()) *Cache_Build_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cache_Build_Call) Return(_a0 error) *Cache_Build_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cache_Build_Call) RunAndReturn(run func() error) *Cache_Build_Call {
	_c.Call.Return(run)
	return _c
}

// Config provides a mock function with no fields
func (_m *Cache) Config() docker.CacheConfig {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Config")
	}

	var r0 docker.CacheConfig
	if rf, ok := ret.Get(0).(func() docker.CacheConfig); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(docker.CacheConfig)
	}

	return r0
}

// Cache_Config_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Config'
type Cache_Config_Call struct {
	*mock.Call
}

// Config is a helper method to define mock.On call
func (_e *Cache_Expecter) Config() *Cache_Config_Call {
	return &Cache_Config_Call{Call: _e.mock.On("Config")}
}

func (_c *Cache_Config_Call) Run(run func()) *Cache_Config_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cache_Config_Call) Return(_a0 docker.CacheConfig) *Cache_Config_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cache_Config_Call) RunAndReturn(run func() docker.CacheConfig) *Cache_Config_Call {
	_c.Call.Return(run)
	return _c
}

// Fresh provides a mock function with no fields
func (_m *Cache) Fresh() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Fresh")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Cache_Fresh_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Fresh'
type Cache_Fresh_Call struct {
	*mock.Call
}

// Fresh is a helper method to define mock.On call
func (_e *Cache_Expecter) Fresh() *Cache_Fresh_Call {
	return &Cache_Fresh_Call{Call: _e.mock.On("Fresh")}
}

func (_c *Cache_Fresh_Call) Run(run func()) *Cache_Fresh_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cache_Fresh_Call) Return(_a0 error) *Cache_Fresh_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cache_Fresh_Call) RunAndReturn(run func() error) *Cache_Fresh_Call {
	_c.Call.Return(run)
	return _c
}

// Image provides a mock function with given fields: image
func (_m *Cache) Image(image docker.Image) {
	_m.Called(image)
}

// Cache_Image_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Image'
type Cache_Image_Call struct {
	*mock.Call
}

// Image is a helper method to define mock.On call
//   - image docker.Image
func (_e *Cache_Expecter) Image(image interface{}) *Cache_Image_Call {
	return &Cache_Image_Call{Call: _e.mock.On("Image", image)}
}

func (_c *Cache_Image_Call) Run(run func(image docker.Image)) *Cache_Image_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(docker.Image))
	})
	return _c
}

func (_c *Cache_Image_Call) Return() *Cache_Image_Call {
	_c.Call.Return()
	return _c
}

func (_c *Cache_Image_Call) RunAndReturn(run func(docker.Image)) *Cache_Image_Call {
	_c.Run(run)
	return _c
}

// Ready provides a mock function with no fields
func (_m *Cache) Ready() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Ready")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Cache_Ready_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Ready'
type Cache_Ready_Call struct {
	*mock.Call
}

// Ready is a helper method to define mock.On call
func (_e *Cache_Expecter) Ready() *Cache_Ready_Call {
	return &Cache_Ready_Call{Call: _e.mock.On("Ready")}
}

func (_c *Cache_Ready_Call) Run(run func()) *Cache_Ready_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cache_Ready_Call) Return(_a0 error) *Cache_Ready_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cache_Ready_Call) RunAndReturn(run func() error) *Cache_Ready_Call {
	_c.Call.Return(run)
	return _c
}

// Reuse provides a mock function with given fields: containerID, port
func (_m *Cache) Reuse(containerID string, port int) error {
	ret := _m.Called(containerID, port)

	if len(ret) == 0 {
		panic("no return value specified for Reuse")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int) error); ok {
		r0 = rf(containerID, port)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Cache_Reuse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reuse'
type Cache_Reuse_Call struct {
	*mock.Call
}

// Reuse is a helper method to define mock.On call
//   - containerID string
//   - port int
func (_e *Cache_Expecter) Reuse(containerID interface{}, port interface{}) *Cache_Reuse_Call {
	return &Cache_Reuse_Call{Call: _e.mock.On("Reuse", containerID, port)}
}

func (_c *Cache_Reuse_Call) Run(run func(containerID string, port int)) *Cache_Reuse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(int))
	})
	return _c
}

func (_c *Cache_Reuse_Call) Return(_a0 error) *Cache_Reuse_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cache_Reuse_Call) RunAndReturn(run func(string, int) error) *Cache_Reuse_Call {
	_c.Call.Return(run)
	return _c
}

// Shutdown provides a mock function with no fields
func (_m *Cache) Shutdown() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Shutdown")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Cache_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type Cache_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
func (_e *Cache_Expecter) Shutdown() *Cache_Shutdown_Call {
	return &Cache_Shutdown_Call{Call: _e.mock.On("Shutdown")}
}

func (_c *Cache_Shutdown_Call) Run(run func()) *Cache_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cache_Shutdown_Call) Return(_a0 error) *Cache_Shutdown_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cache_Shutdown_Call) RunAndReturn(run func() error) *Cache_Shutdown_Call {
	_c.Call.Return(run)
	return _c
}

// NewCache creates a new instance of Cache. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCache(t interface {
	mock.TestingT
	Cleanup(func())
}) *Cache {
	mock := &Cache{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
