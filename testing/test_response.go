package testing

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

type TestResponseImpl struct {
	t                 *testing.T
	mu                sync.Mutex
	response          *http.Response
	content           string
	sessionAttributes map[string]any
}

func NewTestResponse(t *testing.T, response *http.Response) contractstesting.TestResponse {
	return &TestResponseImpl{t: t, response: response}
}

func (r *TestResponseImpl) Json() (map[string]any, error) {
	content, err := r.getContent()
	if err != nil {
		return nil, err
	}

	testAble, err := NewAssertableJSON(r.t, content)
	if err != nil {
		return nil, err
	}

	return testAble.Json(), nil
}

func (r *TestResponseImpl) Headers() http.Header {
	return r.response.Header
}

func (r *TestResponseImpl) Cookies() []*http.Cookie {
	return r.response.Cookies()
}

func (r *TestResponseImpl) Cookie(name string) *http.Cookie {
	return r.getCookie(name)
}

func (r *TestResponseImpl) Session() (map[string]any, error) {
	if r.sessionAttributes != nil {
		return r.sessionAttributes, nil
	}

	if sessionFacade == nil {
		return nil, errors.SessionFacadeNotSet
	}

	// Retrieve session driver
	driver, err := sessionFacade.Driver()
	if err != nil {
		return nil, err
	}

	// Build session
	session, err := sessionFacade.BuildSession(driver)
	if err != nil {
		return nil, err
	}

	r.sessionAttributes = session.All()
	sessionFacade.ReleaseSession(session)

	return r.sessionAttributes, nil
}

func (r *TestResponseImpl) IsSuccessful() bool {
	statusCode := r.getStatusCode()
	return statusCode >= 200 && statusCode < 300
}

func (r *TestResponseImpl) IsServerError() bool {
	statusCode := r.getStatusCode()
	return statusCode >= 500 && statusCode < 600
}

func (r *TestResponseImpl) Content() (string, error) {
	return r.getContent()
}

func (r *TestResponseImpl) AssertStatus(status int) contractstesting.TestResponse {
	actual := r.getStatusCode()
	assert.Equal(r.t, status, actual, fmt.Sprintf("Expected response status code [%d] but received %d.", status, actual))
	return r
}

func (r *TestResponseImpl) AssertOk() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusOK)
}

func (r *TestResponseImpl) AssertCreated() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusCreated)
}

func (r *TestResponseImpl) AssertAccepted() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusAccepted)
}

func (r *TestResponseImpl) AssertNoContent(status ...int) contractstesting.TestResponse {
	expectedStatus := http.StatusNoContent
	if len(status) > 0 {
		expectedStatus = status[0]
	}

	r.AssertStatus(expectedStatus)

	content, err := r.getContent()
	assert.Nil(r.t, err)
	assert.Empty(r.t, content)

	return r
}

func (r *TestResponseImpl) AssertMovedPermanently() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusMovedPermanently)
}

func (r *TestResponseImpl) AssertFound() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusFound)
}

func (r *TestResponseImpl) AssertNotModified() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusNotModified)
}

func (r *TestResponseImpl) AssertPartialContent() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusPartialContent)
}

func (r *TestResponseImpl) AssertTemporaryRedirect() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusTemporaryRedirect)
}

func (r *TestResponseImpl) AssertBadRequest() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusBadRequest)
}

func (r *TestResponseImpl) AssertUnauthorized() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusUnauthorized)
}

func (r *TestResponseImpl) AssertPaymentRequired() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusPaymentRequired)
}

func (r *TestResponseImpl) AssertForbidden() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusForbidden)
}

func (r *TestResponseImpl) AssertNotFound() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusNotFound)
}

func (r *TestResponseImpl) AssertMethodNotAllowed() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusMethodNotAllowed)
}

func (r *TestResponseImpl) AssertNotAcceptable() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusNotAcceptable)
}

func (r *TestResponseImpl) AssertConflict() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusConflict)
}

func (r *TestResponseImpl) AssertRequestTimeout() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusRequestTimeout)
}

func (r *TestResponseImpl) AssertGone() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusGone)
}

func (r *TestResponseImpl) AssertUnsupportedMediaType() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusUnsupportedMediaType)
}

func (r *TestResponseImpl) AssertUnprocessableEntity() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusUnprocessableEntity)
}

func (r *TestResponseImpl) AssertTooManyRequests() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusTooManyRequests)
}

func (r *TestResponseImpl) AssertInternalServerError() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusInternalServerError)
}

func (r *TestResponseImpl) AssertServiceUnavailable() contractstesting.TestResponse {
	return r.AssertStatus(http.StatusServiceUnavailable)
}

func (r *TestResponseImpl) AssertHeader(headerName, value string) contractstesting.TestResponse {
	got := r.getHeader(headerName)
	assert.NotEmpty(r.t, got, fmt.Sprintf("Header [%s] not present on response.", headerName))
	if got != "" {
		assert.Equal(r.t, value, got, fmt.Sprintf("Header [%s] was found, but value [%s] does not match [%s].", headerName, got, value))
	}
	return r
}

func (r *TestResponseImpl) AssertHeaderMissing(headerName string) contractstesting.TestResponse {
	got := r.getHeader(headerName)
	assert.Empty(r.t, got, fmt.Sprintf("Unexpected header [%s] is present on response.", headerName))
	return r
}

func (r *TestResponseImpl) AssertCookie(name, value string) contractstesting.TestResponse {
	cookie := r.getCookie(name)
	assert.NotNil(r.t, cookie, fmt.Sprintf("Cookie [%s] not present on response.", name))

	if cookie == nil {
		return r
	}

	assert.Equal(r.t, value, cookie.Value, fmt.Sprintf("Cookie [%s] was found, but value [%s] does not match [%s]", name, cookie.Value, value))

	return r
}

func (r *TestResponseImpl) AssertCookieExpired(name string) contractstesting.TestResponse {
	cookie := r.getCookie(name)
	assert.NotNil(r.t, cookie, fmt.Sprintf("Cookie [%s] not present on response.", name))

	if cookie == nil {
		return r
	}

	expirationTime := carbon.FromStdTime(cookie.Expires)
	assert.True(r.t, r.isCookieExpired(cookie), fmt.Sprintf("Cookie [%s] is not expired; it expires at [%s].", name, expirationTime.ToString()))

	return r
}

func (r *TestResponseImpl) AssertCookieNotExpired(name string) contractstesting.TestResponse {
	cookie := r.getCookie(name)
	assert.NotNil(r.t, cookie, fmt.Sprintf("Cookie [%s] not present on response.", name))

	if cookie == nil {
		return r
	}

	expirationTime := carbon.FromStdTime(cookie.Expires)
	assert.True(r.t, !r.isCookieExpired(cookie), fmt.Sprintf("Cookie [%s] is expired; it expired at [%s].", name, expirationTime))
	return r
}

func (r *TestResponseImpl) AssertCookieMissing(name string) contractstesting.TestResponse {
	assert.Nil(r.t, r.getCookie(name), fmt.Sprintf("Cookie [%s] is present on response.", name))

	return r
}

func (r *TestResponseImpl) AssertSuccessful() contractstesting.TestResponse {
	assert.True(r.t, r.IsSuccessful(), fmt.Sprintf("Expected response status code >=200, <300 but received %d.", r.getStatusCode()))

	return r
}

func (r *TestResponseImpl) AssertServerError() contractstesting.TestResponse {
	assert.True(r.t, r.IsServerError(), fmt.Sprintf("Expected response status code >=500, <600 but received %d.", r.getStatusCode()))

	return r
}

func (r *TestResponseImpl) AssertDontSee(value []string, escaped ...bool) contractstesting.TestResponse {
	content, err := r.getContent()
	assert.Nil(r.t, err)

	shouldEscape := true
	if len(escaped) > 0 {
		shouldEscape = escaped[0]
	}

	for _, v := range value {
		checkValue := v
		if shouldEscape {
			checkValue = html.EscapeString(v)
		}

		assert.NotContains(r.t, content, checkValue, fmt.Sprintf("Response should not contain '%s', but it was found.", checkValue))
	}

	return r
}

func (r *TestResponseImpl) AssertSee(value []string, escaped ...bool) contractstesting.TestResponse {
	content, err := r.getContent()
	assert.Nil(r.t, err)

	shouldEscape := true
	if len(escaped) > 0 {
		shouldEscape = escaped[0]
	}

	for _, v := range value {
		checkValue := v
		if shouldEscape {
			checkValue = html.EscapeString(v)
		}

		assert.Contains(r.t, content, checkValue, fmt.Sprintf("Expected to see '%s' in response, but it was not found.", checkValue))
	}

	return r
}

func (r *TestResponseImpl) AssertSeeInOrder(value []string, escaped ...bool) contractstesting.TestResponse {
	content, err := r.getContent()
	assert.Nil(r.t, err)

	shouldEscape := true
	if len(escaped) > 0 {
		shouldEscape = escaped[0]
	}

	previousIndex := -1
	for _, v := range value {
		checkValue := v
		if shouldEscape {
			checkValue = html.EscapeString(v)
		}

		currentIndex := strings.Index(content[previousIndex+1:], checkValue)
		assert.GreaterOrEqual(r.t, currentIndex, 0, fmt.Sprintf("Expected to see '%s' in response in the correct order, but it was not found.", checkValue))
		previousIndex += currentIndex + len(checkValue)
	}

	return r
}

func (r *TestResponseImpl) AssertJson(data map[string]any) contractstesting.TestResponse {
	content, err := r.getContent()
	assert.Nil(r.t, err)

	assertableJson, err := NewAssertableJSON(r.t, content)
	assert.Nil(r.t, err)

	for key, value := range data {
		assertableJson.Where(key, value)
	}

	return r
}

func (r *TestResponseImpl) AssertExactJson(data map[string]any) contractstesting.TestResponse {
	actual, err := r.Json()
	assert.Nil(r.t, err)
	assert.Equal(r.t, data, actual, "The JSON response does not match exactly with the expected content")
	return r
}

func (r *TestResponseImpl) AssertJsonMissing(data map[string]any) contractstesting.TestResponse {
	actual, err := r.Json()
	assert.Nil(r.t, err)

	for key, expectedValue := range data {
		actualValue, found := actual[key]
		if found {
			assert.NotEqual(r.t, expectedValue, actualValue, "Found unexpected key-value pair in JSON response: key '%s' with value '%v'", key, actualValue)
		}
	}
	return r
}

func (r *TestResponseImpl) AssertFluentJson(callback func(json contractstesting.AssertableJSON)) contractstesting.TestResponse {
	content, err := r.getContent()
	assert.Nil(r.t, err)

	assertableJson, err := NewAssertableJSON(r.t, content)
	assert.Nil(r.t, err)

	callback(assertableJson)

	return r
}

func (r *TestResponseImpl) getStatusCode() int {
	return r.response.StatusCode
}

func (r *TestResponseImpl) getContent() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.content != "" {
		return r.content, nil
	}

	defer func() {
		_ = r.response.Body.Close()
	}()

	content, err := io.ReadAll(r.response.Body)
	if err != nil {
		return "", err
	}

	r.content = string(content)
	return r.content, nil
}

func (r *TestResponseImpl) getCookie(name string) *http.Cookie {
	for _, c := range r.response.Cookies() {
		if c.Name == name {
			return c
		}
	}

	return nil
}

func (r *TestResponseImpl) getHeader(name string) string {
	return r.response.Header.Get(name)
}

func (r *TestResponseImpl) isCookieExpired(cookie *http.Cookie) bool {
	if cookie.MaxAge > 0 {
		return false
	}

	if cookie.MaxAge < 0 {
		return true
	}

	// MaxAge == 0 means no Max-Age specified; check Expires attribute
	if cookie.Expires.IsZero() {
		// Session cookie; consider not expired until the session ends
		return false
	}

	return cookie.Expires.Before(time.Now())
}
