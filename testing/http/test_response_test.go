package http

import (
	"html"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	contractshttp "github.com/goravel/framework/contracts/testing/http"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mockssession "github.com/goravel/framework/mocks/session"
)

func TestAssertOk(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertOk()
}

func TestAssertCreated(t *testing.T) {
	res := createTestResponse(http.StatusCreated)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertCreated()
}

func TestAssertAccepted(t *testing.T) {
	res := createTestResponse(http.StatusAccepted)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertAccepted()
}

func TestAssertNoContent(t *testing.T) {
	res := createTestResponse(http.StatusNoContent)
	res.Body = http.NoBody

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertNoContent()
}

func TestAssertMovedPermanently(t *testing.T) {
	res := createTestResponse(http.StatusMovedPermanently)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertMovedPermanently()
}

func TestAssertFound(t *testing.T) {
	res := createTestResponse(http.StatusFound)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertFound()
}

func TestAssertNotModified(t *testing.T) {
	res := createTestResponse(http.StatusNotModified)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertNotModified()
}

func TestAssertPartialContent(t *testing.T) {
	res := createTestResponse(http.StatusPartialContent)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertPartialContent()
}

func TestAssertTemporaryRedirect(t *testing.T) {
	res := createTestResponse(http.StatusTemporaryRedirect)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertTemporaryRedirect()
}

func TestAssertBadRequest(t *testing.T) {
	res := createTestResponse(http.StatusBadRequest)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertBadRequest()
}

func TestAssertUnauthorized(t *testing.T) {
	res := createTestResponse(http.StatusUnauthorized)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertUnauthorized()
}

func TestAssertPaymentRequired(t *testing.T) {
	res := createTestResponse(http.StatusPaymentRequired)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertPaymentRequired()
}

func TestAssertForbidden(t *testing.T) {
	res := createTestResponse(http.StatusForbidden)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertForbidden()
}

func TestAssertNotFound(t *testing.T) {
	res := createTestResponse(http.StatusNotFound)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertNotFound()
}

func TestAssertMethodNotAllowed(t *testing.T) {
	res := createTestResponse(http.StatusMethodNotAllowed)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertMethodNotAllowed()
}

func TestAssertNotAcceptable(t *testing.T) {
	res := createTestResponse(http.StatusNotAcceptable)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertNotAcceptable()
}

func TestAssertConflict(t *testing.T) {
	res := createTestResponse(http.StatusConflict)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertConflict()
}

func TestAssertRequestTimeout(t *testing.T) {
	res := createTestResponse(http.StatusRequestTimeout)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertRequestTimeout()
}

func TestAssertGone(t *testing.T) {
	res := createTestResponse(http.StatusGone)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertGone()
}

func TestAssertUnsupportedMediaType(t *testing.T) {
	res := createTestResponse(http.StatusUnsupportedMediaType)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertUnsupportedMediaType()
}

func TestAssertUnprocessableEntity(t *testing.T) {
	res := createTestResponse(http.StatusUnprocessableEntity)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertUnprocessableEntity()
}

func TestAssertTooManyRequests(t *testing.T) {
	res := createTestResponse(http.StatusTooManyRequests)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertTooManyRequests()
}

func TestAssertInternalServerError(t *testing.T) {
	res := createTestResponse(http.StatusInternalServerError)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertInternalServerError()
}

func TestAssertServiceUnavailable(t *testing.T) {
	res := createTestResponse(http.StatusServiceUnavailable)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertServiceUnavailable()
}

func TestAssertHeader(t *testing.T) {
	headerName, headerValue := "Content-Type", "application/json"
	res := createTestResponse(http.StatusCreated)
	res.Header.Set(headerName, headerValue)

	r := NewTestResponse(t, res, json.New(), nil)

	r.AssertHeader(headerName, headerValue).AssertCreated()
}

func TestAssertHeaderMissing(t *testing.T) {
	res := createTestResponse(http.StatusCreated)

	r := NewTestResponse(t, res, json.New(), nil)

	r.AssertHeaderMissing("X-Custom-Header").AssertCreated()
}

func TestAssertSuccessful(t *testing.T) {
	res := createTestResponse(http.StatusPartialContent)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertSuccessful()
}

func TestServerError(t *testing.T) {
	res := createTestResponse(http.StatusInternalServerError)
	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertServerError()
}

func TestAssertCookie(t *testing.T) {
	res := createTestResponse(http.StatusCreated)
	res.Header.Add("Set-Cookie", (&http.Cookie{
		Name:     "session_id",
		Value:    "12345",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}).String())

	r := NewTestResponse(t, res, json.New(), nil)

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

	r := NewTestResponse(t, res, json.New(), nil)

	r.AssertCookie("session_id", "expired").
		AssertCookieExpired("session_id")
}

func TestAssertCookieMissing(t *testing.T) {
	res := createTestResponse(http.StatusOK)

	r := NewTestResponse(t, res, json.New(), nil)

	r.AssertCookieMissing("session_id")
}

func TestAssertSee(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	res.Body = io.NopCloser(strings.NewReader("Hello, World! This is a test response."))

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertSee([]string{"Hello", "test"})
}

func TestAssertSeeEscaped(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	escapedContent := html.EscapeString("<div>Hello, World!</div>")
	res.Body = io.NopCloser(strings.NewReader(escapedContent))

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertSee([]string{"<div>Hello, World!</div>"}, true)
}

func TestAssertDontSee(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	res.Body = io.NopCloser(strings.NewReader("This is a safe response."))

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertDontSee([]string{"error", "failure"})
}

func TestAssertDontSeeEscaped(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	res.Body = io.NopCloser(strings.NewReader("<div>Unauthorized access</div>"))

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertDontSee([]string{"<div>Unauthorized access</div>"}, true)
}

func TestAssertSeeInOrder(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	res.Body = io.NopCloser(strings.NewReader("Hello, this is a test for seeing values in order."))

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertSeeInOrder([]string{"Hello", "test", "values"})
}

func TestAssertJson(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	res.Body = io.NopCloser(strings.NewReader(`{"key1": "value1", "key2": 42}`))

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertJson(map[string]any{"key1": "value1"})
}

func TestAssertExactJson(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	res.Body = io.NopCloser(strings.NewReader(`{"key1": "value1", "key2": 42}`))

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertExactJson(map[string]any{"key1": "value1", "key2": float64(42)})
}

func TestAssertJsonMissing(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	res.Body = io.NopCloser(strings.NewReader(`{"key1": "value1", "key2": 42}`))

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertJsonMissing(map[string]any{"key3": "value3"})
}

func TestAssertFluentJson(t *testing.T) {
	sampleJson := `{"name": "krishan", "age": 22, "email": "krishan@example.com"}`
	res := createTestResponse(http.StatusOK)
	res.Body = io.NopCloser(strings.NewReader(sampleJson))

	r := NewTestResponse(t, res, json.New(), nil)

	r.AssertFluentJson(func(json contractshttp.AssertableJSON) {
		json.Has("name").Where("name", "krishan")
		json.Has("age").Where("age", float64(22))
		json.Has("email").Where("email", "krishan@example.com")
	}).AssertFluentJson(func(json contractshttp.AssertableJSON) {
		json.Missing("non_existent_field")
	})
}

func TestAssertSeeInOrderWithEscape(t *testing.T) {
	res := createTestResponse(http.StatusOK)
	escapedContent := html.EscapeString("Hello, <div>ordered</div> values")
	res.Body = io.NopCloser(strings.NewReader(escapedContent))

	r := NewTestResponse(t, res, json.New(), nil)
	r.AssertSeeInOrder([]string{"Hello,", "<div>ordered</div>"}, true)
}

func TestSession_Success(t *testing.T) {
	mockSessionManager := mockssession.NewManager(t)
	mockDriver := mockssession.NewDriver(t)
	mockSession := mockssession.NewSession(t)

	sessionData := map[string]any{
		"user_id":   123,
		"user_role": "admin",
	}

	mockSessionManager.On("Driver").Return(mockDriver, nil).Once()
	mockSessionManager.On("BuildSession", mockDriver).Return(mockSession, nil).Once()
	mockSessionManager.On("ReleaseSession", mockSession).Once()
	mockSession.On("All").Return(sessionData).Once()

	cookie := &http.Cookie{
		Name:  "session_id",
		Value: "test_session_id",
	}
	response := createTestResponse(http.StatusOK)
	response.Header.Add("Set-Cookie", cookie.String())

	testResponse := &TestResponseImpl{
		session: mockSessionManager,
	}
	session, err := testResponse.Session()

	require.NoError(t, err)
	require.Equal(t, sessionData, session)
}

func TestSession_DriverError(t *testing.T) {
	mockSessionManager := mockssession.NewManager(t)

	mockSessionManager.On("Driver").Return(nil, errors.New("driver error")).Once()

	testResponse := &TestResponseImpl{
		session: mockSessionManager,
	}
	_, err := testResponse.Session()

	require.EqualError(t, err, "driver error")
}

func TestSession_BuildSessionError(t *testing.T) {
	mockSessionManager := mockssession.NewManager(t)
	mockDriver := mockssession.NewDriver(t)

	mockSessionManager.On("Driver").Return(mockDriver, nil).Once()
	mockSessionManager.On("BuildSession", mockDriver).Return(nil, errors.New("build session error")).Once()

	testResponse := &TestResponseImpl{
		session: mockSessionManager,
	}
	_, err := testResponse.Session()

	require.EqualError(t, err, "build session error")
}

func TestSession_SessionFacadeNotSet(t *testing.T) {
	testResponse := &TestResponseImpl{
		session: nil,
	}
	_, err := testResponse.Session()

	require.ErrorIs(t, err, errors.SessionFacadeNotSet)
}

func createTestResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{},
	}
}
