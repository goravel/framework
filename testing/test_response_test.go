package testing

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAssertOk(t *testing.T) {
	recorder := createTestResponse(http.StatusOK)
	response := NewTestResponse(t, recorder)
	response.AssertOk()
}

func TestAssertCreated(t *testing.T) {
	recorder := createTestResponse(http.StatusCreated)
	response := NewTestResponse(t, recorder)
	response.AssertCreated()
}

func TestAssertAccepted(t *testing.T) {
	recorder := createTestResponse(http.StatusAccepted)
	response := NewTestResponse(t, recorder)
	response.AssertAccepted()
}

func TestAssertNoContent(t *testing.T) {
	recorder := createTestResponse(http.StatusNoContent)
	response := NewTestResponse(t, recorder)
	response.AssertNoContent()
}

func TestAssertMovedPermanently(t *testing.T) {
	recorder := createTestResponse(http.StatusMovedPermanently)
	response := NewTestResponse(t, recorder)
	response.AssertMovedPermanently()
}

func TestAssertFound(t *testing.T) {
	recorder := createTestResponse(http.StatusFound)
	response := NewTestResponse(t, recorder)
	response.AssertFound()
}

func TestAssertNotModified(t *testing.T) {
	recorder := createTestResponse(http.StatusNotModified)
	response := NewTestResponse(t, recorder)
	response.AssertNotModified()
}

func TestAssertPartialContent(t *testing.T) {
	recorder := createTestResponse(http.StatusPartialContent)
	response := NewTestResponse(t, recorder)
	response.AssertPartialContent()
}

func TestAssertTemporaryRedirect(t *testing.T) {
	recorder := createTestResponse(http.StatusTemporaryRedirect)
	response := NewTestResponse(t, recorder)
	response.AssertTemporaryRedirect()
}

func TestAssertBadRequest(t *testing.T) {
	recorder := createTestResponse(http.StatusBadRequest)
	response := NewTestResponse(t, recorder)
	response.AssertBadRequest()
}

func TestAssertUnauthorized(t *testing.T) {
	recorder := createTestResponse(http.StatusUnauthorized)
	response := NewTestResponse(t, recorder)
	response.AssertUnauthorized()
}

func TestAssertPaymentRequired(t *testing.T) {
	recorder := createTestResponse(http.StatusPaymentRequired)
	response := NewTestResponse(t, recorder)
	response.AssertPaymentRequired()
}

func TestAssertForbidden(t *testing.T) {
	recorder := createTestResponse(http.StatusForbidden)
	response := NewTestResponse(t, recorder)
	response.AssertForbidden()
}

func TestAssertNotFound(t *testing.T) {
	recorder := createTestResponse(http.StatusNotFound)
	response := NewTestResponse(t, recorder)
	response.AssertNotFound()
}

func TestAssertMethodNotAllowed(t *testing.T) {
	recorder := createTestResponse(http.StatusMethodNotAllowed)
	response := NewTestResponse(t, recorder)
	response.AssertMethodNotAllowed()
}

func TestAssertNotAcceptable(t *testing.T) {
	recorder := createTestResponse(http.StatusNotAcceptable)
	response := NewTestResponse(t, recorder)
	response.AssertNotAcceptable()
}

func TestAssertConflict(t *testing.T) {
	recorder := createTestResponse(http.StatusConflict)
	response := NewTestResponse(t, recorder)
	response.AssertConflict()
}

func TestAssertRequestTimeout(t *testing.T) {
	recorder := createTestResponse(http.StatusRequestTimeout)
	response := NewTestResponse(t, recorder)
	response.AssertRequestTimeout()
}

func TestAssertGone(t *testing.T) {
	recorder := createTestResponse(http.StatusGone)
	response := NewTestResponse(t, recorder)
	response.AssertGone()
}

func TestAssertUnsupportedMediaType(t *testing.T) {
	recorder := createTestResponse(http.StatusUnsupportedMediaType)
	response := NewTestResponse(t, recorder)
	response.AssertUnsupportedMediaType()
}

func TestAssertUnprocessableEntity(t *testing.T) {
	recorder := createTestResponse(http.StatusUnprocessableEntity)
	response := NewTestResponse(t, recorder)
	response.AssertUnprocessableEntity()
}

func TestAssertTooManyRequests(t *testing.T) {
	recorder := createTestResponse(http.StatusTooManyRequests)
	response := NewTestResponse(t, recorder)
	response.AssertTooManyRequests()
}

func TestAssertInternalServerError(t *testing.T) {
	recorder := createTestResponse(http.StatusInternalServerError)
	response := NewTestResponse(t, recorder)
	response.AssertInternalServerError()
}

func TestAssertHeader(t *testing.T) {
	headerName, headerValue := "Content-Type", "application/json"
	recorder := createTestResponse(http.StatusCreated)
	recorder.Result().Header.Set(headerName, headerValue)

	response := NewTestResponse(t, recorder)

	response.AssertHeader(headerName, headerValue).AssertCreated()
}

func TestAssertHeaderMissing(t *testing.T) {
	recorder := createTestResponse(http.StatusCreated)

	response := NewTestResponse(t, recorder)

	response.AssertHeaderMissing("X-Custom-Header").AssertCreated()
}

func TestAssertCookie(t *testing.T) {
	recorder := createTestResponse(http.StatusCreated)
	recorder.Result().Header.Add("Set-Cookie", (&http.Cookie{
		Name:     "session_id",
		Value:    "12345",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}).String())

	response := NewTestResponse(t, recorder)

	response.AssertCookie("session_id", "12345").
		AssertCookieNotExpired("session_id").
		AssertCreated()
}

func TestAssertCookieExpired(t *testing.T) {
	recorder := createTestResponse(http.StatusOK)
	recorder.Result().Header.Add("Set-Cookie", (&http.Cookie{
		Name:    "session_id",
		Value:   "expired",
		Expires: time.Now().Add(-24 * time.Hour),
	}).String())

	response := NewTestResponse(t, recorder)

	response.AssertCookie("session_id", "expired").
		AssertCookieExpired("session_id")
}

func TestAssertCookieMissing(t *testing.T) {
	recorder := createTestResponse(http.StatusOK)

	response := NewTestResponse(t, recorder)

	response.AssertCookieMissing("session_id")
}

func createTestResponse(statusCode int) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(statusCode)
	return recorder
}
