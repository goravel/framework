package client

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Request interface {
	Get(uri string) (Response, error)
	Post(uri string, body io.Reader) (Response, error)
	Put(uri string, body io.Reader) (Response, error)
	Delete(uri string, body io.Reader) (Response, error)
	Patch(uri string, body io.Reader) (Response, error)
	Head(uri string) (Response, error)
	Options(uri string) (Response, error)

	Accept(contentType string) Request
	AcceptJSON() Request
	AsForm() Request
	//Attach(key string, value any) Request
	Bind(value any) Request
	ConnectTimeout(duration time.Duration) Request
	FlushHeaders() Request
	ReplaceHeaders(headers map[string]string) Request
	Retry(times int, sleep time.Duration) Request
	Timeout(duration time.Duration) Request
	WithBasicAuth(username, password string) Request
	WithContext(ctx context.Context) Request
	WithCookies(cookies []*http.Cookie) Request
	WithCookie(cookie *http.Cookie) Request
	WithDigestAuth(username, password string) Request
	WithHeader(key, value string) Request
	WithHeaders(map[string]string) Request
	WithQueryParameter(key, value string) Request
	WithQueryParameters(map[string]string) Request
	WithQueryString(query string) Request
	WithoutHeader(key string) Request
	WithToken(token string, ttype ...string) Request
	WithoutToken() Request
	WithUrlParameters(map[string]string) Request
}
