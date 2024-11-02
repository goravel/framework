package testing

type TestResponse interface {
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
	AssertHeaderMissing(headerName string) TestResponse
	AssertCookie(name, value string) TestResponse
	AssertCookieExpired(name string) TestResponse
	AssertCookieNotExpired(name string) TestResponse
	AssertCookieMissing(name string) TestResponse
}
