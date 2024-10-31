package testing

import (
	"net/http"
	"testing"
	"time"
)

func TestAssertOk(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	r := NewTestResponse(t, res)
	r.AssertOk()
}

func TestAssertCreated(t *testing.T) {
	res := createTestResponse(http.StatusCreated)
	r := NewTestResponse(t, res)
	r.AssertCreated()
}

func TestAssertAccepted(t *testing.T) {
	res := createTestResponse(http.StatusAccepted)
	r := NewTestResponse(t, res)
	r.AssertAccepted()
}

func TestAssertNoContent(t *testing.T) {
	res := createTestResponse(http.StatusNoContent)
	r := NewTestResponse(t, res)
	r.AssertNoContent()
}

func TestAssertMovedPermanently(t *testing.T) {
	res := createTestResponse(http.StatusMovedPermanently)
	r := NewTestResponse(t, res)
	r.AssertMovedPermanently()
}

func TestAssertFound(t *testing.T) {
	res := createTestResponse(http.StatusFound)
	r := NewTestResponse(t, res)
	r.AssertFound()
}

func TestAssertNotModified(t *testing.T) {
	res := createTestResponse(http.StatusNotModified)
	r := NewTestResponse(t, res)
	r.AssertNotModified()
}

func TestAssertPartialContent(t *testing.T) {
	res := createTestResponse(http.StatusPartialContent)
	r := NewTestResponse(t, res)
	r.AssertPartialContent()
}

func TestAssertTemporaryRedirect(t *testing.T) {
	res := createTestResponse(http.StatusTemporaryRedirect)
	r := NewTestResponse(t, res)
	r.AssertTemporaryRedirect()
}

func TestAssertBadRequest(t *testing.T) {
	res := createTestResponse(http.StatusBadRequest)
	r := NewTestResponse(t, res)
	r.AssertBadRequest()
}

func TestAssertUnauthorized(t *testing.T) {
	res := createTestResponse(http.StatusUnauthorized)
	r := NewTestResponse(t, res)
	r.AssertUnauthorized()
}

func TestAssertPaymentRequired(t *testing.T) {
	res := createTestResponse(http.StatusPaymentRequired)
	r := NewTestResponse(t, res)
	r.AssertPaymentRequired()
}

func TestAssertForbidden(t *testing.T) {
	res := createTestResponse(http.StatusForbidden)
	r := NewTestResponse(t, res)
	r.AssertForbidden()
}

func TestAssertNotFound(t *testing.T) {
	res := createTestResponse(http.StatusNotFound)
	r := NewTestResponse(t, res)
	r.AssertNotFound()
}

func TestAssertMethodNotAllowed(t *testing.T) {
	res := createTestResponse(http.StatusMethodNotAllowed)
	r := NewTestResponse(t, res)
	r.AssertMethodNotAllowed()
}

func TestAssertNotAcceptable(t *testing.T) {
	res := createTestResponse(http.StatusNotAcceptable)
	r := NewTestResponse(t, res)
	r.AssertNotAcceptable()
}

func TestAssertConflict(t *testing.T) {
	res := createTestResponse(http.StatusConflict)
	r := NewTestResponse(t, res)
	r.AssertConflict()
}

func TestAssertRequestTimeout(t *testing.T) {
	res := createTestResponse(http.StatusRequestTimeout)
	r := NewTestResponse(t, res)
	r.AssertRequestTimeout()
}

func TestAssertGone(t *testing.T) {
	res := createTestResponse(http.StatusGone)
	r := NewTestResponse(t, res)
	r.AssertGone()
}

func TestAssertUnsupportedMediaType(t *testing.T) {
	res := createTestResponse(http.StatusUnsupportedMediaType)
	r := NewTestResponse(t, res)
	r.AssertUnsupportedMediaType()
}

func TestAssertUnprocessableEntity(t *testing.T) {
	res := createTestResponse(http.StatusUnprocessableEntity)
	r := NewTestResponse(t, res)
	r.AssertUnprocessableEntity()
}

func TestAssertTooManyRequests(t *testing.T) {
	res := createTestResponse(http.StatusTooManyRequests)
	r := NewTestResponse(t, res)
	r.AssertTooManyRequests()
}

func TestAssertInternalServerError(t *testing.T) {
	res := createTestResponse(http.StatusInternalServerError)
	r := NewTestResponse(t, res)
	r.AssertInternalServerError()
}

func TestAssertHeader(t *testing.T) {
	headerName, headerValue := "Content-Type", "application/json"
	res := createTestResponse(http.StatusCreated)
	res.Header.Set(headerName, headerValue)

	r := NewTestResponse(t, res)

	r.AssertHeader(headerName, headerValue).AssertCreated()
}

func TestAssertHeaderMissing(t *testing.T) {
	res := createTestResponse(http.StatusCreated)

	r := NewTestResponse(t, res)

	r.AssertHeaderMissing("X-Custom-Header").AssertCreated()
}

func TestAssertCookie(t *testing.T) {
	res := createTestResponse(http.StatusCreated)
	res.Header.Add("Set-Cookie", (&http.Cookie{
		Name:     "session_id",
		Value:    "12345",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}).String())

	r := NewTestResponse(t, res)

	r.AssertCookie("session_id", "12345").
		AssertCookieNotExpired("session_id").
		AssertCreated()
}

func TestAssertCookieExpired(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	res.Header.Add("Set-Cookie", (&http.Cookie{
		Name:    "session_id",
		Value:   "expired",
		Expires: time.Now().Add(-24 * time.Hour),
	}).String())

	r := NewTestResponse(t, res)

	r.AssertCookie("session_id", "expired").
		AssertCookieExpired("session_id")
}

func TestAssertCookieMissing(t *testing.T) {
	res := createTestResponse(http.StatusOK)

	r := NewTestResponse(t, res)

	r.AssertCookieMissing("session_id")
}

func createTestResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{},
	}
}
