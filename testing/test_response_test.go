package testing

import (
	"net/http"
	"testing"
	"time"
)

func TestAssertStatus(t *testing.T) {
	resp := createTestResponse(http.StatusOK)
	r := NewTestResponse(t, resp)

	r.AssertStatus(http.StatusOK)
}

func TestAssertHeader(t *testing.T) {
	headerName, headerValue := "Content-Type", "application/json"
	resp := createTestResponse(http.StatusCreated)
	resp.Header.Set(headerName, headerValue)

	r := NewTestResponse(t, resp)

	r.AssertHeader(headerName, headerValue).
		AssertHeaderMissing("X-Custom-Header").
		AssertCreated()
}

func TestAssertCookie(t *testing.T) {
	resp := createTestResponse(http.StatusCreated)
	resp.Header.Add("Set-Cookie", (&http.Cookie{
		Name:     "session_id",
		Value:    "12345",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}).String())

	r := NewTestResponse(t, resp)

	r.AssertCookie("session_id", "12345").
		AssertCookieNotExpired("session_id").
		AssertCreated()
}

func TestAssertCookieExpired(t *testing.T) {
	resp := createTestResponse(http.StatusOK)
	resp.Header.Add("Set-Cookie", (&http.Cookie{
		Name:    "session_id",
		Value:   "expired",
		Expires: time.Now().Add(-24 * time.Hour),
	}).String())

	r := NewTestResponse(t, resp)

	r.AssertCookie("session_id", "expired").
		AssertCookieExpired("session_id")
}

func TestAssertCookieMissing(t *testing.T) {
	resp := createTestResponse(http.StatusOK)

	r := NewTestResponse(t, resp)

	r.AssertCookieMissing("session_id")
}

func createTestResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{},
	}
}
