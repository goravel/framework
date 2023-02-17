package middleware

import (
	"fmt"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httplimit "github.com/goravel/framework/http/limit"
)

func Throttle(name string) http.Middleware {
	return func(ctx http.Context) {
		if limiter := facades.RateLimiter.Limiter(name); limiter != nil {
			if limits := limiter(ctx); len(limits) > 0 {
				for _, limit := range limits {
					if instance, ok := limit.(*httplimit.Limit); ok {
						fmt.Println(instance.Key, instance.MaxAttempts, instance.DecayMinutes)
						if instance.ResponseCallback != nil {
							instance.ResponseCallback(ctx)
						}
						// TODO Determine whether to pass the Limit check
					}
				}
			}
		}

		ctx.Request().Next()

		// TODO calculate remaining attempts(if needed)
	}
}
