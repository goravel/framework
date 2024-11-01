package testing

import (
	"context"
	"io"
)

type TestRequest interface {
	WithHeaders(headers map[string]string) TestRequest
	WithHeader(key, value string) TestRequest
	WithoutHeader(key string) TestRequest
	WithCookies(cookies map[string]string) TestRequest
	WithCookie(key, value string) TestRequest
	WithContext(ctx context.Context) TestRequest
	Get(uri string) (TestResponse, error)
	Post(uri string, body io.Reader) (TestResponse, error)
	Put(uri string, body io.Reader) (TestResponse, error)
	Patch(uri string, body io.Reader) (TestResponse, error)
	Delete(uri string, body io.Reader) (TestResponse, error)
	Head(uri string, body io.Reader) (TestResponse, error)
	Options(uri string) (TestResponse, error)
}
