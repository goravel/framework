package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httplimit "github.com/goravel/framework/http/limit"
	supporttime "github.com/goravel/framework/support/time"
)

func Throttle(name string) http.Middleware {
	return func(ctx http.Context) {
		if limiter := facades.RateLimiter.Limiter(name); limiter != nil {
			if limits := limiter(ctx); len(limits) > 0 {
				for _, limit := range limits {
					if instance, ok := limit.(*httplimit.Limit); ok {
						key, timer := key(ctx, name, instance)
						currentTimes := 1

						if facades.Cache.Has(timer) {
							value := facades.Cache.GetInt(key, 0)
							if value >= instance.MaxAttempts {
								expireSecond := facades.Cache.GetInt(timer, 0) + instance.DecayMinutes*60
								ctx.Response().Header("X-RateLimit-Reset", cast.ToString(expireSecond))
								ctx.Response().Header("Retry-After", cast.ToString(expireSecond-int(supporttime.Now().Unix())))
								if instance.ResponseCallback != nil {
									instance.ResponseCallback(ctx)
									return
								} else {
									ctx.Request().AbortWithStatus(http.StatusTooManyRequests)
									return
								}
							} else {
								var err error
								if currentTimes, err = facades.Cache.Increment(key); err != nil {
									panic(err)
								}
							}
						} else {
							expireMinute := time.Duration(instance.DecayMinutes) * time.Minute

							err := facades.Cache.Put(timer, supporttime.Now().Unix(), expireMinute)
							if err != nil {
								panic(err)
							}

							err = facades.Cache.Put(key, currentTimes, expireMinute)
							if err != nil {
								panic(err)
							}
						}

						// add the headers for the passed request
						ctx.Response().Header("X-RateLimit-Limit", cast.ToString(instance.MaxAttempts))
						ctx.Response().Header("X-RateLimit-Remaining", cast.ToString(instance.MaxAttempts-currentTimes))
					}
				}
			}
		}

		ctx.Request().Next()
	}
}

func key(ctx http.Context, limiter string, limit *httplimit.Limit) (string, string) {
	// if no key is set, use the path and ip address as the default key
	var key, timer string
	prefix := facades.Config.GetString("cache.prefix")
	if len(limit.Key) == 0 {
		hash := md5.Sum([]byte(ctx.Request().Path()))
		key = fmt.Sprintf("%s:throttle:%s:%s:%s", prefix, limiter, hex.EncodeToString(hash[:]), ctx.Request().Ip())
	} else {
		hash := md5.Sum([]byte(limit.Key))
		key = fmt.Sprintf("%s:throttle:%s:%s", prefix, limiter, hex.EncodeToString(hash[:]))
	}
	timer = key + ":timer"

	return key, timer
}
