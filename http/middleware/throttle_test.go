package middleware

import (
	"context"
	nethttp "net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	contractsession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/http/limit"
	cachemocks "github.com/goravel/framework/mocks/cache"
	httpmocks "github.com/goravel/framework/mocks/http"
	"github.com/goravel/framework/support/carbon"
)

func TestThrottle(t *testing.T) {
	var (
		ctx                   *TestContext
		mockRateLimiterFacade *httpmocks.RateLimiter
		mockCache             *cachemocks.Cache
	)

	now := carbon.Now()
	carbon.SetTestNow(now)
	defer carbon.UnsetTestNow()

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

				mockCache.On("Get", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "throttle:test:0:")
				})).Return(limit.NewBucket(1, time.Minute)).Once()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Equal(t, "1", ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Equal(t, "0", ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
		{
			name: "success when not over MaxAttempts",
			setup: func() {
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limit.PerMinute(2),
					}
				}).Once()

				mockCache.On("Get", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "throttle:test:0:")
				})).Return(limit.NewBucket(2, time.Minute)).Once()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Equal(t, "2", ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Equal(t, "1", ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
		{
			name: "success when over MaxAttempts",
			setup: func() {
				limiter := limit.PerMinute(1)
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limiter,
					}
				}).Twice()

				bucket := limit.NewBucket(1, time.Minute)
				mockCache.On("Get", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "throttle:test:0:")
				})).Return(bucket).Twice()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				assert.Equal(t, strconv.FormatInt(now.Timestamp()+60, 10), ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Equal(t, "60", ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Equal(t, "1", ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Equal(t, "0", ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
		{
			name: "success when multiple limiters case 1",
			setup: func() {
				limiter1 := limit.PerDay(10)
				limiter2 := limit.PerMinute(5)
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limiter1, limiter2,
					}
				}).Twice()

				bucket1 := limit.NewBucket(10, 24*time.Hour)
				mockCache.On("Get", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "throttle:test:0:")
				})).Return(bucket1).Twice()
				bucket2 := limit.NewBucket(5, time.Minute)
				mockCache.On("Get", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "throttle:test:1:")
				})).Return(bucket2).Twice()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Empty(t, ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Equal(t, "5", ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Equal(t, "3", ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
		{
			name: "success when multiple limiters case 2",
			setup: func() {
				limiter1 := limit.PerDay(10)
				limiter2 := limit.PerMinute(1)
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limiter1, limiter2,
					}
				}).Twice()

				bucket1 := limit.NewBucket(10, 24*time.Hour)
				mockCache.On("Get", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "throttle:test:0:")
				})).Return(bucket1).Twice()
				bucket2 := limit.NewBucket(1, time.Minute)
				mockCache.On("Get", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "throttle:test:1:")
				})).Return(bucket2).Twice()

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})

				assert.NotPanics(t, func() {
					Throttle("test")(ctx)
				})
			},
			assert: func() {
				// should return last limiter's reset time (limiter configured to 1 per minute)
				assert.Equal(t, strconv.FormatInt(now.Timestamp()+60, 10), ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Equal(t, "60", ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Equal(t, "1", ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Equal(t, "0", ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
		{
			name: "success when multiple limiters case 3",
			setup: func() {
				limiter1 := limit.PerDay(5)
				limiter2 := limit.PerMinute(1)
				mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
					return []contractshttp.Limit{
						limiter1, limiter2,
					}
				}).Times(10)

				bucket1 := limit.NewBucket(5, 24*time.Hour)
				mockCache.On("Get", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "throttle:test:0:")
				})).Return(bucket1).Times(10)
				bucket2 := limit.NewBucket(1, time.Minute)
				mockCache.On("Get", mock.MatchedBy(func(key string) bool {
					return strings.HasPrefix(key, "throttle:test:1:")
				})).Return(bucket2).Times(5)

				for i := 0; i < 10; i++ {
					assert.NotPanics(t, func() {
						Throttle("test")(ctx)
					})
				}
			},
			assert: func() {
				assert.Equal(t, strconv.FormatInt(now.Timestamp()+86400, 10), ctx.Response().(*TestResponse).Headers["X-RateLimit-Reset"])
				assert.Equal(t, "86400", ctx.Response().(*TestResponse).Headers["Retry-After"])
				assert.Equal(t, "5", ctx.Response().(*TestResponse).Headers["X-RateLimit-Limit"])
				assert.Equal(t, "0", ctx.Response().(*TestResponse).Headers["X-RateLimit-Remaining"])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx = new(TestContext)
			mockCache = &cachemocks.Cache{}
			mockRateLimiterFacade = &httpmocks.RateLimiter{}
			http.CacheFacade = mockCache
			http.RateLimiterFacade = mockRateLimiterFacade
			test.setup()
			test.assert()

			mockCache.AssertExpectations(t)
			mockRateLimiterFacade.AssertExpectations(t)
		})
	}
}

func Benchmark_Throttle(b *testing.B) {
	var (
		ctx                   *TestContext
		mockRateLimiterFacade *httpmocks.RateLimiter
		mockCache             *cachemocks.Cache
	)

	now := carbon.Now()
	carbon.SetTestNow(now)
	defer carbon.UnsetTestNow()

	ctx = new(TestContext)
	mockCache = &cachemocks.Cache{}
	mockRateLimiterFacade = &httpmocks.RateLimiter{}
	http.CacheFacade = mockCache
	http.RateLimiterFacade = mockRateLimiterFacade

	b.Run("WithOneLimiter", func(b *testing.B) {
		mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
			return []contractshttp.Limit{
				limit.PerDay(10),
			}
		}).Times(b.N)
		mockCache.On("Get", mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "throttle:test:0:")
		})).Return(limit.NewBucket(1, time.Minute)).Times(b.N)

		for i := 0; i < b.N; i++ {
			Throttle("test")(ctx)
		}

		mockCache.AssertExpectations(b)
		mockRateLimiterFacade.AssertExpectations(b)
	})

	b.Run("WithTwoLimiters", func(b *testing.B) {
		mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
			return []contractshttp.Limit{
				limit.PerDay(10),
				limit.PerMinute(5),
			}
		}).Times(b.N)
		mockCache.On("Get", mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "throttle:test:0:")
		})).Return(limit.NewBucket(10, 24*time.Hour))
		mockCache.On("Get", mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "throttle:test:1:")
		})).Return(limit.NewBucket(5, time.Minute))

		for i := 0; i < b.N; i++ {
			Throttle("test")(ctx)
		}

		mockRateLimiterFacade.AssertExpectations(b)
	})

	b.Run("WithThreeLimiters", func(b *testing.B) {
		mockRateLimiterFacade.On("Limiter", "test").Return(func(ctx contractshttp.Context) []contractshttp.Limit {
			return []contractshttp.Limit{
				limit.PerMinute(5),
				limit.PerHour(10),
				limit.PerDay(100),
			}
		}).Times(b.N)
		mockCache.On("Get", mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "throttle:test:0:")
		})).Return(limit.NewBucket(10, 24*time.Hour))
		mockCache.On("Get", mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "throttle:test:1:")
		})).Return(limit.NewBucket(5, time.Minute))
		mockCache.On("Get", mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "throttle:test:2:")
		})).Return(limit.NewBucket(1, time.Second))

		for i := 0; i < b.N; i++ {
			Throttle("test")(ctx)
		}

		mockRateLimiterFacade.AssertExpectations(b)
	})
}

type TestContext struct {
	response contractshttp.ContextResponse
}

func (r *TestContext) Deadline() (deadline time.Time, ok bool) {

	panic("do not need to implement it")
}

func (r *TestContext) Done() <-chan struct{} {

	panic("do not need to implement it")
}

func (r *TestContext) Err() error {

	panic("do not need to implement it")
}

func (r *TestContext) Value(key any) any {

	panic("do not need to implement it")
}

func (r *TestContext) Context() context.Context {

	panic("do not need to implement it")
}

func (r *TestContext) WithValue(key any, value any) {
	panic("do not need to implement it")
}

func (r *TestContext) Request() contractshttp.ContextRequest {
	return new(TestRequest)
}

func (r *TestContext) Response() contractshttp.ContextResponse {
	if r.response == nil {
		r.response = &TestResponse{
			Headers: make(map[string]string),
		}
	}

	return r.response
}

type TestRequest struct{}

func (r *TestRequest) Header(key string, defaultValue ...string) string {

	panic("do not need to implement it")
}

func (r *TestRequest) Headers() nethttp.Header {

	panic("do not need to implement it")
}

func (r *TestRequest) Method() string {

	panic("do not need to implement it")
}

func (r *TestRequest) Path() string {
	return "/test"
}

func (r *TestRequest) Url() string {

	panic("do not need to implement it")
}

func (r *TestRequest) FullUrl() string {

	panic("do not need to implement it")
}

func (r *TestRequest) Ip() string {
	return "127.0.0.1"
}

func (r *TestRequest) Host() string {

	panic("do not need to implement it")
}

func (r *TestRequest) All() map[string]any {

	panic("do not need to implement it")
}

func (r *TestRequest) Cookie(key string, defaultValue ...string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) Bind(obj any) error {
	panic("do not need to implement it")
}

func (r *TestRequest) BindQuery(any) error {
	panic("do not need to implement it")
}

func (r *TestRequest) Route(key string) string {

	panic("do not need to implement it")
}

func (r *TestRequest) RouteInt(key string) int {
	panic("do not need to implement it")
}

func (r *TestRequest) RouteInt64(key string) int64 {

	panic("do not need to implement it")
}

func (r *TestRequest) Query(key string, defaultValue ...string) string {

	panic("do not need to implement it")
}

func (r *TestRequest) QueryInt(key string, defaultValue ...int) int {

	panic("do not need to implement it")
}

func (r *TestRequest) QueryInt64(key string, defaultValue ...int64) int64 {

	panic("do not need to implement it")
}

func (r *TestRequest) QueryBool(key string, defaultValue ...bool) bool {

	panic("do not need to implement it")
}

func (r *TestRequest) QueryArray(key string) []string {

	panic("do not need to implement it")
}

func (r *TestRequest) QueryMap(key string) map[string]string {

	panic("do not need to implement it")
}

func (r *TestRequest) Queries() map[string]string {

	panic("do not need to implement it")
}

func (r *TestRequest) HasSession() bool {
	panic("do not need to implement it")
}

func (r *TestRequest) Session() contractsession.Session {
	panic("do not need to implement it")
}

func (r *TestRequest) SetSession(contractsession.Session) contractshttp.ContextRequest {
	panic("do not need to implement it")
}

func (r *TestRequest) Input(key string, defaultValue ...string) string {

	panic("do not need to implement it")
}

func (r *TestRequest) InputArray(key string, defaultValue ...[]string) []string {

	panic("do not need to implement it")
}

func (r *TestRequest) InputMap(key string, defaultValue ...map[string]string) map[string]string {

	panic("do not need to implement it")
}

func (r *TestRequest) InputInt(key string, defaultValue ...int) int {

	panic("do not need to implement it")
}

func (r *TestRequest) InputInt64(key string, defaultValue ...int64) int64 {

	panic("do not need to implement it")
}

func (r *TestRequest) InputBool(key string, defaultValue ...bool) bool {

	panic("do not need to implement it")
}

func (r *TestRequest) File(name string) (filesystem.File, error) {

	panic("do not need to implement it")
}

func (r *TestRequest) AbortWithStatus(code int) {}

func (r *TestRequest) AbortWithStatusJson(code int, jsonObj any) {

	panic("do not need to implement it")
}

func (r *TestRequest) Next() {}

func (r *TestRequest) Origin() *nethttp.Request {

	panic("do not need to implement it")
}

func (r *TestRequest) Validate(rules map[string]string, options ...validation.Option) (validation.Validator, error) {

	panic("do not need to implement it")
}

func (r *TestRequest) ValidateRequest(request contractshttp.FormRequest) (validation.Errors, error) {

	panic("do not need to implement it")
}

type TestResponse struct {
	Headers map[string]string
}

func (r *TestResponse) Cookie(cookie contractshttp.Cookie) contractshttp.ContextResponse {
	panic("do not need to implement it")
}

func (r *TestResponse) Data(code int, contentType string, data []byte) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Download(filepath, filename string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) File(filepath string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Header(key, value string) contractshttp.ContextResponse {
	r.Headers[key] = value

	return r
}

func (r *TestResponse) Json(code int, obj any) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) NoContent(...int) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Origin() contractshttp.ResponseOrigin {
	panic("do not need to implement it")
}

func (r *TestResponse) Redirect(code int, location string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) String(code int, format string, values ...any) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Success() contractshttp.ResponseStatus {
	panic("do not need to implement it")
}

func (r *TestResponse) Status(code int) contractshttp.ResponseStatus {
	panic("do not need to implement it")
}

func (r *TestResponse) Stream(int, func(w contractshttp.StreamWriter) error) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) WithoutCookie(name string) contractshttp.ContextResponse {
	panic("do not need to implement it")
}

func (r *TestResponse) Writer() nethttp.ResponseWriter {
	panic("do not need to implement it")
}

func (r *TestResponse) Flush() {
	panic("do not need to implement it")
}

func (r *TestResponse) View() contractshttp.ResponseView {
	panic("do not need to implement it")
}

type TestLimit struct{}

func (r *TestLimit) By(key string) contractshttp.Limit {
	panic("do not need to implement it")
}

func (r *TestLimit) Response(f func(ctx contractshttp.Context)) contractshttp.Limit {
	panic("do not need to implement it")
}
