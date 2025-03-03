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
	s.mockJson = json.NewJson()
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
	resp := newMockResponse(404, "Not Found", nil, s.mockJson)
	s.True(resp.ClientError())

	resp = newMockResponse(500, "Server Error", nil, s.mockJson)
	s.False(resp.ClientError())
}

func (s *ResponseTestSuite) TestServerError() {
	resp := newMockResponse(500, "Internal Server Error", nil, s.mockJson)
	s.True(resp.ServerError())

	resp = newMockResponse(200, "OK", nil, s.mockJson)
	s.False(resp.ServerError())
}

func (s *ResponseTestSuite) TestFailed() {
	s.True(newMockResponse(404, "", nil, s.mockJson).Failed())
	s.True(newMockResponse(500, "", nil, s.mockJson).Failed())
	s.False(newMockResponse(200, "", nil, s.mockJson).Failed())
}

func (s *ResponseTestSuite) TestRedirect() {
	resp := newMockResponse(302, "Found", nil, s.mockJson)
	s.True(resp.Redirect())

	resp = newMockResponse(200, "OK", nil, s.mockJson)
	s.False(resp.Redirect())
}

func (s *ResponseTestSuite) TestSuccessful() {
	s.True(newMockResponse(200, "", nil, s.mockJson).Successful())
	s.True(newMockResponse(201, "", nil, s.mockJson).Successful())
	s.False(newMockResponse(404, "", nil, s.mockJson).Successful())
}

func (s *ResponseTestSuite) TestHeader() {
	headers := map[string]string{"Content-Type": "application/json"}
	resp := newMockResponse(200, "", headers, s.mockJson)
	s.Equal("application/json", resp.Header("Content-Type"))
}

func (s *ResponseTestSuite) TestHeaders() {
	headers := map[string]string{"Content-Type": "application/json"}
	resp := newMockResponse(200, "", headers, s.mockJson)
	s.Equal("application/json", resp.Headers().Get("Content-Type"))
}

func (s *ResponseTestSuite) TestCookies() {
	cookie := &http.Cookie{Name: "session", Value: "xyz123"}
	resp := newMockResponseWithCookies(200, "", nil, []*http.Cookie{cookie}, s.mockJson)
	s.Len(resp.Cookies(), 1)
	s.Equal("session", resp.Cookies()[0].Name)
}

func (s *ResponseTestSuite) TestCookie() {
	cookie := &http.Cookie{Name: "session", Value: "xyz123"}
	resp := newMockResponseWithCookies(200, "", nil, []*http.Cookie{cookie}, s.mockJson)
	s.Equal("xyz123", resp.Cookie("session").Value)
}

func (s *ResponseTestSuite) TestGetContent() {
	body := `{"message": "cached"}`
	resp := newMockResponse(200, body, nil, s.mockJson)

	var wg sync.WaitGroup
	wg.Add(2)

	var content1, content2 string
	go func() {
		defer wg.Done()
		content1, _ = resp.Body()
	}()
	go func() {
		defer wg.Done()
		content2, _ = resp.Body()
	}()
	wg.Wait()

	s.Equal(body, content1)
	s.Equal(body, content2)
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
	resp := httptest.NewRecorder()
	for key, value := range headers {
		resp.Header().Set(key, value)
	}
	resp.WriteHeader(status)
	if body != "" {
		resp.Body.WriteString(body)
	}

	return NewResponse(resp.Result(), json)
}

func newMockResponseWithCookies(status int, body string, headers map[string]string, cookies []*http.Cookie, json foundation.Json) *Response {
	resp := newMockResponse(status, body, headers, json)
	for _, cookie := range cookies {
		resp.response.Header.Add("Set-Cookie", cookie.String())
	}
	return resp
}
