package testing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	contractstesting "github.com/goravel/framework/contracts/testing"
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

func (m *TestRequest) WithHeaders(headers map[string]string) contractstesting.TestRequest {
	collect.Merge(m.defaultHeaders, headers)
	return m
}

func (m *TestRequest) WithHeader(key, value string) contractstesting.TestRequest {
	maps.Set(m.defaultHeaders, key, value)
	return m
}

func (m *TestRequest) WithoutHeader(key string) contractstesting.TestRequest {
	maps.Forget(m.defaultHeaders, key)
	return m
}

func (m *TestRequest) WithCookies(cookies map[string]string) contractstesting.TestRequest {
	collect.Merge(m.defaultCookies, cookies)
	return m
}

func (m *TestRequest) WithCookie(key, value string) contractstesting.TestRequest {
	maps.Set(m.defaultCookies, key, value)
	return m
}

func (m *TestRequest) Get(uri string, headers ...map[string]string) (contractstesting.TestResponse, error) {
	return m.call(http.MethodGet, uri, headers...)
}

func (m *TestRequest) call(method string, uri string, headers ...map[string]string) (contractstesting.TestResponse, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range m.defaultHeaders {
		req.Header.Set(key, value)
	}
	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Set(key, value)
		}
	}

	for _, cookie := range m.defaultCookies {
		c := http.Cookie{Name: cookie, Value: cookie}
		req.AddCookie(&c)
	}

	recorder := httptest.NewRecorder()

	routeFacade.ServeHTTP(recorder, req)

	return NewTestResponse(m.t, recorder.Result()), nil
}
