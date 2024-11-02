// Code generated by mockery. DO NOT EDIT.

package testing

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	testing "github.com/goravel/framework/contracts/testing"
)

// TestRequest is an autogenerated mock type for the TestRequest type
type TestRequest struct {
	mock.Mock
}

type TestRequest_Expecter struct {
	mock *mock.Mock
}

func (_m *TestRequest) EXPECT() *TestRequest_Expecter {
	return &TestRequest_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: uri, body
func (_m *TestRequest) Delete(uri string, body io.Reader) (testing.TestResponse, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 testing.TestResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (testing.TestResponse, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) testing.TestResponse); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(uri, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TestRequest_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type TestRequest_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - uri string
//   - body io.Reader
func (_e *TestRequest_Expecter) Delete(uri interface{}, body interface{}) *TestRequest_Delete_Call {
	return &TestRequest_Delete_Call{Call: _e.mock.On("Delete", uri, body)}
}

func (_c *TestRequest_Delete_Call) Run(run func(uri string, body io.Reader)) *TestRequest_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader))
	})
	return _c
}

func (_c *TestRequest_Delete_Call) Return(_a0 testing.TestResponse, _a1 error) *TestRequest_Delete_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TestRequest_Delete_Call) RunAndReturn(run func(string, io.Reader) (testing.TestResponse, error)) *TestRequest_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: uri
func (_m *TestRequest) Get(uri string) (testing.TestResponse, error) {
	ret := _m.Called(uri)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 testing.TestResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (testing.TestResponse, error)); ok {
		return rf(uri)
	}
	if rf, ok := ret.Get(0).(func(string) testing.TestResponse); ok {
		r0 = rf(uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uri)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TestRequest_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type TestRequest_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - uri string
func (_e *TestRequest_Expecter) Get(uri interface{}) *TestRequest_Get_Call {
	return &TestRequest_Get_Call{Call: _e.mock.On("Get", uri)}
}

func (_c *TestRequest_Get_Call) Run(run func(uri string)) *TestRequest_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *TestRequest_Get_Call) Return(_a0 testing.TestResponse, _a1 error) *TestRequest_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TestRequest_Get_Call) RunAndReturn(run func(string) (testing.TestResponse, error)) *TestRequest_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Head provides a mock function with given fields: uri, body
func (_m *TestRequest) Head(uri string, body io.Reader) (testing.TestResponse, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Head")
	}

	var r0 testing.TestResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (testing.TestResponse, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) testing.TestResponse); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(uri, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TestRequest_Head_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Head'
type TestRequest_Head_Call struct {
	*mock.Call
}

// Head is a helper method to define mock.On call
//   - uri string
//   - body io.Reader
func (_e *TestRequest_Expecter) Head(uri interface{}, body interface{}) *TestRequest_Head_Call {
	return &TestRequest_Head_Call{Call: _e.mock.On("Head", uri, body)}
}

func (_c *TestRequest_Head_Call) Run(run func(uri string, body io.Reader)) *TestRequest_Head_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader))
	})
	return _c
}

func (_c *TestRequest_Head_Call) Return(_a0 testing.TestResponse, _a1 error) *TestRequest_Head_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TestRequest_Head_Call) RunAndReturn(run func(string, io.Reader) (testing.TestResponse, error)) *TestRequest_Head_Call {
	_c.Call.Return(run)
	return _c
}

// Options provides a mock function with given fields: uri
func (_m *TestRequest) Options(uri string) (testing.TestResponse, error) {
	ret := _m.Called(uri)

	if len(ret) == 0 {
		panic("no return value specified for Options")
	}

	var r0 testing.TestResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (testing.TestResponse, error)); ok {
		return rf(uri)
	}
	if rf, ok := ret.Get(0).(func(string) testing.TestResponse); ok {
		r0 = rf(uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uri)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TestRequest_Options_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Options'
type TestRequest_Options_Call struct {
	*mock.Call
}

// Options is a helper method to define mock.On call
//   - uri string
func (_e *TestRequest_Expecter) Options(uri interface{}) *TestRequest_Options_Call {
	return &TestRequest_Options_Call{Call: _e.mock.On("Options", uri)}
}

func (_c *TestRequest_Options_Call) Run(run func(uri string)) *TestRequest_Options_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *TestRequest_Options_Call) Return(_a0 testing.TestResponse, _a1 error) *TestRequest_Options_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TestRequest_Options_Call) RunAndReturn(run func(string) (testing.TestResponse, error)) *TestRequest_Options_Call {
	_c.Call.Return(run)
	return _c
}

// Patch provides a mock function with given fields: uri, body
func (_m *TestRequest) Patch(uri string, body io.Reader) (testing.TestResponse, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Patch")
	}

	var r0 testing.TestResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (testing.TestResponse, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) testing.TestResponse); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(uri, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TestRequest_Patch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Patch'
type TestRequest_Patch_Call struct {
	*mock.Call
}

// Patch is a helper method to define mock.On call
//   - uri string
//   - body io.Reader
func (_e *TestRequest_Expecter) Patch(uri interface{}, body interface{}) *TestRequest_Patch_Call {
	return &TestRequest_Patch_Call{Call: _e.mock.On("Patch", uri, body)}
}

func (_c *TestRequest_Patch_Call) Run(run func(uri string, body io.Reader)) *TestRequest_Patch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader))
	})
	return _c
}

func (_c *TestRequest_Patch_Call) Return(_a0 testing.TestResponse, _a1 error) *TestRequest_Patch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TestRequest_Patch_Call) RunAndReturn(run func(string, io.Reader) (testing.TestResponse, error)) *TestRequest_Patch_Call {
	_c.Call.Return(run)
	return _c
}

// Post provides a mock function with given fields: uri, body
func (_m *TestRequest) Post(uri string, body io.Reader) (testing.TestResponse, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Post")
	}

	var r0 testing.TestResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (testing.TestResponse, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) testing.TestResponse); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(uri, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TestRequest_Post_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Post'
type TestRequest_Post_Call struct {
	*mock.Call
}

// Post is a helper method to define mock.On call
//   - uri string
//   - body io.Reader
func (_e *TestRequest_Expecter) Post(uri interface{}, body interface{}) *TestRequest_Post_Call {
	return &TestRequest_Post_Call{Call: _e.mock.On("Post", uri, body)}
}

func (_c *TestRequest_Post_Call) Run(run func(uri string, body io.Reader)) *TestRequest_Post_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader))
	})
	return _c
}

func (_c *TestRequest_Post_Call) Return(_a0 testing.TestResponse, _a1 error) *TestRequest_Post_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TestRequest_Post_Call) RunAndReturn(run func(string, io.Reader) (testing.TestResponse, error)) *TestRequest_Post_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: uri, body
func (_m *TestRequest) Put(uri string, body io.Reader) (testing.TestResponse, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 testing.TestResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (testing.TestResponse, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) testing.TestResponse); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(uri, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TestRequest_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type TestRequest_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - uri string
//   - body io.Reader
func (_e *TestRequest_Expecter) Put(uri interface{}, body interface{}) *TestRequest_Put_Call {
	return &TestRequest_Put_Call{Call: _e.mock.On("Put", uri, body)}
}

func (_c *TestRequest_Put_Call) Run(run func(uri string, body io.Reader)) *TestRequest_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader))
	})
	return _c
}

func (_c *TestRequest_Put_Call) Return(_a0 testing.TestResponse, _a1 error) *TestRequest_Put_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TestRequest_Put_Call) RunAndReturn(run func(string, io.Reader) (testing.TestResponse, error)) *TestRequest_Put_Call {
	_c.Call.Return(run)
	return _c
}

// WithContext provides a mock function with given fields: ctx
func (_m *TestRequest) WithContext(ctx context.Context) testing.TestRequest {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for WithContext")
	}

	var r0 testing.TestRequest
	if rf, ok := ret.Get(0).(func(context.Context) testing.TestRequest); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestRequest)
		}
	}

	return r0
}

// TestRequest_WithContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithContext'
type TestRequest_WithContext_Call struct {
	*mock.Call
}

// WithContext is a helper method to define mock.On call
//   - ctx context.Context
func (_e *TestRequest_Expecter) WithContext(ctx interface{}) *TestRequest_WithContext_Call {
	return &TestRequest_WithContext_Call{Call: _e.mock.On("WithContext", ctx)}
}

func (_c *TestRequest_WithContext_Call) Run(run func(ctx context.Context)) *TestRequest_WithContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *TestRequest_WithContext_Call) Return(_a0 testing.TestRequest) *TestRequest_WithContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TestRequest_WithContext_Call) RunAndReturn(run func(context.Context) testing.TestRequest) *TestRequest_WithContext_Call {
	_c.Call.Return(run)
	return _c
}

// WithCookie provides a mock function with given fields: key, value
func (_m *TestRequest) WithCookie(key string, value string) testing.TestRequest {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for WithCookie")
	}

	var r0 testing.TestRequest
	if rf, ok := ret.Get(0).(func(string, string) testing.TestRequest); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestRequest)
		}
	}

	return r0
}

// TestRequest_WithCookie_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithCookie'
type TestRequest_WithCookie_Call struct {
	*mock.Call
}

// WithCookie is a helper method to define mock.On call
//   - key string
//   - value string
func (_e *TestRequest_Expecter) WithCookie(key interface{}, value interface{}) *TestRequest_WithCookie_Call {
	return &TestRequest_WithCookie_Call{Call: _e.mock.On("WithCookie", key, value)}
}

func (_c *TestRequest_WithCookie_Call) Run(run func(key string, value string)) *TestRequest_WithCookie_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *TestRequest_WithCookie_Call) Return(_a0 testing.TestRequest) *TestRequest_WithCookie_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TestRequest_WithCookie_Call) RunAndReturn(run func(string, string) testing.TestRequest) *TestRequest_WithCookie_Call {
	_c.Call.Return(run)
	return _c
}

// WithCookies provides a mock function with given fields: cookies
func (_m *TestRequest) WithCookies(cookies map[string]string) testing.TestRequest {
	ret := _m.Called(cookies)

	if len(ret) == 0 {
		panic("no return value specified for WithCookies")
	}

	var r0 testing.TestRequest
	if rf, ok := ret.Get(0).(func(map[string]string) testing.TestRequest); ok {
		r0 = rf(cookies)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestRequest)
		}
	}

	return r0
}

// TestRequest_WithCookies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithCookies'
type TestRequest_WithCookies_Call struct {
	*mock.Call
}

// WithCookies is a helper method to define mock.On call
//   - cookies map[string]string
func (_e *TestRequest_Expecter) WithCookies(cookies interface{}) *TestRequest_WithCookies_Call {
	return &TestRequest_WithCookies_Call{Call: _e.mock.On("WithCookies", cookies)}
}

func (_c *TestRequest_WithCookies_Call) Run(run func(cookies map[string]string)) *TestRequest_WithCookies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]string))
	})
	return _c
}

func (_c *TestRequest_WithCookies_Call) Return(_a0 testing.TestRequest) *TestRequest_WithCookies_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TestRequest_WithCookies_Call) RunAndReturn(run func(map[string]string) testing.TestRequest) *TestRequest_WithCookies_Call {
	_c.Call.Return(run)
	return _c
}

// WithHeader provides a mock function with given fields: key, value
func (_m *TestRequest) WithHeader(key string, value string) testing.TestRequest {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for WithHeader")
	}

	var r0 testing.TestRequest
	if rf, ok := ret.Get(0).(func(string, string) testing.TestRequest); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestRequest)
		}
	}

	return r0
}

// TestRequest_WithHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithHeader'
type TestRequest_WithHeader_Call struct {
	*mock.Call
}

// WithHeader is a helper method to define mock.On call
//   - key string
//   - value string
func (_e *TestRequest_Expecter) WithHeader(key interface{}, value interface{}) *TestRequest_WithHeader_Call {
	return &TestRequest_WithHeader_Call{Call: _e.mock.On("WithHeader", key, value)}
}

func (_c *TestRequest_WithHeader_Call) Run(run func(key string, value string)) *TestRequest_WithHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *TestRequest_WithHeader_Call) Return(_a0 testing.TestRequest) *TestRequest_WithHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TestRequest_WithHeader_Call) RunAndReturn(run func(string, string) testing.TestRequest) *TestRequest_WithHeader_Call {
	_c.Call.Return(run)
	return _c
}

// WithHeaders provides a mock function with given fields: headers
func (_m *TestRequest) WithHeaders(headers map[string]string) testing.TestRequest {
	ret := _m.Called(headers)

	if len(ret) == 0 {
		panic("no return value specified for WithHeaders")
	}

	var r0 testing.TestRequest
	if rf, ok := ret.Get(0).(func(map[string]string) testing.TestRequest); ok {
		r0 = rf(headers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestRequest)
		}
	}

	return r0
}

// TestRequest_WithHeaders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithHeaders'
type TestRequest_WithHeaders_Call struct {
	*mock.Call
}

// WithHeaders is a helper method to define mock.On call
//   - headers map[string]string
func (_e *TestRequest_Expecter) WithHeaders(headers interface{}) *TestRequest_WithHeaders_Call {
	return &TestRequest_WithHeaders_Call{Call: _e.mock.On("WithHeaders", headers)}
}

func (_c *TestRequest_WithHeaders_Call) Run(run func(headers map[string]string)) *TestRequest_WithHeaders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]string))
	})
	return _c
}

func (_c *TestRequest_WithHeaders_Call) Return(_a0 testing.TestRequest) *TestRequest_WithHeaders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TestRequest_WithHeaders_Call) RunAndReturn(run func(map[string]string) testing.TestRequest) *TestRequest_WithHeaders_Call {
	_c.Call.Return(run)
	return _c
}

// WithoutHeader provides a mock function with given fields: key
func (_m *TestRequest) WithoutHeader(key string) testing.TestRequest {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for WithoutHeader")
	}

	var r0 testing.TestRequest
	if rf, ok := ret.Get(0).(func(string) testing.TestRequest); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(testing.TestRequest)
		}
	}

	return r0
}

// TestRequest_WithoutHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithoutHeader'
type TestRequest_WithoutHeader_Call struct {
	*mock.Call
}

// WithoutHeader is a helper method to define mock.On call
//   - key string
func (_e *TestRequest_Expecter) WithoutHeader(key interface{}) *TestRequest_WithoutHeader_Call {
	return &TestRequest_WithoutHeader_Call{Call: _e.mock.On("WithoutHeader", key)}
}

func (_c *TestRequest_WithoutHeader_Call) Run(run func(key string)) *TestRequest_WithoutHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *TestRequest_WithoutHeader_Call) Return(_a0 testing.TestRequest) *TestRequest_WithoutHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TestRequest_WithoutHeader_Call) RunAndReturn(run func(string) testing.TestRequest) *TestRequest_WithoutHeader_Call {
	_c.Call.Return(run)
	return _c
}

// NewTestRequest creates a new instance of TestRequest. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTestRequest(t interface {
	mock.TestingT
	Cleanup(func())
}) *TestRequest {
	mock := &TestRequest{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
