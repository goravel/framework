package testing

import (
	"context"
	"io"
)

type TestRequest interface {
	Get(uri string) (TestResponse, error)
	Post(uri string, body io.Reader) (TestResponse, error)
	Put(uri string, body io.Reader) (TestResponse, error)
	Delete(uri string, body io.Reader) (TestResponse, error)
	Patch(uri string, body io.Reader) (TestResponse, error)
	Head(uri string) (TestResponse, error)
	Options(uri string) (TestResponse, error)

	Bind(value any) TestRequest
	FlushHeaders() TestRequest
	WithBasicAuth(username, password string) TestRequest
	WithContext(ctx context.Context) TestRequest
	WithCookies(cookies map[string]string) TestRequest
	WithCookie(key, value string) TestRequest
	WithHeader(key, value string) TestRequest
	WithHeaders(headers map[string]string) TestRequest
	WithoutHeader(key string) TestRequest
	WithToken(token string, ttype ...string) TestRequest
	WithoutToken() TestRequest
	WithSession(attributes map[string]any) TestRequest
}
