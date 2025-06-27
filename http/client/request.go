package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	supportmaps "github.com/goravel/framework/support/maps"
)

var _ client.Request = (*Request)(nil)

type Request struct {
	ctx         context.Context
	bind        any
	json        foundation.Json
	client      *http.Client
	config      *client.Config
	headers     http.Header
	queryParams url.Values
	urlParams   map[string]string
	cookies     []*http.Cookie
}

func NewRequest(config *client.Config, json foundation.Json) *Request {
	return &Request{
		ctx:         context.Background(),
		config:      config,
		client:      getHttpClient(config),
		headers:     http.Header{},
		cookies:     []*http.Cookie{},
		queryParams: url.Values{},
		urlParams:   map[string]string{},
		json:        json,
	}
}

func (r *Request) Get(uri string) (client.Response, error) {
	return r.doRequest(http.MethodGet, uri, nil)
}

func (r *Request) Post(uri string, body io.Reader) (client.Response, error) {
	return r.doRequest(http.MethodPost, uri, body)
}

func (r *Request) Put(uri string, body io.Reader) (client.Response, error) {
	return r.doRequest(http.MethodPut, uri, body)
}

func (r *Request) Delete(uri string, body io.Reader) (client.Response, error) {
	return r.doRequest(http.MethodDelete, uri, body)
}

func (r *Request) Patch(uri string, body io.Reader) (client.Response, error) {
	return r.doRequest(http.MethodPatch, uri, body)
}

func (r *Request) Head(uri string) (client.Response, error) {
	return r.doRequest(http.MethodHead, uri, nil)
}

func (r *Request) Options(uri string) (client.Response, error) {
	return r.doRequest(http.MethodOptions, uri, nil)
}

func (r *Request) Accept(contentType string) client.Request {
	return r.WithHeader("Accept", contentType)
}

func (r *Request) AcceptJSON() client.Request {
	return r.Accept("application/json")
}

func (r *Request) AsForm() client.Request {
	return r.WithHeader("Content-Type", "application/x-www-form-urlencoded")
}

func (r *Request) Bind(value any) client.Request {
	r.bind = value
	return r
}

func (r *Request) Clone() client.Request {
	clone := *r
	clone.headers = r.headers.Clone()
	copy(clone.cookies, r.cookies)
	clone.queryParams = url.Values{}
	for k, v := range r.queryParams {
		clone.queryParams[k] = append([]string{}, v...)
	}

	clone.urlParams = make(map[string]string)
	maps.Copy(clone.urlParams, r.urlParams)

	return &clone
}

func (r *Request) FlushHeaders() client.Request {
	r.headers = make(http.Header)
	return r
}

func (r *Request) ReplaceHeaders(headers map[string]string) client.Request {
	return r.WithHeaders(headers)
}

func (r *Request) WithBasicAuth(username, password string) client.Request {
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	return r.WithToken(encoded, "Basic")
}

func (r *Request) WithContext(ctx context.Context) client.Request {
	r.ctx = ctx
	return r
}

func (r *Request) WithCookies(cookies []*http.Cookie) client.Request {
	r.cookies = append(r.cookies, cookies...)
	return r
}

func (r *Request) WithCookie(cookie *http.Cookie) client.Request {
	r.cookies = append(r.cookies, cookie)
	return r
}

func (r *Request) WithHeader(key, value string) client.Request {
	r.headers.Set(key, value)
	return r
}

func (r *Request) WithHeaders(headers map[string]string) client.Request {
	for k, v := range headers {
		r.WithHeader(k, v)
	}
	return r
}

func (r *Request) WithQueryParameter(key, value string) client.Request {
	r.queryParams.Set(key, value)
	return r
}

func (r *Request) WithQueryParameters(params map[string]string) client.Request {
	for k, v := range params {
		r.WithQueryParameter(k, v)
	}
	return r
}

func (r *Request) WithQueryString(query string) client.Request {
	params, err := url.ParseQuery(strings.TrimSpace(query))
	if err != nil {
		return r
	}

	for k, v := range params {
		for _, vv := range v {
			r.queryParams.Add(k, vv)
		}
	}
	return r
}

func (r *Request) WithoutHeader(key string) client.Request {
	r.headers.Del(key)
	return r
}

func (r *Request) WithToken(token string, ttype ...string) client.Request {
	tt := "Bearer"
	if len(ttype) > 0 {
		tt = ttype[0]
	}
	return r.WithHeader("Authorization", fmt.Sprintf("%s %s", tt, token))
}

func (r *Request) WithoutToken() client.Request {
	return r.WithoutHeader("Authorization")
}

func (r *Request) WithUrlParameter(key, value string) client.Request {
	supportmaps.Set(r.urlParams, key, url.PathEscape(value))
	return r
}

func (r *Request) WithUrlParameters(params map[string]string) client.Request {
	for k, v := range params {
		r.WithUrlParameter(k, v)
	}
	return r
}

func (r *Request) doRequest(method, uri string, body io.Reader) (client.Response, error) {
	parsedURL, err := r.parseRequestURL(uri)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(r.ctx, method, parsedURL, body)
	if err != nil {
		return nil, err
	}

	req.Header = r.headers

	for _, value := range r.cookies {
		req.AddCookie(value)
	}

	res, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	response := NewResponse(res, r.json)
	if r.bind != nil {
		body, err := response.Body()
		if err != nil {
			return nil, err
		}

		if err := r.json.UnmarshalString(body, r.bind); err != nil {
			return nil, err
		}
	}

	return response, nil
}

func (r *Request) parseRequestURL(uri string) (string, error) {
	baseURL := r.config.BaseUrl

	// Prepend base URL if needed
	if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
		uri = strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(uri, "/")
	}

	var buf strings.Builder
	buf.Grow(len(uri) + 10)

	n := len(uri)
	i := 0
	for i < n {
		if uri[i] == '{' {
			j := i + 1
			for j < n && uri[j] != '}' {
				j++
			}

			if j == n {
				buf.WriteString(uri[i:])
				break
			}

			key := uri[i+1 : j]
			if value, found := r.urlParams[key]; found {
				buf.WriteString(value)
			} else {
				buf.WriteString(uri[i : j+1])
			}

			i = j + 1
		} else {
			start := i
			for i < n && uri[i] != '{' {
				i++
			}
			buf.WriteString(uri[start:i])
		}
	}

	reqURL, err := url.Parse(buf.String())
	if err != nil {
		return "", err
	}

	if len(r.queryParams) > 0 {
		if len(strings.TrimSpace(reqURL.RawQuery)) == 0 {
			reqURL.RawQuery = r.queryParams.Encode()
		} else {
			reqURL.RawQuery = reqURL.RawQuery + "&" + r.queryParams.Encode()
		}
	}

	return reqURL.String(), nil
}
