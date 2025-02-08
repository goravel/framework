package http

import "net/http"

type Response interface {
	Content() (string, error)
	Cookie(name string) *http.Cookie
	Cookies() []*http.Cookie
	Headers() http.Header
	IsServerError() bool
	IsSuccessful() bool
	Json() (map[string]any, error)
	Session() (map[string]any, error)

	AssertStatus(status int) Response
	AssertOk() Response
	AssertCreated() Response
	AssertAccepted() Response
	AssertNoContent(status ...int) Response
	AssertMovedPermanently() Response
	AssertFound() Response
	AssertNotModified() Response
	AssertPartialContent() Response
	AssertTemporaryRedirect() Response
	AssertBadRequest() Response
	AssertUnauthorized() Response
	AssertPaymentRequired() Response
	AssertForbidden() Response
	AssertNotFound() Response
	AssertMethodNotAllowed() Response
	AssertNotAcceptable() Response
	AssertConflict() Response
	AssertRequestTimeout() Response
	AssertGone() Response
	AssertUnsupportedMediaType() Response
	AssertUnprocessableEntity() Response
	AssertTooManyRequests() Response
	AssertInternalServerError() Response
	AssertServiceUnavailable() Response
	AssertHeader(headerName, value string) Response
	AssertHeaderMissing(string) Response
	AssertCookie(name, value string) Response
	AssertCookieExpired(string) Response
	AssertCookieNotExpired(string) Response
	AssertCookieMissing(string) Response
	AssertSuccessful() Response
	AssertServerError() Response
	AssertDontSee([]string, ...bool) Response
	AssertSee([]string, ...bool) Response
	AssertSeeInOrder([]string, ...bool) Response
	AssertJson(map[string]any) Response
	AssertExactJson(map[string]any) Response
	AssertJsonMissing(map[string]any) Response
	AssertFluentJson(func(json AssertableJSON)) Response
}
