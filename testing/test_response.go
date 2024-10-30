package testing

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/support/carbon"
)

type TestResponseImpl struct {
	t        *testing.T
	mu       sync.RWMutex
	response *http.Response
	content  string
}

func NewTestResponse(t *testing.T, resp *http.Response) contractstesting.TestResponse {
	return &TestResponseImpl{t: t, response: resp}
}

func (r *TestResponseImpl) AssertStatus(status int) contractstesting.TestResponse {
	actual := r.getStatusCode()
	assert.Equal(r.T(), status, actual, fmt.Sprintf("Expected response status code [%d] but received %d.", status, actual))
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

	body, err := r.getContent()
	assert.NoError(r.T(), err)
	assert.Empty(r.T(), body)

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
	got := r.response.Header.Get(headerName)
	assert.NotEmpty(r.T(), got, fmt.Sprintf("Header [%s] not present on response.", headerName))
	if got != "" {
		assert.Equal(r.T(), value, got, fmt.Sprintf("Header [%s] was found, but value [%s] does not match [%s].", headerName, got, value))
	}
	return r
}

func (r *TestResponseImpl) AssertHeaderMissing(headerName string) contractstesting.TestResponse {
	got := r.response.Header.Get(headerName)
	assert.Empty(r.T(), got, fmt.Sprintf("Unexpected header [%s] is present on response.", headerName))
	return r
}

func (r *TestResponseImpl) AssertCookie(name, value string) contractstesting.TestResponse {
	cookie := r.getCookie(name)
	assert.NotNil(r.T(), cookie, fmt.Sprintf("Cookie [%s] not present on response.", name))

	if cookie == nil {
		return r
	}

	assert.Equal(r.T(), value, cookie.Value, fmt.Sprintf("Cookie [%s] was found, but value [%s] does not match [%s]", name, cookie.Value, value))

	return r
}

func (r *TestResponseImpl) AssertCookieExpired(name string) contractstesting.TestResponse {
	cookie := r.getCookie(name)
	assert.NotNil(r.T(), cookie, fmt.Sprintf("Cookie [%s] not present on response.", name))

	if cookie == nil {
		return r
	}

	expirationTime := carbon.FromStdTime(cookie.Expires)
	if expirationTime.IsZero() && cookie.MaxAge > 0 {
		expirationTime = carbon.FromStdTime(time.Unix(int64(cookie.MaxAge), 0))
	}

	assert.True(r.T(), !expirationTime.IsZero() && expirationTime.Lt(carbon.Now()), fmt.Sprintf("Cookie [%s] is not expired; it expires at [%s].", name, expirationTime.ToString()))
	return r
}

func (r *TestResponseImpl) AssertCookieNotExpired(name string) contractstesting.TestResponse {
	cookie := r.getCookie(name)
	assert.NotNil(r.T(), cookie, fmt.Sprintf("Cookie [%s] not present on response.", name))

	if cookie == nil {
		return r
	}

	expirationTime := carbon.FromStdTime(cookie.Expires)
	if expirationTime.IsZero() && cookie.MaxAge > 0 {
		expirationTime = carbon.FromStdTime(time.Unix(int64(cookie.MaxAge), 0))
	}

	assert.True(r.T(), expirationTime.IsZero() || expirationTime.Gt(carbon.Now()), fmt.Sprintf("Cookie [%s] is expired; it expired at [%s].", name, expirationTime))
	return r
}

func (r *TestResponseImpl) AssertCookieMissing(name string) contractstesting.TestResponse {
	assert.Nil(r.T(), r.getCookie(name), fmt.Sprintf("Cookie [%s] is present on response.", name))

	return r
}

func (r *TestResponseImpl) T() *testing.T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.t
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

	defer r.response.Body.Close()
	body, err := io.ReadAll(r.response.Body)
	if err != nil {
		return "", err
	}

	r.content = string(body)
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
