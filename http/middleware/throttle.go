package middleware

import (
	"fmt"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http"
	httplimit "github.com/goravel/framework/http/limit"
	"github.com/goravel/framework/support/carbon"
)

const (
	// HeaderRateLimitLimit, HeaderRateLimitRemaining, and HeaderRateLimitReset
	// are the recommended return header values from IETF on rate limiting. Reset
	// is in UTC time.
	HeaderRateLimitLimit     = "X-RateLimit-Limit"
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderRateLimitReset     = "X-RateLimit-Reset"

	// HeaderRetryAfter is the header used to indicate when a client should retry
	// requests (when the rate limit expires), in UTC time.
	HeaderRetryAfter = "Retry-After"
)

func Throttle(name string) httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		if limiter := http.RateLimiterFacade.Limiter(name); limiter != nil {
			if limits := limiter(ctx); len(limits) > 0 {
				for index, limit := range limits {
					if instance, exist := limit.(*httplimit.Limit); exist {
						tokens, remaining, reset, ok, _ := instance.Store.Take(ctx, key(ctx, instance, name, index))

						resetTime := carbon.FromTimestampNano(int64(reset)).SetTimezone(carbon.UTC)
						retryAfter := carbon.Now().DiffInSeconds(resetTime)
						ctx.Response().Header(HeaderRateLimitLimit, strconv.FormatUint(tokens, 10))
						ctx.Response().Header(HeaderRateLimitRemaining, strconv.FormatUint(remaining, 10))

						if !ok {
							ctx.Response().Header(HeaderRateLimitReset, strconv.Itoa(int(resetTime.Timestamp())))
							ctx.Response().Header(HeaderRetryAfter, strconv.Itoa(int(retryAfter)))
							response(ctx, instance)
							break
						}
					}
				}
			}
		}

		ctx.Request().Next()
	}
}

func key(ctx httpcontract.Context, limit *httplimit.Limit, name string, index int) string {
	// if no key is set, use the path and ip address as the default key
	if len(limit.Key) == 0 && ctx.Request() != nil {
		return fmt.Sprintf("throttle:%s:%d:%s:%s", name, index, ctx.Request().Ip(), ctx.Request().Path())
	}

	return fmt.Sprintf("throttle:%s:%d:%s", name, index, limit.Key)
}

func response(ctx httpcontract.Context, limit *httplimit.Limit) {
	if limit.ResponseCallback != nil {
		limit.ResponseCallback(ctx)
	} else {
		if ctx.Request() != nil {
			ctx.Request().AbortWithStatus(httpcontract.StatusTooManyRequests)
		}
	}
}
