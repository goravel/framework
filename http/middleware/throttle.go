package middleware

import (
	"fmt"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
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
	return func(next httpcontract.Handler) httpcontract.Handler {
		return httpcontract.HandleFunc(func(ctx httpcontract.Context) httpcontract.Response {
			if limiter := http.RateLimiterFacade.Limiter(name); limiter != nil {
				if limits := limiter(ctx); len(limits) > 0 {
					for index, limit := range limits {
						// TODO: We should not use the limit instance directly, but use the contract instead, it's very hard to test currently.
						// Add test cases after optimizing the logic: https://github.com/goravel/goravel/issues/629
						if instance, exist := limit.(*httplimit.Limit); exist {
							tokens, remaining, reset, ok, err := instance.Store.Take(ctx, key(ctx, instance, name, index))
							if err != nil {
								http.LogFacade.Error(errors.HttpRateLimitFailedToCheckThrottle.Args(err))
								break
							}

							resetTime := carbon.FromTimestampNano(int64(reset)).SetTimezone(carbon.UTC)
							retryAfter := carbon.Now().DiffInSeconds(resetTime)
							ctx.Response().Header(HeaderRateLimitLimit, strconv.FormatUint(tokens, 10))
							ctx.Response().Header(HeaderRateLimitRemaining, strconv.FormatUint(remaining, 10))

							if !ok {
								ctx.Response().Header(HeaderRateLimitReset, strconv.Itoa(int(resetTime.Timestamp())))
								ctx.Response().Header(HeaderRetryAfter, strconv.Itoa(int(retryAfter)))
								return response(ctx, instance)
							}
						}
					}
				}
			}

			return next.ServeHTTP(ctx)
		})
	}
}

func key(ctx httpcontract.Context, limit *httplimit.Limit, name string, index int) string {
	// if no key is set, use the path and ip address as the default key
	if len(limit.Key) == 0 && ctx.Request() != nil {
		return fmt.Sprintf("throttle:%s:%d:%s:%s", name, index, ctx.Request().Ip(), ctx.Request().Path())
	}

	return fmt.Sprintf("throttle:%s:%d:%s", name, index, limit.Key)
}

func response(ctx httpcontract.Context, limit *httplimit.Limit) httpcontract.Response {
	if limit.ResponseCallback != nil {
		limit.ResponseCallback(ctx)
	}
	return ctx.Response().Status(httpcontract.StatusTooManyRequests).String(httpcontract.StatusText(httpcontract.StatusTooManyRequests))
}
