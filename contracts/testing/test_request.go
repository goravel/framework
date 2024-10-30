package testing

type TestRequest interface {
	WithHeaders(headers map[string]string) TestRequest
	WithHeader(key, value string) TestRequest
	WithoutHeader(key string) TestRequest
	WithCookies(cookies map[string]any) TestRequest
	WithCookie(key string, value any) TestRequest
	Get(uri string) (TestResponse, error)
}
