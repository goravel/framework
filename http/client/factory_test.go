package client

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
)

type FactoryTestSuite struct {
	suite.Suite
	json    foundation.Json
	factory *Factory
	config  *FactoryConfig
}

func TestFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}

func (s *FactoryTestSuite) SetupTest() {
	s.json = json.New()
	s.config = &FactoryConfig{
		Default: "main",
		Clients: map[string]Config{
			"main": {
				BaseUrl: "https://main.com",
				Timeout: 10 * time.Second,
			},
			"stripe": {
				BaseUrl: "https://api.stripe.com",
				Timeout: 5 * time.Second,
			},
		},
	}
	var err error
	s.factory, err = NewFactory(s.config, s.json)
	s.NoError(err)
}

func (s *FactoryTestSuite) TestClient_Resolution() {
	s.Run("resolves default client", func() {
		req := s.factory.Client()
		s.NotNil(req)

		s.Equal(10*time.Second, req.HttpClient().Timeout)
	})

	s.Run("resolves specific client", func() {
		req := s.factory.Client("stripe")
		s.NotNil(req)
		s.Equal(5*time.Second, req.HttpClient().Timeout)
	})

	s.Run("caches http client instances (Singleton Pool)", func() {
		req1 := s.factory.Client("stripe")
		req2 := s.factory.Client("stripe")

		s.Same(req1.HttpClient(), req2.HttpClient(), "Factory should return the exact same *http.Client for the same name")

		req3 := s.factory.Client("main")
		s.NotSame(req1.HttpClient(), req3.HttpClient(), "Different config names must result in different *http.Client instances")
	})
}

func (s *FactoryTestSuite) TestErrorHandling() {
	s.Run("handles nil config safely", func() {
		f, err := NewFactory(nil, s.json)
		s.Nil(f)
		s.ErrorIs(err, errors.HttpClientConfigNotSet)
	})

	s.Run("returns lazy error for missing client", func() {
		req := s.factory.Client("missing_client")
		s.NotNil(req)

		resp, err := req.Get("/")
		s.Nil(resp)
		s.ErrorIs(err, errors.HttpClientConnectionNotFound)
		s.Contains(err.Error(), "[missing_client]")
	})

	s.Run("returns lazy error when default is empty in config", func() {
		cfg := &FactoryConfig{
			Default: "",
			Clients: map[string]Config{"main": {}},
		}
		f, err := NewFactory(cfg, s.json)
		s.ErrorIs(err, errors.HttpClientDefaultNotSet)
		s.Nil(f)
	})
}

func (s *FactoryTestSuite) TestRouting_Integration() {
	serverA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("response_from_A"))
	}))
	defer serverA.Close()

	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("response_from_B"))
	}))
	defer serverB.Close()

	cfg := &FactoryConfig{
		Default: "server_a",
		Clients: map[string]Config{
			"server_a": {BaseUrl: serverA.URL},
			"server_b": {BaseUrl: serverB.URL},
		},
	}
	f, err := NewFactory(cfg, s.json)
	s.NoError(err)

	s.Run("proxy methods hit default server", func() {
		resp, err := f.Get("/")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)
		s.Equal("response_from_A", body)
	})

	s.Run("named request hits specific server", func() {
		resp, err := f.Client("server_b").Get("/")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)
		s.Equal("response_from_B", body)
	})
}

func (s *FactoryTestSuite) TestConcurrency() {
	cfg := &FactoryConfig{
		Default: "main",
		Clients: map[string]Config{
			"main": {BaseUrl: "https://main.com", Timeout: 1 * time.Second},
			"new1": {BaseUrl: "https://new1.com", Timeout: 2 * time.Second},
			"new2": {BaseUrl: "https://new2.com", Timeout: 3 * time.Second},
		},
	}
	f, err := NewFactory(cfg, s.json)
	s.NoError(err)

	var wg sync.WaitGroup

	timeoutMap := map[string]time.Duration{
		"main": 1 * time.Second,
		"new1": 2 * time.Second,
		"new2": 3 * time.Second,
	}

	f.Client("main") // Pre-warm

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			var name string
			if idx%2 == 0 {
				name = "main"
			} else {
				name = "new1"
				if idx%3 == 0 {
					name = "new2"
				}
			}

			req := f.Client(name)
			s.NotNil(req)

			s.Equal(timeoutMap[name], req.HttpClient().Timeout)
			concreteReq, ok := req.(*Request)
			s.True(ok)
			s.Equal(cfg.Clients[name].BaseUrl, concreteReq.baseUrl)
		}(i)
	}
	wg.Wait()
}

func (s *FactoryTestSuite) TestBaseUrl_Override() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("hit"))
	}))
	defer server.Close()

	cfg := &FactoryConfig{
		Default: "main",
		Clients: map[string]Config{
			"main": {BaseUrl: "https://wrong-url.com"},
		},
	}
	f, err := NewFactory(cfg, s.json)
	s.NoError(err)

	s.Run("overrides config base url", func() {
		resp, err := f.BaseUrl(server.URL).Get("/")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)
		s.Equal("hit", body)
	})
}

func (s *FactoryTestSuite) TestProxy_HttpMethods() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	f, err := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]Config{"test": {BaseUrl: server.URL}}}, s.json)
	s.NoError(err)

	tests := []struct {
		name   string
		action func() (client.Response, error)
		expect string
	}{
		{"Get", func() (client.Response, error) { return f.Get("/") }, "GET"},
		{"Post", func() (client.Response, error) { return f.Post("/", nil) }, "POST"},
		{"Put", func() (client.Response, error) { return f.Put("/", nil) }, "PUT"},
		{"Patch", func() (client.Response, error) { return f.Patch("/", nil) }, "PATCH"},
		{"Delete", func() (client.Response, error) { return f.Delete("/", nil) }, "DELETE"},
		{"Head", func() (client.Response, error) { return f.Head("/") }, ""},
		{"Options", func() (client.Response, error) { return f.Options("/") }, "OPTIONS"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp, err := tt.action()
			s.NoError(err)
			if tt.expect != "" {
				body, err := resp.Body()
				s.NoError(err)
				s.Equal(tt.expect, body)
			} else {
				s.Equal(200, resp.Status())
			}
		})
	}
}

func (s *FactoryTestSuite) TestProxy_Headers() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := map[string]string{
			"Content-Type":  r.Header.Get("Content-Type"),
			"Authorization": r.Header.Get("Authorization"),
			"Accept":        r.Header.Get("Accept"),
			"X-Custom":      r.Header.Get("X-Custom"),
			"X-Multi":       r.Header.Get("X-Multi"),
			"A":             r.Header.Get("A"),
		}
		output, _ := s.json.Marshal(headers)
		_, _ = w.Write(output)
	}))
	defer server.Close()

	f, err := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]Config{"test": {BaseUrl: server.URL}}}, s.json)
	s.NoError(err)

	s.Run("WithHeader & WithHeaders", func() {
		resp, err := f.WithHeader("X-Custom", "1").WithHeaders(map[string]string{"X-Multi": "2"}).Get("/")
		s.NoError(err)

		var h map[string]string
		s.NoError(resp.Bind(&h))
		s.Equal("1", h["X-Custom"])
		s.Equal("2", h["X-Multi"])
	})

	s.Run("ReplaceHeaders & FlushHeaders", func() {
		resp, err := f.WithHeader("A", "B").ReplaceHeaders(map[string]string{"X-Custom": "replaced"}).Get("/")
		s.NoError(err)

		var h map[string]string
		s.NoError(resp.Bind(&h))
		s.Equal("replaced", h["X-Custom"])
		s.Equal("B", h["A"], "Expected 'ReplaceHeaders' to merge/preserve existing headers, not wipe them")

		resp2, err := f.WithHeader("A", "B").FlushHeaders().Get("/")
		s.NoError(err)

		var h2 map[string]string
		s.NoError(resp2.Bind(&h2))
		s.Equal("", h2["A"], "Expected 'FlushHeaders' to remove previous headers")
	})

	s.Run("WithoutHeader", func() {
		resp, err := f.WithHeader("X-Custom", "val").WithoutHeader("X-Custom").Get("/")
		s.NoError(err)

		var h map[string]string
		s.NoError(resp.Bind(&h))
		s.Equal("", h["X-Custom"])
	})

	s.Run("Auth Helpers", func() {
		resp, err := f.WithToken("secret").Get("/")
		s.NoError(err)

		var h map[string]string
		s.NoError(resp.Bind(&h))
		s.Equal("Bearer secret", h["Authorization"])

		resp2, err := f.WithToken("secret").WithoutToken().Get("/")
		s.NoError(err)

		var h2 map[string]string
		s.NoError(resp2.Bind(&h2))
		s.Equal("", h2["Authorization"])

		resp3, err := f.WithBasicAuth("user", "pass").Get("/")
		s.NoError(err)

		var h3 map[string]string
		s.NoError(resp3.Bind(&h3))
		s.Contains(h3["Authorization"], "Basic ")
	})

	s.Run("Content Type Helpers", func() {
		resp, err := f.Accept("text/html").AsForm().Get("/")
		s.NoError(err)

		var h map[string]string
		s.NoError(resp.Bind(&h))
		s.Equal("text/html", h["Accept"])
		s.Equal("application/x-www-form-urlencoded", h["Content-Type"])

		resp2, err := f.AcceptJSON().Get("/")
		s.NoError(err)

		var h2 map[string]string
		s.NoError(resp2.Bind(&h2))
		s.Equal("application/json", h2["Accept"])
	})
}

func (s *FactoryTestSuite) TestProxy_QueryParameters() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.RawQuery))
	}))
	defer server.Close()

	f, err := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]Config{"test": {BaseUrl: server.URL}}}, s.json)
	s.NoError(err)

	s.Run("WithQueryParameter", func() {
		resp, err := f.WithQueryParameter("page", "1").Get("/")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)

		s.Contains(body, "page=1")
	})

	s.Run("WithQueryParameters", func() {
		resp, err := f.WithQueryParameters(map[string]string{"sort": "asc", "limit": "10"}).Get("/")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)

		s.Contains(body, "sort=asc")
		s.Contains(body, "limit=10")
	})

	s.Run("WithQueryString", func() {
		resp, err := f.WithQueryString("raw=true&manual=1").Get("/")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)

		s.Contains(body, "raw=true")
		s.Contains(body, "manual=1")
	})
}

func (s *FactoryTestSuite) TestProxy_UrlParameters() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer server.Close()

	f, err := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]Config{"test": {BaseUrl: server.URL}}}, s.json)
	s.NoError(err)

	s.Run("WithUrlParameter", func() {
		resp, err := f.WithUrlParameter("id", "42").Get("/users/{id}")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)
		s.Equal("/users/42", body)
	})

	s.Run("WithUrlParameters", func() {
		resp, err := f.WithUrlParameters(map[string]string{"id": "99", "action": "edit"}).Get("/users/{id}/{action}")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)
		s.Equal("/users/99/edit", body)
	})
}

func (s *FactoryTestSuite) TestProxy_Cookies() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := r.Cookie("session")
		if c != nil {
			_, _ = w.Write([]byte(c.Value))
		}
	}))
	defer server.Close()

	f, err := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]Config{"test": {BaseUrl: server.URL}}}, s.json)
	s.NoError(err)

	s.Run("WithCookie", func() {
		cookie := &http.Cookie{Name: "session", Value: "abc-123"}
		resp, err := f.WithCookie(cookie).Get("/")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)
		s.Equal("abc-123", body)
	})

	s.Run("WithCookies", func() {
		cookies := []*http.Cookie{
			{Name: "session", Value: "multi-cookie"},
		}
		resp, err := f.WithCookies(cookies).Get("/")
		s.NoError(err)

		body, err := resp.Body()
		s.NoError(err)
		s.Equal("multi-cookie", body)
	})
}

func (s *FactoryTestSuite) TestProxy_Misc() {
	s.Run("Clone", func() {
		req1 := s.factory.WithHeader("A", "B")
		req2 := req1.Clone()

		s.NotSame(req1, req2)
	})
}
