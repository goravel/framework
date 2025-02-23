// Code generated by mockery. DO NOT EDIT.

package client

import (
	context "context"

	client "github.com/goravel/framework/contracts/http/client"

	http "net/http"

	io "io"

	mock "github.com/stretchr/testify/mock"
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

// Accept provides a mock function with given fields: contentType
func (_m *Request) Accept(contentType string) client.Request {
	ret := _m.Called(contentType)

	if len(ret) == 0 {
		panic("no return value specified for Accept")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(string) client.Request); ok {
		r0 = rf(contentType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_Accept_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Accept'
type Request_Accept_Call struct {
	*mock.Call
}

// Accept is a helper method to define mock.On call
//   - contentType string
func (_e *Request_Expecter) Accept(contentType interface{}) *Request_Accept_Call {
	return &Request_Accept_Call{Call: _e.mock.On("Accept", contentType)}
}

func (_c *Request_Accept_Call) Run(run func(contentType string)) *Request_Accept_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Request_Accept_Call) Return(_a0 client.Request) *Request_Accept_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_Accept_Call) RunAndReturn(run func(string) client.Request) *Request_Accept_Call {
	_c.Call.Return(run)
	return _c
}

// AcceptJSON provides a mock function with no fields
func (_m *Request) AcceptJSON() client.Request {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for AcceptJSON")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func() client.Request); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_AcceptJSON_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AcceptJSON'
type Request_AcceptJSON_Call struct {
	*mock.Call
}

// AcceptJSON is a helper method to define mock.On call
func (_e *Request_Expecter) AcceptJSON() *Request_AcceptJSON_Call {
	return &Request_AcceptJSON_Call{Call: _e.mock.On("AcceptJSON")}
}

func (_c *Request_AcceptJSON_Call) Run(run func()) *Request_AcceptJSON_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Request_AcceptJSON_Call) Return(_a0 client.Request) *Request_AcceptJSON_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_AcceptJSON_Call) RunAndReturn(run func() client.Request) *Request_AcceptJSON_Call {
	_c.Call.Return(run)
	return _c
}

// AsForm provides a mock function with no fields
func (_m *Request) AsForm() client.Request {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for AsForm")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func() client.Request); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_AsForm_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AsForm'
type Request_AsForm_Call struct {
	*mock.Call
}

// AsForm is a helper method to define mock.On call
func (_e *Request_Expecter) AsForm() *Request_AsForm_Call {
	return &Request_AsForm_Call{Call: _e.mock.On("AsForm")}
}

func (_c *Request_AsForm_Call) Run(run func()) *Request_AsForm_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Request_AsForm_Call) Return(_a0 client.Request) *Request_AsForm_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_AsForm_Call) RunAndReturn(run func() client.Request) *Request_AsForm_Call {
	_c.Call.Return(run)
	return _c
}

// Bind provides a mock function with given fields: value
func (_m *Request) Bind(value interface{}) client.Request {
	ret := _m.Called(value)

	if len(ret) == 0 {
		panic("no return value specified for Bind")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(interface{}) client.Request); ok {
		r0 = rf(value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
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

func (_c *Request_Bind_Call) Return(_a0 client.Request) *Request_Bind_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_Bind_Call) RunAndReturn(run func(interface{}) client.Request) *Request_Bind_Call {
	_c.Call.Return(run)
	return _c
}

// Clone provides a mock function with no fields
func (_m *Request) Clone() client.Request {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Clone")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func() client.Request); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_Clone_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Clone'
type Request_Clone_Call struct {
	*mock.Call
}

// Clone is a helper method to define mock.On call
func (_e *Request_Expecter) Clone() *Request_Clone_Call {
	return &Request_Clone_Call{Call: _e.mock.On("Clone")}
}

func (_c *Request_Clone_Call) Run(run func()) *Request_Clone_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Request_Clone_Call) Return(_a0 client.Request) *Request_Clone_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_Clone_Call) RunAndReturn(run func() client.Request) *Request_Clone_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: uri, body
func (_m *Request) Delete(uri string, body io.Reader) (client.Response, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 client.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (client.Response, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) client.Response); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Response)
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

func (_c *Request_Delete_Call) Return(_a0 client.Response, _a1 error) *Request_Delete_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Delete_Call) RunAndReturn(run func(string, io.Reader) (client.Response, error)) *Request_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// FlushHeaders provides a mock function with no fields
func (_m *Request) FlushHeaders() client.Request {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for FlushHeaders")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func() client.Request); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
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

func (_c *Request_FlushHeaders_Call) Return(_a0 client.Request) *Request_FlushHeaders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_FlushHeaders_Call) RunAndReturn(run func() client.Request) *Request_FlushHeaders_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: uri
func (_m *Request) Get(uri string) (client.Response, error) {
	ret := _m.Called(uri)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 client.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (client.Response, error)); ok {
		return rf(uri)
	}
	if rf, ok := ret.Get(0).(func(string) client.Response); ok {
		r0 = rf(uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Response)
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

func (_c *Request_Get_Call) Return(_a0 client.Response, _a1 error) *Request_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Get_Call) RunAndReturn(run func(string) (client.Response, error)) *Request_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Head provides a mock function with given fields: uri
func (_m *Request) Head(uri string) (client.Response, error) {
	ret := _m.Called(uri)

	if len(ret) == 0 {
		panic("no return value specified for Head")
	}

	var r0 client.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (client.Response, error)); ok {
		return rf(uri)
	}
	if rf, ok := ret.Get(0).(func(string) client.Response); ok {
		r0 = rf(uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Response)
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

func (_c *Request_Head_Call) Return(_a0 client.Response, _a1 error) *Request_Head_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Head_Call) RunAndReturn(run func(string) (client.Response, error)) *Request_Head_Call {
	_c.Call.Return(run)
	return _c
}

// Options provides a mock function with given fields: uri
func (_m *Request) Options(uri string) (client.Response, error) {
	ret := _m.Called(uri)

	if len(ret) == 0 {
		panic("no return value specified for Options")
	}

	var r0 client.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (client.Response, error)); ok {
		return rf(uri)
	}
	if rf, ok := ret.Get(0).(func(string) client.Response); ok {
		r0 = rf(uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Response)
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

func (_c *Request_Options_Call) Return(_a0 client.Response, _a1 error) *Request_Options_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Options_Call) RunAndReturn(run func(string) (client.Response, error)) *Request_Options_Call {
	_c.Call.Return(run)
	return _c
}

// Patch provides a mock function with given fields: uri, body
func (_m *Request) Patch(uri string, body io.Reader) (client.Response, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Patch")
	}

	var r0 client.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (client.Response, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) client.Response); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Response)
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

func (_c *Request_Patch_Call) Return(_a0 client.Response, _a1 error) *Request_Patch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Patch_Call) RunAndReturn(run func(string, io.Reader) (client.Response, error)) *Request_Patch_Call {
	_c.Call.Return(run)
	return _c
}

// Post provides a mock function with given fields: uri, body
func (_m *Request) Post(uri string, body io.Reader) (client.Response, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Post")
	}

	var r0 client.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (client.Response, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) client.Response); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Response)
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

func (_c *Request_Post_Call) Return(_a0 client.Response, _a1 error) *Request_Post_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Post_Call) RunAndReturn(run func(string, io.Reader) (client.Response, error)) *Request_Post_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: uri, body
func (_m *Request) Put(uri string, body io.Reader) (client.Response, error) {
	ret := _m.Called(uri, body)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 client.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (client.Response, error)); ok {
		return rf(uri, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) client.Response); ok {
		r0 = rf(uri, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Response)
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

func (_c *Request_Put_Call) Return(_a0 client.Response, _a1 error) *Request_Put_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Request_Put_Call) RunAndReturn(run func(string, io.Reader) (client.Response, error)) *Request_Put_Call {
	_c.Call.Return(run)
	return _c
}

// ReplaceHeaders provides a mock function with given fields: headers
func (_m *Request) ReplaceHeaders(headers map[string]string) client.Request {
	ret := _m.Called(headers)

	if len(ret) == 0 {
		panic("no return value specified for ReplaceHeaders")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(map[string]string) client.Request); ok {
		r0 = rf(headers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_ReplaceHeaders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReplaceHeaders'
type Request_ReplaceHeaders_Call struct {
	*mock.Call
}

// ReplaceHeaders is a helper method to define mock.On call
//   - headers map[string]string
func (_e *Request_Expecter) ReplaceHeaders(headers interface{}) *Request_ReplaceHeaders_Call {
	return &Request_ReplaceHeaders_Call{Call: _e.mock.On("ReplaceHeaders", headers)}
}

func (_c *Request_ReplaceHeaders_Call) Run(run func(headers map[string]string)) *Request_ReplaceHeaders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]string))
	})
	return _c
}

func (_c *Request_ReplaceHeaders_Call) Return(_a0 client.Request) *Request_ReplaceHeaders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_ReplaceHeaders_Call) RunAndReturn(run func(map[string]string) client.Request) *Request_ReplaceHeaders_Call {
	_c.Call.Return(run)
	return _c
}

// WithBasicAuth provides a mock function with given fields: username, password
func (_m *Request) WithBasicAuth(username string, password string) client.Request {
	ret := _m.Called(username, password)

	if len(ret) == 0 {
		panic("no return value specified for WithBasicAuth")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(string, string) client.Request); ok {
		r0 = rf(username, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
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

func (_c *Request_WithBasicAuth_Call) Return(_a0 client.Request) *Request_WithBasicAuth_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithBasicAuth_Call) RunAndReturn(run func(string, string) client.Request) *Request_WithBasicAuth_Call {
	_c.Call.Return(run)
	return _c
}

// WithContext provides a mock function with given fields: ctx
func (_m *Request) WithContext(ctx context.Context) client.Request {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for WithContext")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(context.Context) client.Request); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
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

func (_c *Request_WithContext_Call) Return(_a0 client.Request) *Request_WithContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithContext_Call) RunAndReturn(run func(context.Context) client.Request) *Request_WithContext_Call {
	_c.Call.Return(run)
	return _c
}

// WithCookie provides a mock function with given fields: cookie
func (_m *Request) WithCookie(cookie *http.Cookie) client.Request {
	ret := _m.Called(cookie)

	if len(ret) == 0 {
		panic("no return value specified for WithCookie")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(*http.Cookie) client.Request); ok {
		r0 = rf(cookie)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_WithCookie_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithCookie'
type Request_WithCookie_Call struct {
	*mock.Call
}

// WithCookie is a helper method to define mock.On call
//   - cookie *http.Cookie
func (_e *Request_Expecter) WithCookie(cookie interface{}) *Request_WithCookie_Call {
	return &Request_WithCookie_Call{Call: _e.mock.On("WithCookie", cookie)}
}

func (_c *Request_WithCookie_Call) Run(run func(cookie *http.Cookie)) *Request_WithCookie_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*http.Cookie))
	})
	return _c
}

func (_c *Request_WithCookie_Call) Return(_a0 client.Request) *Request_WithCookie_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithCookie_Call) RunAndReturn(run func(*http.Cookie) client.Request) *Request_WithCookie_Call {
	_c.Call.Return(run)
	return _c
}

// WithCookies provides a mock function with given fields: cookies
func (_m *Request) WithCookies(cookies []*http.Cookie) client.Request {
	ret := _m.Called(cookies)

	if len(ret) == 0 {
		panic("no return value specified for WithCookies")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func([]*http.Cookie) client.Request); ok {
		r0 = rf(cookies)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_WithCookies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithCookies'
type Request_WithCookies_Call struct {
	*mock.Call
}

// WithCookies is a helper method to define mock.On call
//   - cookies []*http.Cookie
func (_e *Request_Expecter) WithCookies(cookies interface{}) *Request_WithCookies_Call {
	return &Request_WithCookies_Call{Call: _e.mock.On("WithCookies", cookies)}
}

func (_c *Request_WithCookies_Call) Run(run func(cookies []*http.Cookie)) *Request_WithCookies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]*http.Cookie))
	})
	return _c
}

func (_c *Request_WithCookies_Call) Return(_a0 client.Request) *Request_WithCookies_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithCookies_Call) RunAndReturn(run func([]*http.Cookie) client.Request) *Request_WithCookies_Call {
	_c.Call.Return(run)
	return _c
}

// WithHeader provides a mock function with given fields: key, value
func (_m *Request) WithHeader(key string, value string) client.Request {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for WithHeader")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(string, string) client.Request); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
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

func (_c *Request_WithHeader_Call) Return(_a0 client.Request) *Request_WithHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithHeader_Call) RunAndReturn(run func(string, string) client.Request) *Request_WithHeader_Call {
	_c.Call.Return(run)
	return _c
}

// WithHeaders provides a mock function with given fields: _a0
func (_m *Request) WithHeaders(_a0 map[string]string) client.Request {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for WithHeaders")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(map[string]string) client.Request); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_WithHeaders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithHeaders'
type Request_WithHeaders_Call struct {
	*mock.Call
}

// WithHeaders is a helper method to define mock.On call
//   - _a0 map[string]string
func (_e *Request_Expecter) WithHeaders(_a0 interface{}) *Request_WithHeaders_Call {
	return &Request_WithHeaders_Call{Call: _e.mock.On("WithHeaders", _a0)}
}

func (_c *Request_WithHeaders_Call) Run(run func(_a0 map[string]string)) *Request_WithHeaders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]string))
	})
	return _c
}

func (_c *Request_WithHeaders_Call) Return(_a0 client.Request) *Request_WithHeaders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithHeaders_Call) RunAndReturn(run func(map[string]string) client.Request) *Request_WithHeaders_Call {
	_c.Call.Return(run)
	return _c
}

// WithQueryParameter provides a mock function with given fields: key, value
func (_m *Request) WithQueryParameter(key string, value string) client.Request {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for WithQueryParameter")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(string, string) client.Request); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_WithQueryParameter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithQueryParameter'
type Request_WithQueryParameter_Call struct {
	*mock.Call
}

// WithQueryParameter is a helper method to define mock.On call
//   - key string
//   - value string
func (_e *Request_Expecter) WithQueryParameter(key interface{}, value interface{}) *Request_WithQueryParameter_Call {
	return &Request_WithQueryParameter_Call{Call: _e.mock.On("WithQueryParameter", key, value)}
}

func (_c *Request_WithQueryParameter_Call) Run(run func(key string, value string)) *Request_WithQueryParameter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Request_WithQueryParameter_Call) Return(_a0 client.Request) *Request_WithQueryParameter_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithQueryParameter_Call) RunAndReturn(run func(string, string) client.Request) *Request_WithQueryParameter_Call {
	_c.Call.Return(run)
	return _c
}

// WithQueryParameters provides a mock function with given fields: _a0
func (_m *Request) WithQueryParameters(_a0 map[string]string) client.Request {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for WithQueryParameters")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(map[string]string) client.Request); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_WithQueryParameters_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithQueryParameters'
type Request_WithQueryParameters_Call struct {
	*mock.Call
}

// WithQueryParameters is a helper method to define mock.On call
//   - _a0 map[string]string
func (_e *Request_Expecter) WithQueryParameters(_a0 interface{}) *Request_WithQueryParameters_Call {
	return &Request_WithQueryParameters_Call{Call: _e.mock.On("WithQueryParameters", _a0)}
}

func (_c *Request_WithQueryParameters_Call) Run(run func(_a0 map[string]string)) *Request_WithQueryParameters_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]string))
	})
	return _c
}

func (_c *Request_WithQueryParameters_Call) Return(_a0 client.Request) *Request_WithQueryParameters_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithQueryParameters_Call) RunAndReturn(run func(map[string]string) client.Request) *Request_WithQueryParameters_Call {
	_c.Call.Return(run)
	return _c
}

// WithQueryString provides a mock function with given fields: query
func (_m *Request) WithQueryString(query string) client.Request {
	ret := _m.Called(query)

	if len(ret) == 0 {
		panic("no return value specified for WithQueryString")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(string) client.Request); ok {
		r0 = rf(query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_WithQueryString_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithQueryString'
type Request_WithQueryString_Call struct {
	*mock.Call
}

// WithQueryString is a helper method to define mock.On call
//   - query string
func (_e *Request_Expecter) WithQueryString(query interface{}) *Request_WithQueryString_Call {
	return &Request_WithQueryString_Call{Call: _e.mock.On("WithQueryString", query)}
}

func (_c *Request_WithQueryString_Call) Run(run func(query string)) *Request_WithQueryString_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Request_WithQueryString_Call) Return(_a0 client.Request) *Request_WithQueryString_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithQueryString_Call) RunAndReturn(run func(string) client.Request) *Request_WithQueryString_Call {
	_c.Call.Return(run)
	return _c
}

// WithToken provides a mock function with given fields: token, ttype
func (_m *Request) WithToken(token string, ttype ...string) client.Request {
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

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(string, ...string) client.Request); ok {
		r0 = rf(token, ttype...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
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

func (_c *Request_WithToken_Call) Return(_a0 client.Request) *Request_WithToken_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithToken_Call) RunAndReturn(run func(string, ...string) client.Request) *Request_WithToken_Call {
	_c.Call.Return(run)
	return _c
}

// WithUrlParameter provides a mock function with given fields: key, value
func (_m *Request) WithUrlParameter(key string, value string) client.Request {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for WithUrlParameter")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(string, string) client.Request); ok {
		r0 = rf(key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_WithUrlParameter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithUrlParameter'
type Request_WithUrlParameter_Call struct {
	*mock.Call
}

// WithUrlParameter is a helper method to define mock.On call
//   - key string
//   - value string
func (_e *Request_Expecter) WithUrlParameter(key interface{}, value interface{}) *Request_WithUrlParameter_Call {
	return &Request_WithUrlParameter_Call{Call: _e.mock.On("WithUrlParameter", key, value)}
}

func (_c *Request_WithUrlParameter_Call) Run(run func(key string, value string)) *Request_WithUrlParameter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Request_WithUrlParameter_Call) Return(_a0 client.Request) *Request_WithUrlParameter_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithUrlParameter_Call) RunAndReturn(run func(string, string) client.Request) *Request_WithUrlParameter_Call {
	_c.Call.Return(run)
	return _c
}

// WithUrlParameters provides a mock function with given fields: _a0
func (_m *Request) WithUrlParameters(_a0 map[string]string) client.Request {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for WithUrlParameters")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(map[string]string) client.Request); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
		}
	}

	return r0
}

// Request_WithUrlParameters_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithUrlParameters'
type Request_WithUrlParameters_Call struct {
	*mock.Call
}

// WithUrlParameters is a helper method to define mock.On call
//   - _a0 map[string]string
func (_e *Request_Expecter) WithUrlParameters(_a0 interface{}) *Request_WithUrlParameters_Call {
	return &Request_WithUrlParameters_Call{Call: _e.mock.On("WithUrlParameters", _a0)}
}

func (_c *Request_WithUrlParameters_Call) Run(run func(_a0 map[string]string)) *Request_WithUrlParameters_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]string))
	})
	return _c
}

func (_c *Request_WithUrlParameters_Call) Return(_a0 client.Request) *Request_WithUrlParameters_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithUrlParameters_Call) RunAndReturn(run func(map[string]string) client.Request) *Request_WithUrlParameters_Call {
	_c.Call.Return(run)
	return _c
}

// WithoutHeader provides a mock function with given fields: key
func (_m *Request) WithoutHeader(key string) client.Request {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for WithoutHeader")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func(string) client.Request); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
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

func (_c *Request_WithoutHeader_Call) Return(_a0 client.Request) *Request_WithoutHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithoutHeader_Call) RunAndReturn(run func(string) client.Request) *Request_WithoutHeader_Call {
	_c.Call.Return(run)
	return _c
}

// WithoutToken provides a mock function with no fields
func (_m *Request) WithoutToken() client.Request {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for WithoutToken")
	}

	var r0 client.Request
	if rf, ok := ret.Get(0).(func() client.Request); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Request)
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

func (_c *Request_WithoutToken_Call) Return(_a0 client.Request) *Request_WithoutToken_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Request_WithoutToken_Call) RunAndReturn(run func() client.Request) *Request_WithoutToken_Call {
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
