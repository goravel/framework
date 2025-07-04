// Code generated by mockery. DO NOT EDIT.

package http

import (
	context "context"
	io "io"

	http "github.com/goravel/framework/contracts/testing/http"

	mock "github.com/stretchr/testify/mock"

	nethttp "net/http"
)

// Request is an autogenerated mock type for the Request type
type Request struct {
	mock.Mock
}

type Request_Expecter struct {
	mock *mock.Mock
}

func (_m *Request) EXPECT() *Request_Expecter {
	return &Request_Expecter{mock: &_m.Mock}
}

// Bind provides a mock function with given fields: value
func (_m *Request) Bind(value interface{}) http.Request {
	ret := _m.Called(value)

	if len(ret) == 0 {
		panic("no return value specified for Bind")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func(interface{}) http.Request); ok {
		r0 = rf(value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_Bind_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bind'
type Request_Bind_Call struct {
	*mock.Call
}

// Bind is a helper method to define mock.On call
//   - value interface{}
func (_e *Request_Expecter) Bind(value interface{}) *Request_Bind_Call {
	return &Request_Bind_Call{Call: _e.mock.On("Bind", value)}
}

func (_c *Request_Bind_Call) Run(run func(value interface{})) *Request_Bind_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *Request_Bind_Call) Return(_a0 http.Request) *Request_Bind_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_Bind_Call) RunAndReturn(run func(interface{}) http.Request) *Request_Bind_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: uri, body
func (_m *Request) Delete(uri string, body io.Reader) (http.Response, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (http.Response, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) http.Response); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(uri, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Request_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type Request_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - uri string
//   - body io.Reader
func (_e *Request_Expecter) Delete(uri interface{}, body interface{}) *Request_Delete_Call {
	return &Request_Delete_Call{Call: _e.mock.On("Delete", uri, body)}
}

func (_c *Request_Delete_Call) Run(run func(uri string, body io.Reader)) *Request_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader))
	})
	return _c
}

func (_c *Request_Delete_Call) Return(_a0 http.Response, _a1 error) *Request_Delete_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Delete_Call) RunAndReturn(run func(string, io.Reader) (http.Response, error)) *Request_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// FlushHeaders provides a mock function with no fields
func (_m *Request) FlushHeaders() http.Request {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for FlushHeaders")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func() http.Request); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_FlushHeaders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FlushHeaders'
type Request_FlushHeaders_Call struct {
	*mock.Call
}

// FlushHeaders is a helper method to define mock.On call
func (_e *Request_Expecter) FlushHeaders() *Request_FlushHeaders_Call {
	return &Request_FlushHeaders_Call{Call: _e.mock.On("FlushHeaders")}
}

func (_c *Request_FlushHeaders_Call) Run(run func()) *Request_FlushHeaders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Request_FlushHeaders_Call) Return(_a0 http.Request) *Request_FlushHeaders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_FlushHeaders_Call) RunAndReturn(run func() http.Request) *Request_FlushHeaders_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: uri
func (_m *Request) Get(uri string) (http.Response, error) {
	ret := _m.Called(uri)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (http.Response, error)); ok {
		return rf(uri)
	}
	if rf, ok := ret.Get(0).(func(string) http.Response); ok {
		r0 = rf(uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uri)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Request_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Request_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - uri string
func (_e *Request_Expecter) Get(uri interface{}) *Request_Get_Call {
	return &Request_Get_Call{Call: _e.mock.On("Get", uri)}
}

func (_c *Request_Get_Call) Run(run func(uri string)) *Request_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Request_Get_Call) Return(_a0 http.Response, _a1 error) *Request_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Get_Call) RunAndReturn(run func(string) (http.Response, error)) *Request_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Head provides a mock function with given fields: uri
func (_m *Request) Head(uri string) (http.Response, error) {
	ret := _m.Called(uri)

	if len(ret) == 0 {
		panic("no return value specified for Head")
	}

	var r0 http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (http.Response, error)); ok {
		return rf(uri)
	}
	if rf, ok := ret.Get(0).(func(string) http.Response); ok {
		r0 = rf(uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uri)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Request_Head_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Head'
type Request_Head_Call struct {
	*mock.Call
}

// Head is a helper method to define mock.On call
//   - uri string
func (_e *Request_Expecter) Head(uri interface{}) *Request_Head_Call {
	return &Request_Head_Call{Call: _e.mock.On("Head", uri)}
}

func (_c *Request_Head_Call) Run(run func(uri string)) *Request_Head_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Request_Head_Call) Return(_a0 http.Response, _a1 error) *Request_Head_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Head_Call) RunAndReturn(run func(string) (http.Response, error)) *Request_Head_Call {
	_c.Call.Return(run)
	return _c
}

// Options provides a mock function with given fields: uri
func (_m *Request) Options(uri string) (http.Response, error) {
	ret := _m.Called(uri)

	if len(ret) == 0 {
		panic("no return value specified for Options")
	}

	var r0 http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (http.Response, error)); ok {
		return rf(uri)
	}
	if rf, ok := ret.Get(0).(func(string) http.Response); ok {
		r0 = rf(uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uri)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Request_Options_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Options'
type Request_Options_Call struct {
	*mock.Call
}

// Options is a helper method to define mock.On call
//   - uri string
func (_e *Request_Expecter) Options(uri interface{}) *Request_Options_Call {
	return &Request_Options_Call{Call: _e.mock.On("Options", uri)}
}

func (_c *Request_Options_Call) Run(run func(uri string)) *Request_Options_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Request_Options_Call) Return(_a0 http.Response, _a1 error) *Request_Options_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Options_Call) RunAndReturn(run func(string) (http.Response, error)) *Request_Options_Call {
	_c.Call.Return(run)
	return _c
}

// Patch provides a mock function with given fields: uri, body
func (_m *Request) Patch(uri string, body io.Reader) (http.Response, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Patch")
	}

	var r0 http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (http.Response, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) http.Response); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(uri, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Request_Patch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Patch'
type Request_Patch_Call struct {
	*mock.Call
}

// Patch is a helper method to define mock.On call
//   - uri string
//   - body io.Reader
func (_e *Request_Expecter) Patch(uri interface{}, body interface{}) *Request_Patch_Call {
	return &Request_Patch_Call{Call: _e.mock.On("Patch", uri, body)}
}

func (_c *Request_Patch_Call) Run(run func(uri string, body io.Reader)) *Request_Patch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader))
	})
	return _c
}

func (_c *Request_Patch_Call) Return(_a0 http.Response, _a1 error) *Request_Patch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Patch_Call) RunAndReturn(run func(string, io.Reader) (http.Response, error)) *Request_Patch_Call {
	_c.Call.Return(run)
	return _c
}

// Post provides a mock function with given fields: uri, body
func (_m *Request) Post(uri string, body io.Reader) (http.Response, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Post")
	}

	var r0 http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (http.Response, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) http.Response); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(uri, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Request_Post_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Post'
type Request_Post_Call struct {
	*mock.Call
}

// Post is a helper method to define mock.On call
//   - uri string
//   - body io.Reader
func (_e *Request_Expecter) Post(uri interface{}, body interface{}) *Request_Post_Call {
	return &Request_Post_Call{Call: _e.mock.On("Post", uri, body)}
}

func (_c *Request_Post_Call) Run(run func(uri string, body io.Reader)) *Request_Post_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader))
	})
	return _c
}

func (_c *Request_Post_Call) Return(_a0 http.Response, _a1 error) *Request_Post_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Post_Call) RunAndReturn(run func(string, io.Reader) (http.Response, error)) *Request_Post_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: uri, body
func (_m *Request) Put(uri string, body io.Reader) (http.Response, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (http.Response, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) http.Response); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(uri, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Request_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type Request_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - uri string
//   - body io.Reader
func (_e *Request_Expecter) Put(uri interface{}, body interface{}) *Request_Put_Call {
	return &Request_Put_Call{Call: _e.mock.On("Put", uri, body)}
}

func (_c *Request_Put_Call) Run(run func(uri string, body io.Reader)) *Request_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader))
	})
	return _c
}

func (_c *Request_Put_Call) Return(_a0 http.Response, _a1 error) *Request_Put_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Put_Call) RunAndReturn(run func(string, io.Reader) (http.Response, error)) *Request_Put_Call {
	_c.Call.Return(run)
	return _c
}

// WithBasicAuth provides a mock function with given fields: username, password
func (_m *Request) WithBasicAuth(username string, password string) http.Request {
	ret := _m.Called(username, password)

	if len(ret) == 0 {
		panic("no return value specified for WithBasicAuth")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func(string, string) http.Request); ok {
		r0 = rf(username, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithBasicAuth_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithBasicAuth'
type Request_WithBasicAuth_Call struct {
	*mock.Call
}

// WithBasicAuth is a helper method to define mock.On call
//   - username string
//   - password string
func (_e *Request_Expecter) WithBasicAuth(username interface{}, password interface{}) *Request_WithBasicAuth_Call {
	return &Request_WithBasicAuth_Call{Call: _e.mock.On("WithBasicAuth", username, password)}
}

func (_c *Request_WithBasicAuth_Call) Run(run func(username string, password string)) *Request_WithBasicAuth_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Request_WithBasicAuth_Call) Return(_a0 http.Request) *Request_WithBasicAuth_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithBasicAuth_Call) RunAndReturn(run func(string, string) http.Request) *Request_WithBasicAuth_Call {
	_c.Call.Return(run)
	return _c
}

// WithContext provides a mock function with given fields: ctx
func (_m *Request) WithContext(ctx context.Context) http.Request {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for WithContext")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func(context.Context) http.Request); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithContext'
type Request_WithContext_Call struct {
	*mock.Call
}

// WithContext is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Request_Expecter) WithContext(ctx interface{}) *Request_WithContext_Call {
	return &Request_WithContext_Call{Call: _e.mock.On("WithContext", ctx)}
}

func (_c *Request_WithContext_Call) Run(run func(ctx context.Context)) *Request_WithContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Request_WithContext_Call) Return(_a0 http.Request) *Request_WithContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithContext_Call) RunAndReturn(run func(context.Context) http.Request) *Request_WithContext_Call {
	_c.Call.Return(run)
	return _c
}

// WithCookie provides a mock function with given fields: cookie
func (_m *Request) WithCookie(cookie *nethttp.Cookie) http.Request {
	ret := _m.Called(cookie)

	if len(ret) == 0 {
		panic("no return value specified for WithCookie")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func(*nethttp.Cookie) http.Request); ok {
		r0 = rf(cookie)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithCookie_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithCookie'
type Request_WithCookie_Call struct {
	*mock.Call
}

// WithCookie is a helper method to define mock.On call
//   - cookie *nethttp.Cookie
func (_e *Request_Expecter) WithCookie(cookie interface{}) *Request_WithCookie_Call {
	return &Request_WithCookie_Call{Call: _e.mock.On("WithCookie", cookie)}
}

func (_c *Request_WithCookie_Call) Run(run func(cookie *nethttp.Cookie)) *Request_WithCookie_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*nethttp.Cookie))
	})
	return _c
}

func (_c *Request_WithCookie_Call) Return(_a0 http.Request) *Request_WithCookie_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithCookie_Call) RunAndReturn(run func(*nethttp.Cookie) http.Request) *Request_WithCookie_Call {
	_c.Call.Return(run)
	return _c
}

// WithCookies provides a mock function with given fields: cookies
func (_m *Request) WithCookies(cookies []*nethttp.Cookie) http.Request {
	ret := _m.Called(cookies)

	if len(ret) == 0 {
		panic("no return value specified for WithCookies")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func([]*nethttp.Cookie) http.Request); ok {
		r0 = rf(cookies)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithCookies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithCookies'
type Request_WithCookies_Call struct {
	*mock.Call
}

// WithCookies is a helper method to define mock.On call
//   - cookies []*nethttp.Cookie
func (_e *Request_Expecter) WithCookies(cookies interface{}) *Request_WithCookies_Call {
	return &Request_WithCookies_Call{Call: _e.mock.On("WithCookies", cookies)}
}

func (_c *Request_WithCookies_Call) Run(run func(cookies []*nethttp.Cookie)) *Request_WithCookies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]*nethttp.Cookie))
	})
	return _c
}

func (_c *Request_WithCookies_Call) Return(_a0 http.Request) *Request_WithCookies_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithCookies_Call) RunAndReturn(run func([]*nethttp.Cookie) http.Request) *Request_WithCookies_Call {
	_c.Call.Return(run)
	return _c
}

// WithHeader provides a mock function with given fields: key, value
func (_m *Request) WithHeader(key string, value string) http.Request {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for WithHeader")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func(string, string) http.Request); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithHeader'
type Request_WithHeader_Call struct {
	*mock.Call
}

// WithHeader is a helper method to define mock.On call
//   - key string
//   - value string
func (_e *Request_Expecter) WithHeader(key interface{}, value interface{}) *Request_WithHeader_Call {
	return &Request_WithHeader_Call{Call: _e.mock.On("WithHeader", key, value)}
}

func (_c *Request_WithHeader_Call) Run(run func(key string, value string)) *Request_WithHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Request_WithHeader_Call) Return(_a0 http.Request) *Request_WithHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithHeader_Call) RunAndReturn(run func(string, string) http.Request) *Request_WithHeader_Call {
	_c.Call.Return(run)
	return _c
}

// WithHeaders provides a mock function with given fields: headers
func (_m *Request) WithHeaders(headers map[string]string) http.Request {
	ret := _m.Called(headers)

	if len(ret) == 0 {
		panic("no return value specified for WithHeaders")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func(map[string]string) http.Request); ok {
		r0 = rf(headers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithHeaders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithHeaders'
type Request_WithHeaders_Call struct {
	*mock.Call
}

// WithHeaders is a helper method to define mock.On call
//   - headers map[string]string
func (_e *Request_Expecter) WithHeaders(headers interface{}) *Request_WithHeaders_Call {
	return &Request_WithHeaders_Call{Call: _e.mock.On("WithHeaders", headers)}
}

func (_c *Request_WithHeaders_Call) Run(run func(headers map[string]string)) *Request_WithHeaders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]string))
	})
	return _c
}

func (_c *Request_WithHeaders_Call) Return(_a0 http.Request) *Request_WithHeaders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithHeaders_Call) RunAndReturn(run func(map[string]string) http.Request) *Request_WithHeaders_Call {
	_c.Call.Return(run)
	return _c
}

// WithSession provides a mock function with given fields: attributes
func (_m *Request) WithSession(attributes map[string]interface{}) http.Request {
	ret := _m.Called(attributes)

	if len(ret) == 0 {
		panic("no return value specified for WithSession")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func(map[string]interface{}) http.Request); ok {
		r0 = rf(attributes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithSession_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithSession'
type Request_WithSession_Call struct {
	*mock.Call
}

// WithSession is a helper method to define mock.On call
//   - attributes map[string]interface{}
func (_e *Request_Expecter) WithSession(attributes interface{}) *Request_WithSession_Call {
	return &Request_WithSession_Call{Call: _e.mock.On("WithSession", attributes)}
}

func (_c *Request_WithSession_Call) Run(run func(attributes map[string]interface{})) *Request_WithSession_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]interface{}))
	})
	return _c
}

func (_c *Request_WithSession_Call) Return(_a0 http.Request) *Request_WithSession_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithSession_Call) RunAndReturn(run func(map[string]interface{}) http.Request) *Request_WithSession_Call {
	_c.Call.Return(run)
	return _c
}

// WithToken provides a mock function with given fields: token, ttype
func (_m *Request) WithToken(token string, ttype ...string) http.Request {
	_va := make([]interface{}, len(ttype))
	for _i := range ttype {
		_va[_i] = ttype[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, token)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for WithToken")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func(string, ...string) http.Request); ok {
		r0 = rf(token, ttype...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithToken'
type Request_WithToken_Call struct {
	*mock.Call
}

// WithToken is a helper method to define mock.On call
//   - token string
//   - ttype ...string
func (_e *Request_Expecter) WithToken(token interface{}, ttype ...interface{}) *Request_WithToken_Call {
	return &Request_WithToken_Call{Call: _e.mock.On("WithToken",
		append([]interface{}{token}, ttype...)...)}
}

func (_c *Request_WithToken_Call) Run(run func(token string, ttype ...string)) *Request_WithToken_Call {
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

func (_c *Request_WithToken_Call) Return(_a0 http.Request) *Request_WithToken_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithToken_Call) RunAndReturn(run func(string, ...string) http.Request) *Request_WithToken_Call {
	_c.Call.Return(run)
	return _c
}

// WithoutHeader provides a mock function with given fields: key
func (_m *Request) WithoutHeader(key string) http.Request {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for WithoutHeader")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func(string) http.Request); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithoutHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithoutHeader'
type Request_WithoutHeader_Call struct {
	*mock.Call
}

// WithoutHeader is a helper method to define mock.On call
//   - key string
func (_e *Request_Expecter) WithoutHeader(key interface{}) *Request_WithoutHeader_Call {
	return &Request_WithoutHeader_Call{Call: _e.mock.On("WithoutHeader", key)}
}

func (_c *Request_WithoutHeader_Call) Run(run func(key string)) *Request_WithoutHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Request_WithoutHeader_Call) Return(_a0 http.Request) *Request_WithoutHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithoutHeader_Call) RunAndReturn(run func(string) http.Request) *Request_WithoutHeader_Call {
	_c.Call.Return(run)
	return _c
}

// WithoutToken provides a mock function with no fields
func (_m *Request) WithoutToken() http.Request {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for WithoutToken")
	}

	var r0 http.Request
	if rf, ok := ret.Get(0).(func() http.Request); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Request)
		}
	}

	return r0
}

// Request_WithoutToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithoutToken'
type Request_WithoutToken_Call struct {
	*mock.Call
}

// WithoutToken is a helper method to define mock.On call
func (_e *Request_Expecter) WithoutToken() *Request_WithoutToken_Call {
	return &Request_WithoutToken_Call{Call: _e.mock.On("WithoutToken")}
}

func (_c *Request_WithoutToken_Call) Run(run func()) *Request_WithoutToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Request_WithoutToken_Call) Return(_a0 http.Request) *Request_WithoutToken_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithoutToken_Call) RunAndReturn(run func() http.Request) *Request_WithoutToken_Call {
	_c.Call.Return(run)
	return _c
}

// NewRequest creates a new instance of Request. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRequest(t interface {
	mock.TestingT
	Cleanup(func())
}) *Request {
	mock := &Request{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
