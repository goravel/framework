package testing

import (
	"net/http"
	"testing"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/collect"
	"github.com/goravel/framework/support/maps"
)

type TestRequest struct {
	t              *testing.T
	defaultHeaders map[string]string
	defaultCookies map[string]string
}

func NewTestRequest(t *testing.T) contractstesting.TestRequest {
	return &TestRequest{
		t:              t,
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

func (r *TestRequest) Get(uri string) (contractstesting.TestResponse, error) {
	return r.call(http.MethodGet, uri)
}

func (r *TestRequest) call(method string, uri string) (contractstesting.TestResponse, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}

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
