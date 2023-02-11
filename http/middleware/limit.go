package middleware

import (
	nethttp "net/http"

	"github.com/spf13/cast"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/http/middleware/limiter"
)

// Limit is a middleware that limits the number of requests
// method can be "IP" or "Route"
// limit can be like:
// * 5 reqs/second: "5-S"
// * 10 reqs/minute: "10-M"
// * 1000 reqs/hour: "1000-H"
// * 2000 reqs/day: "2000-D"
func Limit(method, limit string) contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		switch method {
		case "IP":
			ip := ctx.Request().Ip()
			rate, err := limiter.CheckRate(ctx.(*http.GinContext).Instance(), ip, limit)
			if err != nil {
				ctx.Request().AbortWithStatus(nethttp.StatusInternalServerError)
			}

			// X-RateLimit-Limit Max number of requests allowed
			// X-RateLimit-Remaining Number of requests remaining
			// X-RateLimit-Reset Timestamp when the rate limit will reset
			ctx.Response().Header("X-RateLimit-Limit", cast.ToString(rate.Limit))
			ctx.Response().Header("X-RateLimit-Remaining", cast.ToString(rate.Remaining))
			ctx.Response().Header("X-RateLimit-Reset", cast.ToString(rate.Reset))

			if rate.Reached {
				ctx.Request().AbortWithStatus(nethttp.StatusTooManyRequests)
			}
			ctx.Request().Next()

		case "Route":
			route := limiter.RouteToKeyString(ctx.Request().FullUrl() + "-" + ctx.Request().Ip())
			rate, err := limiter.CheckRate(ctx.(*http.GinContext).Instance(), route, limit)
			if err != nil {
				ctx.Request().AbortWithStatus(nethttp.StatusInternalServerError)
			}

			// X-RateLimit-Limit Max number of requests allowed
			// X-RateLimit-Remaining Number of requests remaining
			// X-RateLimit-Reset Timestamp when the rate limit will reset
			ctx.Response().Header("X-RateLimit-Limit", cast.ToString(rate.Limit))
			ctx.Response().Header("X-RateLimit-Remaining", cast.ToString(rate.Remaining))
			ctx.Response().Header("X-RateLimit-Reset", cast.ToString(rate.Reset))

			if rate.Reached {
				ctx.Request().AbortWithStatus(nethttp.StatusTooManyRequests)
			}
			ctx.Request().Next()

		default:
			// if method is not "IP" or "Route", just skip this middleware
			ctx.Request().Next()
		}
	}
}
