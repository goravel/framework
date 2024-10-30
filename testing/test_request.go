package testing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cast"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/collect"
	"github.com/goravel/framework/support/maps"
)

type TestRequest struct {
	t              *testing.T
	defaultHeaders map[string]string
	defaultCookies map[string]any
}

func NewTestRequest(t *testing.T) contractstesting.TestRequest {
	return &TestRequest{
		t:              t,
		defaultHeaders: make(map[string]string),
		defaultCookies: make(map[string]any),
	}
}

func (r *TestRequest) WithHeaders(headers map[string]string) contractstesting.TestRequest {
	collect.Merge(r.defaultHeaders, headers)
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

func (r *TestRequest) WithCookies(cookies map[string]any) contractstesting.TestRequest {
	collect.Merge(r.defaultCookies, cookies)
	return r
}

func (r *TestRequest) WithCookie(key string, value any) contractstesting.TestRequest {
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
		cookie := http.Cookie{Name: name, Value: cast.ToString(value)}
		req.AddCookie(&cookie)
	}

	recorder := httptest.NewRecorder()

	if routeFacade == nil {
		panic(errors.RouteFacadeNotSet.SetModule(errors.ModuleTesting))
	}

	routeFacade.ServeHTTP(recorder, req)

	return NewTestResponse(r.t, recorder.Result()), nil
}
