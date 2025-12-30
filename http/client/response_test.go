package client

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation/json"
)

type ResponseTestSuite struct {
	suite.Suite
	mockJson foundation.Json
}

func TestResponseTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseTestSuite))
}

func (s *ResponseTestSuite) SetupSuite() {
	s.mockJson = json.New()
}

func (s *ResponseTestSuite) TestBody() {
	body := `{"message": "hello"}`
	resp := newMockResponse(200, body, nil, s.mockJson)

	content, err := resp.Body()
	s.Require().NoError(err)
	s.Equal(body, content)
}

func (s *ResponseTestSuite) TestJson() {
	body := `{"message": "test"}`
	resp := newMockResponse(200, body, nil, s.mockJson)

	data, err := resp.Json()
	s.Require().NoError(err)
	s.Equal(map[string]any{"message": "test"}, data)
}

func (s *ResponseTestSuite) TestClientError() {
	s.True(newMockResponse(404, "Not Found", nil, s.mockJson).ClientError())
	s.False(newMockResponse(500, "Server Error", nil, s.mockJson).ClientError())
	s.False(newMockResponse(200, "OK", nil, s.mockJson).ClientError())
}

func (s *ResponseTestSuite) TestServerError() {
	s.True(newMockResponse(500, "Internal Server Error", nil, s.mockJson).ServerError())
	s.False(newMockResponse(200, "OK", nil, s.mockJson).ServerError())
	s.False(newMockResponse(404, "Not Found", nil, s.mockJson).ServerError())
}

func (s *ResponseTestSuite) TestFailed() {
	s.True(newMockResponse(404, "", nil, s.mockJson).Failed())
	s.True(newMockResponse(500, "", nil, s.mockJson).Failed())
	s.False(newMockResponse(200, "", nil, s.mockJson).Failed())
}

func (s *ResponseTestSuite) TestRedirect() {
	s.True(newMockResponse(301, "Moved", nil, s.mockJson).Redirect())
	s.True(newMockResponse(302, "Found", nil, s.mockJson).Redirect())
	s.False(newMockResponse(200, "OK", nil, s.mockJson).Redirect())
}

func (s *ResponseTestSuite) TestSuccessful() {
	s.True(newMockResponse(200, "", nil, s.mockJson).Successful())
	s.True(newMockResponse(201, "", nil, s.mockJson).Successful())
	s.False(newMockResponse(404, "", nil, s.mockJson).Successful())
	s.False(newMockResponse(500, "", nil, s.mockJson).Successful())
}

func (s *ResponseTestSuite) TestHeader() {
	headers := map[string]string{"Content-Type": "application/json"}
	resp := newMockResponse(200, "", headers, s.mockJson)
	s.Equal("application/json", resp.Header("Content-Type"))
	s.Empty(resp.Header("X-Missing"))
}

func (s *ResponseTestSuite) TestHeaders() {
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Custom":     "123",
	}
	resp := newMockResponse(200, "", headers, s.mockJson)
	h := resp.Headers()
	s.Equal("application/json", h.Get("Content-Type"))
	s.Equal("123", h.Get("X-Custom"))
}

func (s *ResponseTestSuite) TestCookies() {
	cookies := []*http.Cookie{
		{Name: "session", Value: "xyz123"},
		{Name: "theme", Value: "dark"},
	}
	resp := newMockResponseWithCookies(200, "", nil, cookies, s.mockJson)

	foundCookies := resp.Cookies()
	s.Len(foundCookies, 2)

	var session, theme *http.Cookie
	for _, c := range foundCookies {
		if c.Name == "session" {
			session = c
		}
		if c.Name == "theme" {
			theme = c
		}
	}

	s.Require().NotNil(session, "Cookie 'session' should be found")
	s.Equal("xyz123", session.Value)

	s.Require().NotNil(theme, "Cookie 'theme' should be found")
	s.Equal("dark", theme.Value)
}

func (s *ResponseTestSuite) TestCookie() {
	cookie := &http.Cookie{Name: "session", Value: "xyz123"}
	resp := newMockResponseWithCookies(200, "", nil, []*http.Cookie{cookie}, s.mockJson)
	c := resp.Cookie("session")
	s.NotNil(c, "Cookie 'session' should exist")
	if c != nil {
		s.Equal("xyz123", c.Value)
	}
	s.Nil(resp.Cookie("missing_cookie"))
}

func (s *ResponseTestSuite) TestGetContent_Concurrency() {
	body := `{"message": "cached"}`
	resp := newMockResponse(200, body, nil, s.mockJson)

	var wg sync.WaitGroup
	const routines = 10
	wg.Add(routines)

	results := make([]string, routines)

	for i := 0; i < routines; i++ {
		go func(index int) {
			defer wg.Done()
			content, err := resp.Body()
			s.NoError(err)
			results[index] = content
		}(i)
	}
	wg.Wait()

	for _, res := range results {
		s.Equal(body, res)
	}
}

func (s *ResponseTestSuite) TestStatusCodeMethods() {
	s.True(newMockResponse(200, "", nil, s.mockJson).OK())
	s.True(newMockResponse(201, "", nil, s.mockJson).Created())
	s.True(newMockResponse(202, "", nil, s.mockJson).Accepted())
	s.True(newMockResponse(204, "", nil, s.mockJson).NoContent())
	s.True(newMockResponse(301, "", nil, s.mockJson).MovedPermanently())
	s.True(newMockResponse(302, "", nil, s.mockJson).Found())
	s.True(newMockResponse(400, "", nil, s.mockJson).BadRequest())
	s.True(newMockResponse(401, "", nil, s.mockJson).Unauthorized())
	s.True(newMockResponse(402, "", nil, s.mockJson).PaymentRequired())
	s.True(newMockResponse(403, "", nil, s.mockJson).Forbidden())
	s.True(newMockResponse(404, "", nil, s.mockJson).NotFound())
	s.True(newMockResponse(408, "", nil, s.mockJson).RequestTimeout())
	s.True(newMockResponse(409, "", nil, s.mockJson).Conflict())
	s.True(newMockResponse(422, "", nil, s.mockJson).UnprocessableEntity())
	s.True(newMockResponse(429, "", nil, s.mockJson).TooManyRequests())
}

func newMockResponse(status int, body string, headers map[string]string, json foundation.Json) *Response {
	recorder := httptest.NewRecorder()

	for key, value := range headers {
		recorder.Header().Set(key, value)
	}

	recorder.WriteHeader(status)
	if body != "" {
		_, _ = recorder.WriteString(body)
	}

	return NewResponse(recorder.Result(), json)
}

func newMockResponseWithCookies(status int, body string, headers map[string]string, cookies []*http.Cookie, json foundation.Json) *Response {
	recorder := httptest.NewRecorder()

	for key, value := range headers {
		recorder.Header().Set(key, value)
	}

	for _, cookie := range cookies {
		http.SetCookie(recorder, cookie)
	}

	recorder.WriteHeader(status)
	if body != "" {
		_, _ = recorder.WriteString(body)
	}

	return NewResponse(recorder.Result(), json)
}
