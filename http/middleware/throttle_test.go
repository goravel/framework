package middleware

import (
	"context"
	"errors"
	nethttp "net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cachemocks "github.com/goravel/framework/contracts/cache/mocks"
	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	httpmocks "github.com/goravel/framework/contracts/http/mocks"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/http/limit"
	"github.com/goravel/framework/support/carbon"
)

func TestThrottle(t *testing.T) {
	var (
		ctx                   *TestContext
		mockCache             *cachemocks.Cache
		mockConfig            *configmocks.Config
		mockRateLimiterFacade *httpmocks.RateLimiter
	)

	now := carbon.Now()
	carbon.SetTestNow(now)

	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "empty limiter",
			setup: func() {
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{}
				}).Once()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
		{
			name: "not http limit",
			setup: func() {
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						&TestLimit{},
					}
				}).Once()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
		{
			name: "success when first request",
			setup: func() {
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limit.PerMinute(1),
					}
				}).Once()
				mockConfig.On("GetString", "cache.prefix").Return("goravel").Once()
				mockCache.On("Has", mock.MatchedBy(func(timer string) bool {
					return strings.HasSuffix(timer, ":timer")
				})).Return(false).Once()
				mockCache.On("Put", mock.MatchedBy(func(timer string) bool {
					return strings.HasSuffix(timer, ":timer")
				}), now.Timestamp(), time.Duration(1)*time.Minute).Return(nil).Once()
				mockCache.On("Put", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "goravel:throttle:test:")
				}), 1, time.Duration(1)*time.Minute).Return(nil).Once()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Equal(t, "1", ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Equal(t, "0", ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
		{
			name: "error when put timer fail in first request",
			setup: func() {
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limit.PerMinute(1),
					}
				}).Once()
				mockConfig.On("GetString", "cache.prefix").Return("goravel").Once()
				mockCache.On("Has", mock.MatchedBy(func(timer string) bool {
					return strings.HasSuffix(timer, ":timer")
				})).Return(false).Once()
				mockCache.On("Put", mock.MatchedBy(func(timer string) bool {
					return strings.HasSuffix(timer, ":timer")
				}), now.Timestamp(), time.Duration(1)*time.Minute).Return(errors.New("error")).Once()

				assert.Panics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {},
		},
		{
			name: "error when put key fail in first request",
			setup: func() {
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limit.PerMinute(1),
					}
				}).Once()
				mockConfig.On("GetString", "cache.prefix").Return("goravel").Once()
				mockCache.On("Has", mock.MatchedBy(func(timer string) bool {
					return strings.HasSuffix(timer, ":timer")
				})).Return(false).Once()
				mockCache.On("Put", mock.MatchedBy(func(timer string) bool {
					return strings.HasSuffix(timer, ":timer")
				}), now.Timestamp(), time.Duration(1)*time.Minute).Return(nil).Once()
				mockCache.On("Put", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "goravel:throttle:test:")
				}), 1, time.Duration(1)*time.Minute).Return(errors.New("error")).Once()

				assert.Panics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {},
		},
		{
			name: "success when not over MaxAttempts",
			setup: func() {
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limit.PerMinute(2),
					}
				}).Once()
				mockConfig.On("GetString", "cache.prefix").Return("goravel").Once()
				mockCache.On("Has", mock.MatchedBy(func(timer string) bool {
					return strings.HasSuffix(timer, ":timer")
				})).Return(true).Once()
				mockCache.On("GetInt", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "goravel:throttle:test:")
				}), 0).Return(1).Once()
				mockCache.On("Increment", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "goravel:throttle:test:")
				})).Return(2, nil).Once()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Equal(t, "2", ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Equal(t, "0", ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
		{
			name: "success when over MaxAttempts",
			setup: func() {
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limit.PerMinute(2),
					}
				}).Once()
				mockConfig.On("GetString", "cache.prefix").Return("goravel").Once()
				mockCache.On("Has", mock.MatchedBy(func(timer string) bool {
					return strings.HasSuffix(timer, ":timer")
				})).Return(true).Once()
				mockCache.On("GetInt", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "goravel:throttle:test:")
				}), 0).Return(2).Once()
				mockCache.On("GetInt", mock.MatchedBy(func(timer string) bool {
					return strings.HasSuffix(timer, ":timer")
				}), 0).Return(int(now.Timestamp())).Once()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				assert.Equal(t, strconv.FormatInt(now.Timestamp()+60, 10), ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Equal(t, "60", ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx = new(TestContext)
			mockCache = cachemocks.NewCache(t)
			mockConfig = configmocks.NewConfig(t)
			mockRateLimiterFacade = httpmocks.NewRateLimiter(t)
			http.CacheFacade = mockCache
			http.ConfigFacade = mockConfig
			http.RateLimiterFacade = mockRateLimiterFacade
			test.setup()
			test.assert()
		})
	}
}

type TestContext struct {
	response contractshttp.Response
}

func (r *TestContext) Deadline() (deadline time.Time, ok bool) {
	//TODO implement me
	panic("implement me")
}

func (r *TestContext) Done() <-chan struct{} {
	//TODO implement me
	panic("implement me")
}

func (r *TestContext) Err() error {
	//TODO implement me
	panic("implement me")
}

func (r *TestContext) Value(key any) any {
	//TODO implement me
	panic("implement me")
}

func (r *TestContext) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (r *TestContext) WithValue(key string, value any) {
	//TODO implement me
	panic("implement me")
}

func (r *TestContext) Request() contractshttp.Request {
	return new(TestRequest)
}

func (r *TestContext) Response() contractshttp.Response {
	if r.response == nil {
		r.response = &TestResponse{
			Headers: make(map[string]string),
		}
	}

	return r.response
}

type TestRequest struct{}

func (r *TestRequest) Header(key string, defaultValue ...string) string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Headers() nethttp.Header {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Method() string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Path() string {
	return "/test"
}

func (r *TestRequest) Url() string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) FullUrl() string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Ip() string {
	return "127.0.0.1"
}

func (r *TestRequest) Host() string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) All() map[string]any {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Bind(obj any) error {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Route(key string) string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) RouteInt(key string) int {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) RouteInt64(key string) int64 {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Query(key string, defaultValue ...string) string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) QueryInt(key string, defaultValue ...int) int {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) QueryInt64(key string, defaultValue ...int64) int64 {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) QueryBool(key string, defaultValue ...bool) bool {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) QueryArray(key string) []string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) QueryMap(key string) map[string]string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Queries() map[string]string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Form(key string, defaultValue ...string) string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Json(key string, defaultValue ...string) string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Input(key string, defaultValue ...string) string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) InputArray(key string, defaultValue ...[]string) []string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) InputMap(key string, defaultValue ...map[string]string) map[string]string {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) InputInt(key string, defaultValue ...int) int {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) InputInt64(key string, defaultValue ...int64) int64 {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) InputBool(key string, defaultValue ...bool) bool {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) File(name string) (filesystem.File, error) {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) AbortWithStatus(code int) {}

func (r *TestRequest) AbortWithStatusJson(code int, jsonObj any) {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Next() {}

func (r *TestRequest) Origin() *nethttp.Request {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) Validate(rules map[string]string, options ...validation.Option) (validation.Validator, error) {
	//TODO implement me
	panic("implement me")
}

func (r *TestRequest) ValidateRequest(request contractshttp.FormRequest) (validation.Errors, error) {
	//TODO implement me
	panic("implement me")
}

type TestResponse struct {
	Headers map[string]string
}

func (r *TestResponse) Data(code int, contentType string, data []byte) {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) Download(filepath, filename string) {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) File(filepath string) {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) Header(key, value string) contractshttp.Response {
	r.Headers[key] = value

	return r
}

func (r *TestResponse) Json(code int, obj any) {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) Origin() contractshttp.ResponseOrigin {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) Redirect(code int, location string) {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) String(code int, format string, values ...any) {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) Success() contractshttp.ResponseSuccess {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) Status(code int) contractshttp.ResponseStatus {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) Writer() nethttp.ResponseWriter {
	//TODO implement me
	panic("implement me")
}

func (r *TestResponse) Flush() {
	//TODO implement me
	panic("implement me")
}

type TestLimit struct{}

func (r *TestLimit) By(key string) contractshttp.Limit {
	//TODO implement me
	panic("implement me")
}

func (r *TestLimit) Response(f func(ctx contractshttp.Context)) contractshttp.Limit {
	//TODO implement me
	panic("implement me")
}
