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
		Clients: map[string]client.Config{
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
	s.factory = NewFactory(s.config, s.json)
}

func (s *FactoryTestSuite) TestClient_Resolution() {
	s.Run("resolves default client", func() {
		c := s.factory.Client()
		s.NotNil(c)
		s.Equal("main", c.Name())
		s.Equal("https://main.com", c.Config().BaseUrl)
	})

	s.Run("resolves specific client", func() {
		c := s.factory.Client("stripe")
		s.NotNil(c)
		s.Equal("stripe", c.Name())
		s.Equal("https://api.stripe.com", c.Config().BaseUrl)
	})

	s.Run("caches client instances (Singleton)", func() {
		c1 := s.factory.Client("stripe")
		c2 := s.factory.Client("stripe")

		// Ensure we aren't re-allocating memory for the same CLIENT
		s.Same(c1, c2, "Factory should return the exact same instance for the same name")

		c3 := s.factory.Client("main")
		s.NotSame(c1, c3, "Different clients must be different instances")
	})
}

func (s *FactoryTestSuite) TestErrorHandling() {
	s.Run("handles nil config safely", func() {
		f := NewFactory(nil, s.json)
		s.NotNil(f)

		// Should return lazy error because no default is configured
		resp, err := f.Client().NewRequest().Get("/")
		s.Nil(resp)
		s.ErrorIs(err, errors.HttpClientDefaultNotSet, "Expected HttpClientDefaultNotSet error")
	})

	s.Run("returns lazy error for missing client", func() {
		c := s.factory.Client("missing_client")
		s.NotNil(c)

		// The error should only trigger when we attempt a request
		resp, err := c.NewRequest().Get("/")
		s.Nil(resp)
		s.ErrorIs(err, errors.HttpClientConnectionNotFound, "Expected HttpClientDefaultNotSet error")
		s.Contains(err.Error(), "[missing_client]")
	})

	s.Run("returns lazy error when default is empty in config", func() {
		cfg := &FactoryConfig{
			Default: "",
			Clients: map[string]client.Config{"main": {}},
		}
		f := NewFactory(cfg, s.json)

		resp, err := f.Client().NewRequest().Get("/")
		s.Nil(resp)
		s.Error(err)
		s.ErrorIs(err, errors.HttpClientDefaultNotSet)
	})
}

func (s *FactoryTestSuite) TestRouting_Integration() {
	serverA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("response_from_A"))
	}))
	defer serverA.Close()

	// Server B (Simulates Stripe)
	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("response_from_B"))
	}))
	defer serverB.Close()

	// Setup Factory with dynamic URLs
	cfg := &FactoryConfig{
		Default: "server_a",
		Clients: map[string]client.Config{
			"server_a": {BaseUrl: serverA.URL},
			"server_b": {BaseUrl: serverB.URL},
		},
	}
	f := NewFactory(cfg, s.json)

	s.Run("proxy methods hit default server", func() {
		resp, err := f.Get("/")
		s.NoError(err)
		body, _ := resp.Body()
		s.Equal("response_from_A", body)
	})

	s.Run("named request hits specific server", func() {
		resp, err := f.Request("server_b").Get("/")
		s.NoError(err)
		body, _ := resp.Body()
		s.Equal("response_from_B", body)
	})
}

func (s *FactoryTestSuite) TestConcurrency() {
	cfg := &FactoryConfig{
		Default: "main",
		Clients: map[string]client.Config{
			"main": {BaseUrl: "https://main.com"},
			"new1": {BaseUrl: "https://new1.com"},
			"new2": {BaseUrl: "https://new2.com"},
		},
	}
	f := NewFactory(cfg, s.json)

	var wg sync.WaitGroup
	// Pre-warm "main"
	f.Client("main")

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Mixed workload:
			if idx%2 == 0 {
				// Access existing (RLock behavior)
				c := f.Client("main")
				s.Equal("main", c.Name())
			} else {
				// Access potentially new (Lock behavior)
				// We toggle between existing and new to stress the mutex
				name := "new1"
				if idx%3 == 0 {
					name = "new2"
				}
				c := f.Client(name)
				s.Equal(name, c.Name())
			}
		}(i)
	}
	wg.Wait()
}

func (s *FactoryTestSuite) TestProxy_HttpMethods() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	f := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]client.Config{"test": {BaseUrl: server.URL}}}, s.json)

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
		{"Head", func() (client.Response, error) { return f.Head("/") }, ""}, // HEAD returns no body
		{"Options", func() (client.Response, error) { return f.Options("/") }, "OPTIONS"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp, err := tt.action()
			s.NoError(err)
			if tt.expect != "" {
				body, _ := resp.Body()
				s.Equal(tt.expect, body)
			} else {
				s.Equal(200, resp.Status())
			}
		})
	}
}

func (s *FactoryTestSuite) TestProxy_Headers() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo specific headers back as JSON for validation
		headers := map[string]string{
			"Content-Type":  r.Header.Get("Content-Type"),
			"Authorization": r.Header.Get("Authorization"),
			"Accept":        r.Header.Get("Accept"),
			"X-Custom":      r.Header.Get("X-Custom"),
			"X-Multi":       r.Header.Get("X-Multi"),
		}
		output, _ := s.json.Marshal(headers)
		_, _ = w.Write(output)
	}))
	defer server.Close()

	f := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]client.Config{"test": {BaseUrl: server.URL}}}, s.json)

	s.Run("WithHeader & WithHeaders", func() {
		resp, _ := f.WithHeader("X-Custom", "1").WithHeaders(map[string]string{"X-Multi": "2"}).Get("/")
		var h map[string]string
		_ = resp.Bind(&h)
		s.Equal("1", h["X-Custom"])
		s.Equal("2", h["X-Multi"])
	})

	s.Run("ReplaceHeaders & FlushHeaders", func() {
		// Replace should overwrite previous headers
		resp, _ := f.WithHeader("A", "B").ReplaceHeaders(map[string]string{"X-Custom": "replaced"}).Get("/")
		var h map[string]string
		_ = resp.Bind(&h)
		s.Equal("replaced", h["X-Custom"])
		s.Equal("", h["Authorization"])

		// Flush should clear all
		resp2, _ := f.WithHeader("A", "B").FlushHeaders().Get("/")
		var h2 map[string]string
		_ = resp2.Bind(&h2)
		s.Equal("", h2["X-Custom"])
	})

	s.Run("WithoutHeader", func() {
		resp, _ := f.WithHeader("X-Custom", "val").WithoutHeader("X-Custom").Get("/")
		var h map[string]string
		_ = resp.Bind(&h)
		s.Equal("", h["X-Custom"])
	})

	s.Run("Auth Helpers", func() {
		// WithToken
		resp, _ := f.WithToken("secret").Get("/")
		var h map[string]string
		_ = resp.Bind(&h)
		s.Equal("Bearer secret", h["Authorization"])

		// WithoutToken
		resp2, _ := f.WithToken("secret").WithoutToken().Get("/")
		var h2 map[string]string
		_ = resp2.Bind(&h2)
		s.Equal("", h2["Authorization"])

		// WithBasicAuth
		resp3, _ := f.WithBasicAuth("user", "pass").Get("/")
		var h3 map[string]string
		_ = resp3.Bind(&h3)
		s.Contains(h3["Authorization"], "Basic ")
	})

	s.Run("Content Type Helpers", func() {
		resp, _ := f.Accept("text/html").AsForm().Get("/")
		var h map[string]string
		_ = resp.Bind(&h)
		s.Equal("text/html", h["Accept"])
		s.Equal("application/x-www-form-urlencoded", h["Content-Type"])

		resp2, _ := f.AcceptJSON().Get("/")
		var h2 map[string]string
		_ = resp2.Bind(&h2)
		s.Equal("application/json", h2["Accept"])
	})
}

func (s *FactoryTestSuite) TestProxy_QueryParameters() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.RawQuery))
	}))
	defer server.Close()

	f := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]client.Config{"test": {BaseUrl: server.URL}}}, s.json)

	s.Run("WithQueryParameter", func() {
		resp, _ := f.WithQueryParameter("page", "1").Get("/")
		body, _ := resp.Body()
		s.Contains(body, "page=1")
	})

	s.Run("WithQueryParameters", func() {
		resp, _ := f.WithQueryParameters(map[string]string{"sort": "asc", "limit": "10"}).Get("/")
		body, _ := resp.Body()
		s.Contains(body, "sort=asc")
		s.Contains(body, "limit=10")
	})

	s.Run("WithQueryString", func() {
		resp, _ := f.WithQueryString("raw=true&manual=1").Get("/")
		body, _ := resp.Body()
		s.Contains(body, "raw=true")
		s.Contains(body, "manual=1")
	})
}

func (s *FactoryTestSuite) TestProxy_UrlParameters() {
	// URL Parameters modify the path, e.g. /users/{id}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer server.Close()

	f := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]client.Config{"test": {BaseUrl: server.URL}}}, s.json)

	s.Run("WithUrlParameter", func() {
		// Note: The mock server URL doesn't have brackets, so we append them for the test
		resp, _ := f.WithUrlParameter("id", "42").Get("/users/{id}")
		body, _ := resp.Body()
		s.Equal("/users/42", body)
	})

	s.Run("WithUrlParameters", func() {
		resp, _ := f.WithUrlParameters(map[string]string{"id": "99", "action": "edit"}).Get("/users/{id}/{action}")
		body, _ := resp.Body()
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

	f := NewFactory(&FactoryConfig{Default: "test", Clients: map[string]client.Config{"test": {BaseUrl: server.URL}}}, s.json)

	s.Run("WithCookie", func() {
		cookie := &http.Cookie{Name: "session", Value: "abc-123"}
		resp, _ := f.WithCookie(cookie).Get("/")
		body, _ := resp.Body()
		s.Equal("abc-123", body)
	})

	s.Run("WithCookies", func() {
		cookies := []*http.Cookie{
			{Name: "session", Value: "multi-cookie"},
		}
		resp, _ := f.WithCookies(cookies).Get("/")
		body, _ := resp.Body()
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
