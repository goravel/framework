package testing

type TestResponse interface {
	IsSuccessful() bool
	IsServerError() bool
	Content() (string, error)
	Json() (map[string]any, error)
	AssertStatus(status int) TestResponse
	AssertOk() TestResponse
	AssertCreated() TestResponse
	AssertAccepted() TestResponse
	AssertNoContent(status ...int) TestResponse
	AssertMovedPermanently() TestResponse
	AssertFound() TestResponse
	AssertNotModified() TestResponse
	AssertPartialContent() TestResponse
	AssertTemporaryRedirect() TestResponse
	AssertBadRequest() TestResponse
	AssertUnauthorized() TestResponse
	AssertPaymentRequired() TestResponse
	AssertForbidden() TestResponse
	AssertNotFound() TestResponse
	AssertMethodNotAllowed() TestResponse
	AssertNotAcceptable() TestResponse
	AssertConflict() TestResponse
	AssertRequestTimeout() TestResponse
	AssertGone() TestResponse
	AssertUnsupportedMediaType() TestResponse
	AssertUnprocessableEntity() TestResponse
	AssertTooManyRequests() TestResponse
	AssertInternalServerError() TestResponse
	AssertServiceUnavailable() TestResponse
	AssertHeader(headerName, value string) TestResponse
	AssertHeaderMissing(string) TestResponse
	AssertCookie(name, value string) TestResponse
	AssertCookieExpired(string) TestResponse
	AssertCookieNotExpired(string) TestResponse
	AssertCookieMissing(string) TestResponse
	AssertSuccessful() TestResponse
	AssertServerError() TestResponse
	AssertDontSee([]string, ...bool) TestResponse
	AssertSee([]string, ...bool) TestResponse
	AssertSeeInOrder([]string, ...bool) TestResponse
	AssertFluentJson(func(json AssertableJSON)) TestResponse
}
