package testing

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/collect"
	"github.com/goravel/framework/support/maps"
)

type TestRequest struct {
	t              *testing.T
	ctx            context.Context
	defaultHeaders map[string]string
	defaultCookies map[string]string
}

func NewTestRequest(t *testing.T) contractstesting.TestRequest {
	return &TestRequest{
		t:              t,
		ctx:            context.Background(),
		defaultHeaders: make(map[string]string),
		defaultCookies: make(map[string]string),
	}
}

func (r *TestRequest) WithHeaders(headers map[string]string) contractstesting.TestRequest {
	r.defaultHeaders = collect.Merge(r.defaultHeaders, headers)
	return r
}

func (r *TestRequest) WithHeader(key, value string) contractstesting.TestRequest {
	maps.Set(r.defaultHeaders, key, value)
	return r
}

func (r *TestRequest) WithoutHeader(key string) contractstesting.TestRequest {
	maps.Forget(r.defaultHeaders, key)
	return r
}

func (r *TestRequest) WithCookies(cookies map[string]string) contractstesting.TestRequest {
	r.defaultCookies = collect.Merge(r.defaultCookies, cookies)
	return r
}

func (r *TestRequest) WithCookie(key, value string) contractstesting.TestRequest {
	maps.Set(r.defaultCookies, key, value)
	return r
}

func (r *TestRequest) WithContext(ctx context.Context) contractstesting.TestRequest {
	r.ctx = ctx
	return r
}

func (r *TestRequest) Get(uri string) (contractstesting.TestResponse, error) {
	return r.call(http.MethodGet, uri, nil)
}

func (r *TestRequest) Post(uri string, body io.Reader) (contractstesting.TestResponse, error) {
	return r.call(http.MethodPost, uri, body)
}

func (r *TestRequest) Put(uri string, body io.Reader) (contractstesting.TestResponse, error) {
	return r.call(http.MethodPut, uri, body)
}

func (r *TestRequest) Patch(uri string, body io.Reader) (contractstesting.TestResponse, error) {
	return r.call(http.MethodPatch, uri, body)
}

func (r *TestRequest) Delete(uri string, body io.Reader) (contractstesting.TestResponse, error) {
	return r.call(http.MethodDelete, uri, body)
}

func (r *TestRequest) Head(uri string) (contractstesting.TestResponse, error) {
	return r.call(http.MethodHead, uri, nil)
}

func (r *TestRequest) Options(uri string) (contractstesting.TestResponse, error) {
	return r.call(http.MethodOptions, uri, nil)
}

func (r *TestRequest) call(method string, uri string, body io.Reader) (contractstesting.TestResponse, error) {
	req := httptest.NewRequestWithContext(r.ctx, method, uri, body)

	for key, value := range r.defaultHeaders {
		req.Header.Set(key, value)
	}

	for name, value := range r.defaultCookies {
		cookie := http.Cookie{Name: name, Value: value}
		req.AddCookie(&cookie)
	}

	if routeFacade == nil {
		r.t.Fatal(errors.RouteFacadeNotSet.SetModule(errors.ModuleTesting))
	}

	response, err := routeFacade.Test(req)
	if err != nil {
		return nil, err
	}

	return NewTestResponse(r.t, response), nil
}
