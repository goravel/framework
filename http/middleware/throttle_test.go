package middleware

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/http"
	mockhttp "github.com/goravel/framework/mocks/http"
	mocklog "github.com/goravel/framework/mocks/log"
)

type ThrottleTestSuite struct {
	suite.Suite
	mockRateLimiter *mockhttp.RateLimiter
	mockCtx         *mockhttp.Context
	mockRequest     *mockhttp.ContextRequest
	mockResponse    *mockhttp.ContextResponse
	mockLimit       *mockhttp.Limit
	mockStore       *mockhttp.Store
	mockLog         *mocklog.Log

	originalRateLimiter httpcontract.RateLimiter
	originalLog         log.Log
}

func TestThrottleTestSuite(t *testing.T) {
	suite.Run(t, new(ThrottleTestSuite))
}

func (s *ThrottleTestSuite) SetupTest() {
	s.mockRateLimiter = mockhttp.NewRateLimiter(s.T())
	s.mockCtx = mockhttp.NewContext(s.T())
	s.mockRequest = mockhttp.NewContextRequest(s.T())
	s.mockResponse = mockhttp.NewContextResponse(s.T())
	s.mockLimit = mockhttp.NewLimit(s.T())
	s.mockStore = mockhttp.NewStore(s.T())
	s.mockLog = mocklog.NewLog(s.T())

	// Store original facades
	s.originalRateLimiter = http.RateLimiterFacade
	s.originalLog = http.LogFacade

	// Set mocks
	http.RateLimiterFacade = s.mockRateLimiter
	http.LogFacade = s.mockLog
}

func (s *ThrottleTestSuite) TearDownTest() {
	// Restore original facades
	http.RateLimiterFacade = s.originalRateLimiter
	http.LogFacade = s.originalLog
}

func (s *ThrottleTestSuite) TestThrottle_NoLimiterFound() {
	s.mockRateLimiter.EXPECT().Limiter("test").Return(nil).Once()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()

	middleware := Throttle("test")
	middleware(s.mockCtx)
}

func (s *ThrottleTestSuite) TestThrottle_LimiterReturnsEmptyLimits() {
	s.mockRateLimiter.EXPECT().Limiter("test").Return(func(ctx httpcontract.Context) []httpcontract.Limit {
		return []httpcontract.Limit{}
	}).Once()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()

	middleware := Throttle("test")
	middleware(s.mockCtx)
}

func (s *ThrottleTestSuite) TestThrottle_RequestAllowed() {
	resetTime := uint64(time.Now().Add(time.Minute).UnixNano())

	s.mockRateLimiter.EXPECT().Limiter("api").Return(func(c httpcontract.Context) []httpcontract.Limit {
		return []httpcontract.Limit{s.mockLimit}
	}).Once()

	s.mockLimit.EXPECT().GetStore().Return(s.mockStore).Once()
	s.mockLimit.EXPECT().GetKey().Return("").Once()
	s.mockStore.EXPECT().Take(mock.Anything, "throttle:api:0:127.0.0.1:/test").Return(uint64(10), uint64(9), resetTime, true, nil).Once()

	// key() calls Request() once (for nil check, Ip, Path), then Throttle calls Request().Next()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Times(2)
	s.mockRequest.EXPECT().Ip().Return("127.0.0.1").Once()
	s.mockRequest.EXPECT().Path().Return("/test").Once()
	s.mockRequest.EXPECT().Next().Once()

	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Times(2)
	s.mockResponse.EXPECT().Header("X-RateLimit-Limit", "10").Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("X-RateLimit-Remaining", "9").Return(s.mockResponse).Once()

	middleware := Throttle("api")
	middleware(s.mockCtx)
}

func (s *ThrottleTestSuite) TestThrottle_RequestAllowedWithCustomKey() {
	resetTime := uint64(time.Now().Add(time.Minute).UnixNano())

	s.mockRateLimiter.EXPECT().Limiter("api").Return(func(c httpcontract.Context) []httpcontract.Limit {
		return []httpcontract.Limit{s.mockLimit}
	}).Once()

	s.mockLimit.EXPECT().GetStore().Return(s.mockStore).Once()
	s.mockLimit.EXPECT().GetKey().Return("user:123").Once()
	s.mockStore.EXPECT().Take(mock.Anything, "throttle:api:0:user:123").Return(uint64(60), uint64(59), resetTime, true, nil).Once()

	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()

	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Times(2)
	s.mockResponse.EXPECT().Header("X-RateLimit-Limit", "60").Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("X-RateLimit-Remaining", "59").Return(s.mockResponse).Once()

	middleware := Throttle("api")
	middleware(s.mockCtx)
}

func (s *ThrottleTestSuite) TestThrottle_RequestRateLimited() {
	resetTime := uint64(time.Now().Add(time.Minute).UnixNano())

	s.mockRateLimiter.EXPECT().Limiter("api").Return(func(c httpcontract.Context) []httpcontract.Limit {
		return []httpcontract.Limit{s.mockLimit}
	}).Once()

	s.mockLimit.EXPECT().GetStore().Return(s.mockStore).Once()
	s.mockLimit.EXPECT().GetKey().Return("").Once()
	s.mockStore.EXPECT().Take(mock.Anything, "throttle:api:0:127.0.0.1:/test").Return(uint64(10), uint64(0), resetTime, false, nil).Once()

	// key() calls Request() once (for nil check, Ip, Path)
	// response() calls Request() once (for nil check and Abort)
	// Total: 2 calls to Request()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Times(2)
	s.mockRequest.EXPECT().Ip().Return("127.0.0.1").Once()
	s.mockRequest.EXPECT().Path().Return("/test").Once()
	s.mockRequest.EXPECT().Abort(httpcontract.StatusTooManyRequests).Once()

	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Times(4)
	s.mockResponse.EXPECT().Header("X-RateLimit-Limit", "10").Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("X-RateLimit-Remaining", "0").Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("X-RateLimit-Reset", mock.AnythingOfType("string")).Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("Retry-After", mock.AnythingOfType("string")).Return(s.mockResponse).Once()

	s.mockLimit.EXPECT().GetResponse().Return(nil).Once()

	middleware := Throttle("api")
	middleware(s.mockCtx)
}

func (s *ThrottleTestSuite) TestThrottle_RequestRateLimitedWithCustomCallback() {
	resetTime := uint64(time.Now().Add(time.Minute).UnixNano())

	s.mockRateLimiter.EXPECT().Limiter("api").Return(func(c httpcontract.Context) []httpcontract.Limit {
		return []httpcontract.Limit{s.mockLimit}
	}).Once()

	s.mockLimit.EXPECT().GetStore().Return(s.mockStore).Once()
	s.mockLimit.EXPECT().GetKey().Return("").Once()
	s.mockStore.EXPECT().Take(mock.Anything, "throttle:api:0:127.0.0.1:/test").Return(uint64(10), uint64(0), resetTime, false, nil).Once()

	// key() calls Request() once (for nil check, Ip, Path)
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Ip().Return("127.0.0.1").Once()
	s.mockRequest.EXPECT().Path().Return("/test").Once()

	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Times(4)
	s.mockResponse.EXPECT().Header("X-RateLimit-Limit", "10").Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("X-RateLimit-Remaining", "0").Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("X-RateLimit-Reset", mock.AnythingOfType("string")).Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("Retry-After", mock.AnythingOfType("string")).Return(s.mockResponse).Once()

	callbackCalled := false
	customCallback := func(c httpcontract.Context) {
		callbackCalled = true
	}
	s.mockLimit.EXPECT().GetResponse().Return(customCallback).Once()

	middleware := Throttle("api")
	middleware(s.mockCtx)

	s.True(callbackCalled)
}

func (s *ThrottleTestSuite) TestThrottle_StoreTakeError() {
	s.mockRateLimiter.EXPECT().Limiter("api").Return(func(c httpcontract.Context) []httpcontract.Limit {
		return []httpcontract.Limit{s.mockLimit}
	}).Once()

	s.mockLimit.EXPECT().GetStore().Return(s.mockStore).Once()
	s.mockLimit.EXPECT().GetKey().Return("").Once()
	s.mockStore.EXPECT().Take(mock.Anything, "throttle:api:0:127.0.0.1:/test").Return(uint64(0), uint64(0), uint64(0), false, errors.New("store error")).Once()

	// key() calls Request() once, Throttle calls Request().Next()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Times(2)
	s.mockRequest.EXPECT().Ip().Return("127.0.0.1").Once()
	s.mockRequest.EXPECT().Path().Return("/test").Once()
	s.mockRequest.EXPECT().Next().Once()

	s.mockLog.EXPECT().Error(mock.Anything).Once()

	middleware := Throttle("api")
	middleware(s.mockCtx)
}

func (s *ThrottleTestSuite) TestThrottle_MultipleLimits() {
	resetTime := uint64(time.Now().Add(time.Minute).UnixNano())

	mockLimit2 := mockhttp.NewLimit(s.T())
	mockStore2 := mockhttp.NewStore(s.T())

	s.mockRateLimiter.EXPECT().Limiter("api").Return(func(c httpcontract.Context) []httpcontract.Limit {
		return []httpcontract.Limit{s.mockLimit, mockLimit2}
	}).Once()

	// First limit passes (GetKey called once per key() call)
	s.mockLimit.EXPECT().GetStore().Return(s.mockStore).Once()
	s.mockLimit.EXPECT().GetKey().Return("user:1").Once()
	s.mockStore.EXPECT().Take(mock.Anything, "throttle:api:0:user:1").Return(uint64(10), uint64(9), resetTime, true, nil).Once()

	// Second limit passes (GetKey called once per key() call)
	mockLimit2.EXPECT().GetStore().Return(mockStore2).Once()
	mockLimit2.EXPECT().GetKey().Return("ip:127.0.0.1").Once()
	mockStore2.EXPECT().Take(mock.Anything, "throttle:api:1:ip:127.0.0.1").Return(uint64(100), uint64(99), resetTime, true, nil).Once()

	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Next().Once()

	s.mockCtx.EXPECT().Response().Return(s.mockResponse).Times(4)
	s.mockResponse.EXPECT().Header("X-RateLimit-Limit", "10").Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("X-RateLimit-Remaining", "9").Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("X-RateLimit-Limit", "100").Return(s.mockResponse).Once()
	s.mockResponse.EXPECT().Header("X-RateLimit-Remaining", "99").Return(s.mockResponse).Once()

	middleware := Throttle("api")
	middleware(s.mockCtx)
}

type KeyTestSuite struct {
	suite.Suite
	mockCtx     *mockhttp.Context
	mockRequest *mockhttp.ContextRequest
	mockLimit   *mockhttp.Limit
}

func TestKeyTestSuite(t *testing.T) {
	suite.Run(t, new(KeyTestSuite))
}

func (s *KeyTestSuite) SetupTest() {
	s.mockCtx = mockhttp.NewContext(s.T())
	s.mockRequest = mockhttp.NewContextRequest(s.T())
	s.mockLimit = mockhttp.NewLimit(s.T())
}

func (s *KeyTestSuite) TestKey_NoKeySet_UsesIpAndPath() {
	s.mockLimit.EXPECT().GetKey().Return("").Once()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Ip().Return("192.168.1.1").Once()
	s.mockRequest.EXPECT().Path().Return("/users").Once()

	result := key(s.mockCtx, s.mockLimit, "api", 0)
	s.Equal("throttle:api:0:192.168.1.1:/users", result)
}

func (s *KeyTestSuite) TestKey_CustomKeySet() {
	s.mockLimit.EXPECT().GetKey().Return("user:456").Once()

	result := key(s.mockCtx, s.mockLimit, "api", 1)
	s.Equal("throttle:api:1:user:456", result)
}

func (s *KeyTestSuite) TestKey_NilRequest() {
	s.mockLimit.EXPECT().GetKey().Return("").Once()
	s.mockCtx.EXPECT().Request().Return(nil).Once()

	result := key(s.mockCtx, s.mockLimit, "api", 0)
	s.Equal("throttle:api:0:", result)
}

type ResponseTestSuite struct {
	suite.Suite
	mockCtx     *mockhttp.Context
	mockRequest *mockhttp.ContextRequest
	mockLimit   *mockhttp.Limit
}

func TestResponseTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseTestSuite))
}

func (s *ResponseTestSuite) SetupTest() {
	s.mockCtx = mockhttp.NewContext(s.T())
	s.mockRequest = mockhttp.NewContextRequest(s.T())
	s.mockLimit = mockhttp.NewLimit(s.T())
}

func (s *ResponseTestSuite) TestResponse_WithResponseCallback() {
	callbackCalled := false
	callback := func(c httpcontract.Context) {
		callbackCalled = true
	}
	s.mockLimit.EXPECT().GetResponse().Return(callback).Once()

	response(s.mockCtx, s.mockLimit)

	s.True(callbackCalled)
}

func (s *ResponseTestSuite) TestResponse_WithoutResponseCallback_DefaultAbort() {
	s.mockLimit.EXPECT().GetResponse().Return(nil).Once()
	s.mockCtx.EXPECT().Request().Return(s.mockRequest).Once()
	s.mockRequest.EXPECT().Abort(httpcontract.StatusTooManyRequests).Once()

	response(s.mockCtx, s.mockLimit)
}

func (s *ResponseTestSuite) TestResponse_NilRequest() {
	s.mockLimit.EXPECT().GetResponse().Return(nil).Once()
	s.mockCtx.EXPECT().Request().Return(nil).Once()

	// Should not panic
	response(s.mockCtx, s.mockLimit)
}
