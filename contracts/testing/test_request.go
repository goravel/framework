package testing

type TestRequest interface {
	WithHeaders(headers map[string]string) TestRequest
	WithHeader(key, value string) TestRequest
	WithoutHeader(key string) TestRequest
	WithCookies(cookies map[string]string) TestRequest
	WithCookie(key, value string) TestRequest
	Get(uri string, headers ...map[string]string) (TestResponse, error)
}
