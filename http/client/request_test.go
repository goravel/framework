package client

import (
	"context"
	encodingjson "encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/mocks/config"
)

type RequestTestSuite struct {
	suite.Suite
	mockConfig *config.Config
	request    client.Request
	once       sync.Once
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(RequestTestSuite))
}

func (s *RequestTestSuite) SetupTest() {
	s.mockConfig = config.NewConfig(s.T())

	s.once.Do(func() {
		s.mockConfig.EXPECT().GetDuration("http.client.timeout", 30*time.Second).Return(30 * time.Second)
		s.mockConfig.EXPECT().GetInt("http.client.max_idle_conns").Return(0)
		s.mockConfig.EXPECT().GetInt("http.client.max_idle_conns_per_host").Return(0)
		s.mockConfig.EXPECT().GetInt("http.client.max_conns_per_host").Return(0)
		s.mockConfig.EXPECT().GetDuration("http.client.idle_conn_timeout").Return(30 * time.Second)
	})

	s.request = NewRequest(s.mockConfig, &testJson{})
}

func (s *RequestTestSuite) TestClone() {
	req := s.request.Clone().WithQueryParameter("key", "value")
	cookie := &http.Cookie{Name: "session", Value: "abc123"}
	req = req.WithCookie(cookie)

	clonedReq := req.Clone().(*requestImpl)

	s.Equal(req.(*requestImpl).queryParams.Encode(), clonedReq.queryParams.Encode())
	clonedReq = clonedReq.WithQueryParameter("newKey", "newValue").(*requestImpl)
	s.NotEqual(req.(*requestImpl).queryParams.Encode(), clonedReq.queryParams.Encode())

	// Ensure cookies are copied properly
	s.Equal(len(req.(*requestImpl).cookies), len(clonedReq.cookies))
	s.Equal(req.(*requestImpl).cookies[0].Value, clonedReq.cookies[0].Value)

	// Ensure modifying cloned cookies does not affect the original
	clonedReq = clonedReq.WithCookie(&http.Cookie{Name: "session", Value: "modified"}).(*requestImpl)
	s.NotEqual(req.(*requestImpl).cookies[0].Value, clonedReq.cookies[len(clonedReq.cookies)-1].Value)

	req = req.WithQueryParameter("param1", "value1").(*requestImpl)
	clonedReq = req.Clone().(*requestImpl)
	s.Equal(req.(*requestImpl).queryParams.Get("param1"), clonedReq.queryParams.Get("param1"))
	clonedReq = clonedReq.WithQueryParameter("param1", "changedValue").(*requestImpl)
	s.NotEqual(req.(*requestImpl).queryParams.Get("param1"), clonedReq.queryParams.Get("param1"))
}

func (s *RequestTestSuite) TestDoRequest_Success() {
	s.mockConfig.EXPECT().GetString("http.client.base_url", "").Return("")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"success"}`))
	}))
	defer server.Close()

	resp, err := s.request.Clone().Get(server.URL)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(200, resp.Status())

	jsonData, err := resp.Json()
	s.NoError(err)
	s.Equal(map[string]interface{}{"message": "success"}, jsonData)
}
func (s *RequestTestSuite) TestDoRequest_Bind() {
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

	s.mockConfig.EXPECT().GetString("http.client.base_url", "").Return("")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
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
	resp, err := s.request.Clone().AcceptJSON().Bind(&msg).Get(server.URL)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(200, resp.Status())

	s.Equal(1, msg.ID)
	s.Equal("Test User", msg.Name)
	s.Equal(true, msg.Active)
	s.Equal([]int{100, 95, 90}, msg.Scores)
	s.Equal("value1", msg.Meta["key1"])
	s.Equal("Admin", msg.Nested.Title)
}

func (s *RequestTestSuite) TestDoRequest_Timeout() {
	s.mockConfig.EXPECT().GetString("http.client.base_url", "").Return("")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	req := s.request.Clone()
	_, err := req.WithContext(ctx).Get(server.URL)
	s.Error(err)
}

func (s *RequestTestSuite) TestDoRequest_404() {
	s.mockConfig.EXPECT().GetString("http.client.base_url", "").Return("")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resp, err := s.request.Clone().Get(server.URL)
	s.NoError(err)
	s.Equal(404, resp.Status())
}

func (s *RequestTestSuite) TestWithHeaders() {
	req := s.request.Clone().WithHeaders(map[string]string{"Content-Type": "application/json"})
	s.Equal("application/json", req.(*requestImpl).headers.Get("Content-Type"))

	req.ReplaceHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"X-CUSTOM":     "custom-header",
	})
	s.Equal("application/x-www-form-urlencoded", req.(*requestImpl).headers.Get("Content-Type"))
	s.Equal("custom-header", req.(*requestImpl).headers.Get("X-CUSTOM"))

	req.WithoutHeader("X-CUSTOM")
	s.Equal("", req.(*requestImpl).headers.Get("X-CUSTOM"))

	req.FlushHeaders()
	s.Equal("", req.(*requestImpl).headers.Get("Content-Type"))
}

func (s *RequestTestSuite) TestWithBasicAuth() {
	req := s.request.Clone().WithBasicAuth("user", "pass")
	header := req.(*requestImpl).headers.Get("Authorization")
	s.Contains(header, "Basic ")
}

func (s *RequestTestSuite) TestQueryParameterMethods() {
	req := s.request.Clone().WithQueryParameter("key", "value")
	s.Equal("value", req.(*requestImpl).queryParams.Get("key"))

	req = req.WithQueryParameter("key", "newValue")
	s.Equal("newValue", req.(*requestImpl).queryParams.Get("key"))

	req = req.WithQueryParameter("emptyValueKey", "")
	s.Equal("", req.(*requestImpl).queryParams.Get("emptyValueKey"))

	req = s.request.Clone().WithQueryParameters(map[string]string{"key1": "value1", "key2": "value2"})
	s.Equal("value1", req.(*requestImpl).queryParams.Get("key1"))
	s.Equal("value2", req.(*requestImpl).queryParams.Get("key2"))

	req = req.WithQueryParameters(map[string]string{"key1": "newValue1"})
	s.Equal("newValue1", req.(*requestImpl).queryParams.Get("key1"))

	req = req.WithQueryParameters(map[string]string{})
	s.Equal("newValue1", req.(*requestImpl).queryParams.Get("key1"))

	req = s.request.Clone().WithQueryString("key1=value1&key2=value2")
	s.Equal("value1", req.(*requestImpl).queryParams.Get("key1"))
	s.Equal("value2", req.(*requestImpl).queryParams.Get("key2"))

	req = req.WithQueryString("multi=value1&multi=value2")
	s.Equal([]string{"value1", "value2"}, req.(*requestImpl).queryParams["multi"])

	req = req.WithQueryString("empty=")
	s.Equal("", req.(*requestImpl).queryParams.Get("empty"))

	req = req.WithQueryString("invalid%%%")
	s.Equal("value1", req.(*requestImpl).queryParams.Get("key1"))
	s.Equal("", req.(*requestImpl).queryParams.Get("key3"))

	req = req.WithQueryString("")
	s.Equal("value1", req.(*requestImpl).queryParams.Get("key1"))
}

func (s *RequestTestSuite) TestAllHttpMethods() {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions}

	for _, method := range methods {
		s.Run(method, func() {
			s.mockConfig.EXPECT().GetString("http.client.base_url", "").Return("")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			var resp client.Response
			var err error
			switch method {
			case http.MethodGet:
				resp, err = s.request.Clone().Get(server.URL)
			case http.MethodPost:
				resp, err = s.request.Clone().Post(server.URL, nil)
			case http.MethodPut:
				resp, err = s.request.Clone().Put(server.URL, nil)
			case http.MethodDelete:
				resp, err = s.request.Clone().Delete(server.URL, nil)
			case http.MethodPatch:
				resp, err = s.request.Clone().Patch(server.URL, nil)
			case http.MethodHead:
				resp, err = s.request.Clone().Head(server.URL)
			case http.MethodOptions:
				resp, err = s.request.Clone().Options(server.URL)
			}

			s.NoError(err)
			s.NotNil(resp)
			s.Equal(http.StatusOK, resp.Status())
		})
	}
}

func (s *RequestTestSuite) TestPrefixAndPathVariables() {
	s.mockConfig.EXPECT().GetString("http.client.base_url", "").Return("")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/{version}/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
		version := r.PathValue("version")
		userID := r.PathValue("userID")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{"id": %s, "name": "User %s", "version": "%s"}`, userID, userID, version)
		w.Write([]byte(response))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	var user struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	req := s.request.Clone().Bind(&user).WithUrlParameters(map[string]string{
		"version": "v1",
		"userID":  "1234",
	}).WithQueryParameter("role", "admin")
	resp, err := req.Get(server.URL + "/api/{version}/users/{userID}")
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(200, resp.Status())
	s.Equal(1234, user.ID)
	s.Equal("User 1234", user.Name)
}

func (s *RequestTestSuite) TestConcurrentRequests() {
	s.mockConfig.EXPECT().GetString("http.client.base_url", "").Return("")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		time.Sleep(5 * time.Millisecond)
		w.Write([]byte(fmt.Sprintf(`{"message":"success-%s"}`, r.URL.Path)))
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
		s.Error(err)
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

type testJson struct{}

func (t *testJson) Marshal(v any) ([]byte, error) {
	return encodingjson.Marshal(v)
}

func (t *testJson) Unmarshal(data []byte, v any) error {
	return encodingjson.Unmarshal(data, v)
}
