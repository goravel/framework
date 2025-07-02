package http

import (
	"context"
	"io"
	"net/http"
)

type Request interface {
	Get(uri string) (Response, error)
	Post(uri string, body io.Reader) (Response, error)
	Put(uri string, body io.Reader) (Response, error)
	Delete(uri string, body io.Reader) (Response, error)
	Patch(uri string, body io.Reader) (Response, error)
	Head(uri string) (Response, error)
	Options(uri string) (Response, error)

	Bind(value any) Request
	FlushHeaders() Request
	WithBasicAuth(username, password string) Request
	WithContext(ctx context.Context) Request
	WithCookies(cookies []*http.Cookie) Request
	WithCookie(cookie *http.Cookie) Request
	WithHeader(key, value string) Request
	WithHeaders(headers map[string]string) Request
	WithoutHeader(key string) Request
	WithToken(token string, ttype ...string) Request
	WithoutToken() Request
	WithSession(attributes map[string]any) Request
}
