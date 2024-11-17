package testing

import (
	"context"
	"encoding/base64"
	"fmt"
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
	t                 *testing.T
	ctx               context.Context
	defaultHeaders    map[string]string
	defaultCookies    map[string]string
	sessionAttributes map[string]any
}

func NewTestRequest(t *testing.T) contractstesting.TestRequest {
	return &TestRequest{
		t:                 t,
		ctx:               context.Background(),
		defaultHeaders:    make(map[string]string),
		defaultCookies:    make(map[string]string),
		sessionAttributes: make(map[string]any),
	}
}

func (r *TestRequest) FlushHeaders() contractstesting.TestRequest {
	r.defaultHeaders = make(map[string]string)
	return r
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

func (r *TestRequest) WithToken(token string, ttype ...string) contractstesting.TestRequest {
	tt := "Bearer"
	if len(ttype) > 0 {
		tt = ttype[0]
	}
	return r.WithHeader("Authorization", fmt.Sprintf("%s %s", tt, token))
}

func (r *TestRequest) WithBasicAuth(username, password string) contractstesting.TestRequest {
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	return r.WithToken(encoded, "Basic")
}

func (r *TestRequest) WithoutToken() contractstesting.TestRequest {
	return r.WithoutHeader("Authorization")
}

func (r *TestRequest) WithSession(attributes map[string]any) contractstesting.TestRequest {
	r.sessionAttributes = collect.Merge(r.sessionAttributes, attributes)
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
	err := r.setSession()
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(method, uri, body).WithContext(r.ctx)

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

func (r *TestRequest) setSession() error {
	if len(r.sessionAttributes) == 0 {
		return nil
	}

	if sessionFacade == nil {
		return errors.SessionFacadeNotSet
	}

	// Retrieve session driver
	driver, err := sessionFacade.Driver()
	if err != nil {
		return err
	}

	// Build session
	session, err := sessionFacade.BuildSession(driver)
	if err != nil {
		return err
	}

	for key, value := range r.sessionAttributes {
		session.Put(key, value)
	}

	r.WithCookie(session.GetName(), session.GetID())

	if err = session.Save(); err != nil {
		return err
	}

	// Release session
	sessionFacade.ReleaseSession(session)
	return nil
}
