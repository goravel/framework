package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/foundation/json"
)

type RequestTestSuite struct {
	suite.Suite
	request    *Request
	httpClient *http.Client
	json       foundation.Json
}

func TestRequestTestSuite(t *testing.T) {
	suite.Run(t, new(RequestTestSuite))
}

func (s *RequestTestSuite) SetupTest() {
	s.json = json.New()
	s.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	s.request = NewRequest(s.httpClient, s.json, "https://api.goravel.dev", "")
}

func (s *RequestTestSuite) TestClone() {
	req := s.request.Clone().
		BaseUrl("https://original.com").
		WithQueryParameter("key", "value")

	cookie := &http.Cookie{Name: "session", Value: "abc123"}
	req = req.WithCookie(cookie)

	originalConcrete := req.(*Request)
	clonedReq := req.Clone().(*Request)

	s.Equal("https://original.com", clonedReq.baseUrl)

	clonedReqWithNewBase := clonedReq.BaseUrl("https://modified.com").(*Request)
	s.Equal("https://original.com", originalConcrete.baseUrl)
	s.Equal("https://modified.com", clonedReqWithNewBase.baseUrl)

	s.Equal(originalConcrete.queryParams.Encode(), clonedReq.queryParams.Encode())

	clonedReq = clonedReq.WithQueryParameter("newKey", "newValue").(*Request)
	s.NotEqual(originalConcrete.queryParams.Encode(), clonedReq.queryParams.Encode())

	s.Equal(len(originalConcrete.cookies), len(clonedReq.cookies))
	s.Equal(originalConcrete.cookies[0].Value, clonedReq.cookies[0].Value)

	clonedReq = clonedReq.WithCookie(&http.Cookie{Name: "session", Value: "modified"}).(*Request)

	s.Equal("abc123", originalConcrete.cookies[0].Value)
	s.Equal("modified", clonedReq.cookies[len(clonedReq.cookies)-1].Value)

	req = req.WithQueryParameter("param1", "value1")
	originalConcrete = req.(*Request)
	clonedReq = req.Clone().(*Request)

	s.Equal(originalConcrete.queryParams.Get("param1"), clonedReq.queryParams.Get("param1"))

	clonedReq = clonedReq.WithQueryParameter("param1", "changedValue").(*Request)
	s.NotEqual(originalConcrete.queryParams.Get("param1"), clonedReq.queryParams.Get("param1"))
}

func (s *RequestTestSuite) TestSend_Success() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"success"}`))
	}))
	defer server.Close()

	resp, err := s.request.Clone().Get(server.URL)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(200, resp.Status())

	jsonData, err := resp.Json()
	s.NoError(err)
	s.Equal(map[string]any{"message": "success"}, jsonData)
}

func (s *RequestTestSuite) TestSend_Bind() {
	type Message struct {
		ID     int               `json:"id"`
		Name   string            `json:"name"`
		Active bool              `json:"active"`
		Scores []int             `json:"scores"`
		Meta   map[string]string `json:"meta"`
		Nested struct {
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"nested"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
          "id": 1,
          "name": "Test User",
          "active": true,
          "scores": [100, 95, 90],
          "meta": {"key1": "value1", "key2": "value2"},
          "nested": {"title": "Admin", "status": "Active"}
       }`))
	}))
	defer server.Close()

	var msg Message
	resp, err := s.request.Clone().AcceptJSON().Get(server.URL)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(200, resp.Status())
	s.NoError(resp.Bind(&msg))

	s.Equal(1, msg.ID)
	s.Equal("Test User", msg.Name)
	s.True(msg.Active)
	s.Equal([]int{100, 95, 90}, msg.Scores)
	s.Equal("value1", msg.Meta["key1"])
	s.Equal("Admin", msg.Nested.Title)
}

func (s *RequestTestSuite) TestSend_Stream() {
	expectedData := "chunk1-chunk2-chunk3"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedData))
	}))
	defer server.Close()

	resp, err := s.request.Clone().Get(server.URL)
	s.NoError(err)

	stream, err := resp.Stream()
	s.NoError(err)
	defer func(stream io.ReadCloser) {
		s.NoError(stream.Close())
	}(stream)

	content, err := io.ReadAll(stream)
	s.NoError(err)
	s.Equal(expectedData, string(content))

	resp2, err := s.request.Clone().Get(server.URL)
	s.NoError(err)

	bodyString, err := resp2.Body()
	s.NoError(err)
	s.Equal(expectedData, bodyString)

	stream2, err := resp2.Stream()
	s.NoError(err)

	content2, err := io.ReadAll(stream2)
	s.NoError(err)
	s.Equal(expectedData, string(content2))
}

func (s *RequestTestSuite) TestSend_Timeout() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set a context timeout shorter than the server sleep time
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req := s.request.Clone()
	_, err := req.WithContext(ctx).Get(server.URL)
	s.Error(err)
}

func (s *RequestTestSuite) TestSend_404() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resp, err := s.request.Clone().Get(server.URL)
	s.NoError(err)
	s.Equal(404, resp.Status())
}

func (s *RequestTestSuite) TestWithHeaders() {
	req := s.request.Clone().WithHeaders(map[string]string{
		"Content-Type": "application/json",
		"Old-Header":   "preserve-me",
	})
	s.Equal("application/json", req.(*Request).headers.Get("Content-Type"))

	req = req.ReplaceHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded", // Should overwrite
		"X-CUSTOM":     "custom-header",                     // Should add
	})

	s.Equal("application/x-www-form-urlencoded", req.(*Request).headers.Get("Content-Type"))
	s.Equal("custom-header", req.(*Request).headers.Get("X-CUSTOM"))
	s.Equal("preserve-me", req.(*Request).headers.Get("Old-Header"))

	req = req.WithoutHeader("X-CUSTOM")
	s.Empty(req.(*Request).headers.Get("X-CUSTOM"))

	req = req.FlushHeaders()
	s.Empty(req.(*Request).headers.Get("Content-Type"))
	s.Empty(req.(*Request).headers.Get("Old-Header"))
}

func (s *RequestTestSuite) TestWithBasicAuth() {
	req := s.request.Clone().WithBasicAuth("user", "pass")
	header := req.(*Request).headers.Get("Authorization")
	s.Contains(header, "Basic ")
}

func (s *RequestTestSuite) TestQueryParameterMethods() {
	req := s.request.Clone().WithQueryParameter("key", "value")
	s.Equal("value", req.(*Request).queryParams.Get("key"))

	req = req.WithQueryParameter("key", "newValue")
	s.Equal("newValue", req.(*Request).queryParams.Get("key"))

	req = req.WithQueryParameter("emptyValueKey", "")
	s.Equal("", req.(*Request).queryParams.Get("emptyValueKey"))

	req = s.request.Clone().WithQueryParameters(map[string]string{"key1": "value1", "key2": "value2"})
	s.Equal("value1", req.(*Request).queryParams.Get("key1"))
	s.Equal("value2", req.(*Request).queryParams.Get("key2"))

	req = req.WithQueryParameters(map[string]string{"key1": "newValue1"})
	s.Equal("newValue1", req.(*Request).queryParams.Get("key1"))

	req = req.WithQueryParameters(map[string]string{})
	s.Equal("newValue1", req.(*Request).queryParams.Get("key1"))

	req = s.request.Clone().WithQueryString("key1=value1&key2=value2")
	s.Equal("value1", req.(*Request).queryParams.Get("key1"))
	s.Equal("value2", req.(*Request).queryParams.Get("key2"))

	req = req.WithQueryString("multi=value1&multi=value2")
	s.Equal([]string{"value1", "value2"}, req.(*Request).queryParams["multi"])

	req = req.WithQueryString("empty=")
	s.Equal("", req.(*Request).queryParams.Get("empty"))

	req = req.WithQueryString("invalid%%%")
	s.Equal("value1", req.(*Request).queryParams.Get("key1"))
	s.Equal("", req.(*Request).queryParams.Get("key3"))

	req = req.WithQueryString("")
	s.Equal("value1", req.(*Request).queryParams.Get("key1"))
}

func (s *RequestTestSuite) TestAllHttpMethods() {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions}

	for _, method := range methods {
		s.Run(method, func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			var resp client.Response
			var err error

			req := s.request.Clone()
			switch method {
			case http.MethodGet:
				resp, err = req.Get(server.URL)
			case http.MethodPost:
				resp, err = req.Post(server.URL, nil)
			case http.MethodPut:
				resp, err = req.Put(server.URL, nil)
			case http.MethodDelete:
				resp, err = req.Delete(server.URL, nil)
			case http.MethodPatch:
				resp, err = req.Patch(server.URL, nil)
			case http.MethodHead:
				resp, err = req.Head(server.URL)
			case http.MethodOptions:
				resp, err = req.Options(server.URL)
			}

			s.NoError(err)
			s.NotNil(resp)
			s.Equal(http.StatusOK, resp.Status())
		})
	}
}

func (s *RequestTestSuite) TestPrefixAndPathVariables() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/{version}/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
		version := r.PathValue("version")
		userID := r.PathValue("userID")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{"id": %s, "name": "User %s", "version": "%s"}`, userID, userID, version)
		_, _ = w.Write([]byte(response))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	var user struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	req := s.request.Clone().WithUrlParameters(map[string]string{
		"version": "v1",
		"userID":  "1234",
	}).WithQueryParameter("role", "admin")

	resp, err := req.Get(server.URL + "/api/{version}/users/{userID}")
	s.NoError(err)
	s.NotNil(resp)
	s.NoError(resp.Bind(&user))
	s.Equal(200, resp.Status())
	s.Equal(1234, user.ID)
	s.Equal("User 1234", user.Name)
	s.Equal("v1", user.Version)
}

func (s *RequestTestSuite) TestConcurrentRequests() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Small delay to ensure context timeout triggers if set
		time.Sleep(5 * time.Millisecond)
		_, _ = w.Write(fmt.Appendf(nil, `{"message":"success-%s"}`, r.URL.Path))
	}))
	defer server.Close()

	reqTimeout := s.request.Clone()
	reqWithParams := s.request.Clone()
	req := s.request.Clone()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := reqTimeout.WithContext(ctx).Get(server.URL)
		s.Error(err, "Should timeout")
		s.Nil(resp)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := reqWithParams.WithUrlParameter("hello", "world").Get(server.URL)
		s.NoError(err)
		s.NotNil(resp)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := req.Get(server.URL)
		s.NoError(err)
		s.NotNil(resp)
	}()

	wg.Wait()
}

func (s *RequestTestSuite) TestParseRequestURL() {
	tests := []struct {
		name        string
		baseURL     string
		uri         string
		urlParams   map[string]string
		queryParams url.Values
		expected    string
		expectError bool
	}{
		{
			name:     "Absolute URL should remain unchanged",
			baseURL:  "https://api.goravel.dev",
			uri:      "https://external.com/data",
			expected: "https://external.com/data",
		},
		{
			name:     "Base URL should be prepended",
			baseURL:  "https://api.goravel.dev",
			uri:      "/users",
			expected: "https://api.goravel.dev/users",
		},
		{
			name:      "Path parameters should be replaced",
			baseURL:   "https://api.goravel.dev",
			uri:       "/users/{id}/posts",
			urlParams: map[string]string{"id": "123"},
			expected:  "https://api.goravel.dev/users/123/posts",
		},
		{
			name:     "Unresolved path parameters remain",
			baseURL:  "https://api.goravel.dev",
			uri:      "/users/{id}/posts",
			expected: "https://api.goravel.dev/users/%7Bid%7D/posts",
		},
		{
			name:        "Completely malformed URL should return an error",
			baseURL:     "https://api.goravel.dev",
			uri:         "http://:invalid",
			expectError: true,
		},
		{
			name:        "Query parameters should be appended",
			baseURL:     "https://api.goravel.dev",
			uri:         "/search",
			queryParams: url.Values{"q": []string{"golang"}, "page": []string{"1"}},
			expected:    "https://api.goravel.dev/search?page=1&q=golang",
		},
		{
			name:        "Existing query parameters should be preserved",
			baseURL:     "https://api.goravel.dev",
			uri:         "/search?sort=asc",
			queryParams: url.Values{"q": []string{"golang"}},
			expected:    "https://api.goravel.dev/search?sort=asc&q=golang",
		},
		{
			name:     "Unclosed path parameter should remain unchanged",
			baseURL:  "https://api.goravel.dev",
			uri:      "/users/{id",
			expected: "https://api.goravel.dev/users/%7Bid",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			r := &Request{
				client:      s.httpClient,
				baseUrl:     tt.baseURL,
				urlParams:   tt.urlParams,
				queryParams: tt.queryParams,
			}
			result, err := r.parseRequestURL(tt.uri)
			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expected, result)
			}
		})
	}
}
