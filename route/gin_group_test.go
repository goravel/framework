package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	mockconfig "github.com/goravel/framework/contracts/config/mocks"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/http/middleware"
	"github.com/goravel/framework/testing/mock"
)

func TestGinGroup(t *testing.T) {
	var (
		gin        *Gin
		mockConfig *mockconfig.Config
	)
	beforeEach := func() {
		mockConfig = mock.Config()
		mockConfig.On("GetBool", "app.debug").Return(true).Once()

		gin = NewGin()
	}
	tests := []struct {
		name       string
		setup      func(req *http.Request)
		method     string
		url        string
		expectCode int
		expectBody string
	}{
		{
			name: "Get",
			setup: func(req *http.Request) {
				gin.Get("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Json(http.StatusOK, httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "GET",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Post",
			setup: func(req *http.Request) {
				gin.Post("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "POST",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Put",
			setup: func(req *http.Request) {
				gin.Put("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "PUT",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Delete",
			setup: func(req *http.Request) {
				gin.Delete("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "DELETE",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Options",
			setup: func(req *http.Request) {
				gin.Options("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "OPTIONS",
			url:        "/input/1",
			expectCode: http.StatusOK,
		},
		{
			name: "Patch",
			setup: func(req *http.Request) {
				gin.Patch("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "PATCH",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Get",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "GET",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Post",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "POST",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Put",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "PUT",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Delete",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "DELETE",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Options",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.allowed_methods").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_origins").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_headers").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.exposed_headers").Return([]string{"*"}).Once()
				mockConfig.On("GetInt", "cors.max_age").Return(0).Once()
				mockConfig.On("GetBool", "cors.supports_credentials").Return(false).Once()
				gin.GlobalMiddleware(middleware.Cors())
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
				req.Header.Set("Origin", "http://127.0.0.1")
				req.Header.Set("Access-Control-Request-Method", "GET")
			},
			method:     "OPTIONS",
			url:        "/any/1",
			expectCode: http.StatusNoContent,
		},
		{
			name: "Any Patch",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "PATCH",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Static",
			setup: func(req *http.Request) {
				gin.Static("static", "../")
			},
			method:     "GET",
			url:        "/static/README.md",
			expectCode: http.StatusOK,
		},
		{
			name: "StaticFile",
			setup: func(req *http.Request) {
				gin.StaticFile("static-file", "../README.md")
			},
			method:     "GET",
			url:        "/static-file",
			expectCode: http.StatusOK,
		},
		{
			name: "StaticFS",
			setup: func(req *http.Request) {
				gin.StaticFS("static-fs", http.Dir("./"))
			},
			method:     "GET",
			url:        "/static-fs",
			expectCode: http.StatusMovedPermanently,
		},
		{
			name: "Abort Middleware",
			setup: func(req *http.Request) {
				gin.Middleware(abortMiddleware()).Get("/middleware/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "GET",
			url:        "/middleware/1",
			expectCode: http.StatusNonAuthoritativeInfo,
		},
		{
			name: "Multiple Middleware",
			setup: func(req *http.Request) {
				gin.Middleware(contextMiddleware(), contextMiddleware1()).Get("/middlewares/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id":   ctx.Request().Input("id"),
						"ctx":  ctx.Value("ctx"),
						"ctx1": ctx.Value("ctx1"),
					})
				})
			},
			method:     "GET",
			url:        "/middlewares/1",
			expectCode: http.StatusOK,
			expectBody: "{\"ctx\":\"Goravel\",\"ctx1\":\"Hello\",\"id\":\"1\"}",
		},
		{
			name: "Multiple Prefix",
			setup: func(req *http.Request) {
				gin.Prefix("prefix1").Prefix("prefix2").Get("input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "GET",
			url:        "/prefix1/prefix2/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Multiple Prefix Group Middleware",
			setup: func(req *http.Request) {
				gin.Prefix("group1").Middleware(contextMiddleware()).Group(func(route1 route.Route) {
					route1.Prefix("group2").Middleware(contextMiddleware1()).Group(func(route2 route.Route) {
						route2.Get("/middleware/{id}", func(ctx httpcontract.Context) {
							ctx.Response().Success().Json(httpcontract.Json{
								"id":   ctx.Request().Input("id"),
								"ctx":  ctx.Value("ctx").(string),
								"ctx1": ctx.Value("ctx1").(string),
							})
						})
					})
					route1.Middleware(contextMiddleware2()).Get("/middleware/{id}", func(ctx httpcontract.Context) {
						ctx.Response().Success().Json(httpcontract.Json{
							"id":   ctx.Request().Input("id"),
							"ctx":  ctx.Value("ctx").(string),
							"ctx2": ctx.Value("ctx2").(string),
						})
					})
				})
			},
			method:     "GET",
			url:        "/group1/group2/middleware/1",
			expectCode: http.StatusOK,
			expectBody: "{\"ctx\":\"Goravel\",\"ctx1\":\"Hello\",\"id\":\"1\"}",
		},
		{
			name: "Multiple Group Middleware",
			setup: func(req *http.Request) {
				gin.Prefix("group1").Middleware(contextMiddleware()).Group(func(route1 route.Route) {
					route1.Prefix("group2").Middleware(contextMiddleware1()).Group(func(route2 route.Route) {
						route2.Get("/middleware/{id}", func(ctx httpcontract.Context) {
							ctx.Response().Success().Json(httpcontract.Json{
								"id":   ctx.Request().Input("id"),
								"ctx":  ctx.Value("ctx").(string),
								"ctx1": ctx.Value("ctx1").(string),
							})
						})
					})
					route1.Middleware(contextMiddleware2()).Get("/middleware/{id}", func(ctx httpcontract.Context) {
						ctx.Response().Success().Json(httpcontract.Json{
							"id":   ctx.Request().Input("id"),
							"ctx":  ctx.Value("ctx").(string),
							"ctx2": ctx.Value("ctx2").(string),
						})
					})
				})
			},
			method:     "GET",
			url:        "/group1/middleware/1",
			expectCode: http.StatusOK,
			expectBody: "{\"ctx\":\"Goravel\",\"ctx2\":\"World\",\"id\":\"1\"}",
		},
		{
			name: "Global Middleware",
			setup: func(req *http.Request) {
				gin.GlobalMiddleware(func(ctx httpcontract.Context) {
					ctx.WithValue("global", "goravel")
					ctx.Request().Next()
				})
				gin.Get("/global-middleware", func(ctx httpcontract.Context) {
					ctx.Response().Json(http.StatusOK, httpcontract.Json{
						"global": ctx.Value("global"),
					})
				})
			},
			method:     "GET",
			url:        "/global-middleware",
			expectCode: http.StatusOK,
			expectBody: "{\"global\":\"goravel\"}",
		},
	}
	for _, test := range tests {
		beforeEach()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(test.method, test.url, nil)
		if test.setup != nil {
			test.setup(req)
		}
		gin.ServeHTTP(w, req)

		if test.expectBody != "" {
			assert.Equal(t, test.expectBody, w.Body.String(), test.name)
		}
		assert.Equal(t, test.expectCode, w.Code, test.name)
	}
}

func abortMiddleware() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		ctx.Request().AbortWithStatus(http.StatusNonAuthoritativeInfo)
	}
}

func contextMiddleware() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		ctx.WithValue("ctx", "Goravel")

		ctx.Request().Next()
	}
}

func contextMiddleware1() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		ctx.WithValue("ctx1", "Hello")

		ctx.Request().Next()
	}
}

func contextMiddleware2() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		ctx.WithValue("ctx2", "World")

		ctx.Request().Next()
	}
}
