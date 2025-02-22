package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/support/collect"
	"github.com/goravel/framework/support/maps"
	"github.com/goravel/framework/support/str"
)

var _ client.Request = &requestImpl{}

type requestImpl struct {
	ctx         context.Context
	client      *http.Client
	config      config.Config
	bind        any
	headers     http.Header
	cookies     []*http.Cookie
	queryParams url.Values
	urlParams   map[string]string
	json        foundation.Json
	timeout     time.Duration
}

func NewRequest(config config.Config, json foundation.Json) client.Request {
	return &requestImpl{
		ctx:         context.Background(),
		config:      config,
		client:      getHttpClient(config),
		headers:     http.Header{},
		cookies:     []*http.Cookie{},
		queryParams: url.Values{},
		urlParams:   map[string]string{},
		json:        json,
		timeout:     config.GetDuration("http.client.timeout", 30*time.Second),
	}
}

func getHttpClient(config config.Config) *http.Client {
	return &http.Client{
		Timeout: config.GetDuration("http.client.timeout"),
		Transport: &http.Transport{
			MaxIdleConns:        config.GetInt("http.client.max_idle_conns"),
			MaxIdleConnsPerHost: config.GetInt("http.client.max_idle_conns_per_host"),
			MaxConnsPerHost:     config.GetInt("http.client.max_conns_per_host"),
			IdleConnTimeout:     config.GetDuration("http.client.idle_conn_timeout"),
		},
	}
}

func (r *requestImpl) doRequest(method, uri string, body io.Reader) (client.Response, error) {
	ctx, cancel := context.WithTimeout(r.ctx, r.timeout)
	defer cancel()

	if !str.Of(uri).StartsWith("/", "http://", "https://") {
		uri = "/" + uri
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header = r.headers
	for key, value := range r.urlParams {
		req.SetPathValue(key, value)
	}

	for _, value := range r.cookies {
		req.AddCookie(value)
	}

	res, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	return NewResponse(res, r.json), nil
}

func (r *requestImpl) Get(uri string) (client.Response, error) {
	return r.doRequest(http.MethodGet, uri, nil)
}

func (r *requestImpl) Post(uri string, body io.Reader) (client.Response, error) {
	return r.doRequest(http.MethodPost, uri, body)
}

func (r *requestImpl) Put(uri string, body io.Reader) (client.Response, error) {
	return r.doRequest(http.MethodPut, uri, body)
}

func (r *requestImpl) Delete(uri string, body io.Reader) (client.Response, error) {
	return r.doRequest(http.MethodDelete, uri, body)
}

func (r *requestImpl) Patch(uri string, body io.Reader) (client.Response, error) {
	return r.doRequest(http.MethodPatch, uri, body)
}

func (r *requestImpl) Head(uri string) (client.Response, error) {
	return r.doRequest(http.MethodHead, uri, nil)
}

func (r *requestImpl) Options(uri string) (client.Response, error) {
	return r.doRequest(http.MethodOptions, uri, nil)
}

func (r *requestImpl) Accept(contentType string) client.Request {
	return r.WithHeader("Accept", contentType)
}

func (r *requestImpl) AcceptJSON() client.Request {
	return r.Accept("application/json")
}

func (r *requestImpl) AsForm() client.Request {
	return r.WithHeader("Content-Type", "application/x-www-form-urlencoded")
}

func (r *requestImpl) Bind(value any) client.Request {
	r.bind = value
	return r
}

func (r *requestImpl) Clone() client.Request {
	clone := *r
	clone.headers = r.headers.Clone()
	copy(clone.cookies, r.cookies)
	clone.queryParams = url.Values{}
	for k, v := range r.queryParams {
		clone.queryParams[k] = append([]string{}, v...)
	}

	clone.urlParams = make(map[string]string)
	for k, v := range r.urlParams {
		clone.urlParams[k] = v
	}

	return &clone
}

func (r *requestImpl) FlushHeaders() client.Request {
	r.headers = make(http.Header)
	return r
}

func (r *requestImpl) ReplaceHeaders(headers map[string]string) client.Request {
	return r.WithHeaders(headers)
}

func (r *requestImpl) Timeout(duration time.Duration) client.Request {
	r.timeout = duration
	return r
}

func (r *requestImpl) WithBasicAuth(username, password string) client.Request {
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	return r.WithToken(encoded, "Basic")
}

func (r *requestImpl) WithContext(ctx context.Context) client.Request {
	r.ctx = ctx
	return r
}

func (r *requestImpl) WithCookies(cookies []*http.Cookie) client.Request {
	r.cookies = append(r.cookies, cookies...)
	return r
}

func (r *requestImpl) WithCookie(cookie *http.Cookie) client.Request {
	r.cookies = append(r.cookies, cookie)
	return r
}

func (r *requestImpl) WithDigestAuth(username, password string) client.Request {
	return r
}

func (r *requestImpl) WithHeader(key, value string) client.Request {
	r.headers.Set(key, value)
	return r
}

func (r *requestImpl) WithHeaders(headers map[string]string) client.Request {
	for k, v := range headers {
		r.WithHeader(k, v)
	}
	return r
}

func (r *requestImpl) WithQueryParameter(key, value string) client.Request {
	r.queryParams.Set(key, value)
	return r
}

func (r *requestImpl) WithQueryParameters(params map[string]string) client.Request {
	for k, v := range params {
		r.WithQueryParameter(k, v)
	}
	return r
}

func (r *requestImpl) WithQueryString(query string) client.Request {
	params, err := url.ParseQuery(strings.TrimSpace(query))
	if err != nil {
		return r
	}

	for k, v := range params {
		for _, vv := range v {
			r.WithQueryParameter(k, vv)
		}
	}
	return r
}

func (r *requestImpl) WithoutHeader(key string) client.Request {
	r.headers.Del(key)
	return r
}

func (r *requestImpl) WithToken(token string, ttype ...string) client.Request {
	tt := "Bearer"
	if len(ttype) > 0 {
		tt = ttype[0]
	}
	return r.WithHeader("Authorization", fmt.Sprintf("%s %s", tt, token))
}

func (r *requestImpl) WithoutToken() client.Request {
	return r.WithoutHeader("Authorization")
}

func (r *requestImpl) WithUrlParameter(key, value string) client.Request {
	maps.Set(r.urlParams, key, url.QueryEscape(value))
	return r
}

func (r *requestImpl) WithUrlParameters(params map[string]string) client.Request {
	r.urlParams = collect.Merge(r.urlParams, params)
	return r
}
