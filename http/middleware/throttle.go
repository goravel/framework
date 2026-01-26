package middleware

import (
	"fmt"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/http"
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
		logFacade := http.App.MakeLog()
		rateLimiterFacade := http.App.MakeRateLimiter()

		if logFacade == nil {
			panic(errors.LogFacadeNotSet)
		}

		if rateLimiterFacade == nil {
			panic(errors.RateLimiterFacadeNotSet)
		}

		if limiter := rateLimiterFacade.Limiter(name); limiter != nil {
			if limits := limiter(ctx); len(limits) > 0 {
				for index, limit := range limits {
					tokens, remaining, reset, ok, err := limit.GetStore().Take(ctx, key(ctx, limit, name, index))
					if err != nil {
						logFacade.Error(errors.HttpRateLimitFailedToCheckThrottle.Args(err))
						break
					}

					resetTime := carbon.FromTimestampNano(int64(reset)).SetTimezone(carbon.UTC)
					retryAfter := carbon.Now().DiffInSeconds(resetTime)
					ctx.Response().Header(HeaderRateLimitLimit, strconv.FormatUint(tokens, 10))
					ctx.Response().Header(HeaderRateLimitRemaining, strconv.FormatUint(remaining, 10))

					if !ok {
						ctx.Response().Header(HeaderRateLimitReset, strconv.Itoa(int(resetTime.Timestamp())))
						ctx.Response().Header(HeaderRetryAfter, strconv.Itoa(int(retryAfter)))
						response(ctx, limit)
						return
					}
				}
			}
		}

		ctx.Request().Next()
	}
}

func key(ctx httpcontract.Context, limit httpcontract.Limit, name string, index int) string {
	// if no key is set, use the path and ip address as the default key
	limitKey := limit.GetKey()
	if len(limitKey) == 0 {
		if request := ctx.Request(); request != nil {
			return fmt.Sprintf("throttle:%s:%d:%s:%s", name, index, request.Ip(), request.Path())
		}
	}

	return fmt.Sprintf("throttle:%s:%d:%s", name, index, limitKey)
}

func response(ctx httpcontract.Context, limit httpcontract.Limit) {
	if callback := limit.GetResponse(); callback != nil {
		callback(ctx)
	} else {
		if request := ctx.Request(); request != nil {
			request.Abort(httpcontract.StatusTooManyRequests)
		}
	}
}
