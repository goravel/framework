package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http"
	httplimit "github.com/goravel/framework/http/limit"
	"github.com/goravel/framework/support/carbon"
)

func Throttle(name string) httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		if limiter := http.RateLimiterFacade.Limiter(name); limiter != nil {
			if limits := limiter(ctx); len(limits) > 0 {
				for _, limit := range limits {
					if instance, ok := limit.(*httplimit.Limit); ok {
						key, timer := key(ctx, name, instance)
						currentTimes := 1

						if http.CacheFacade.Has(timer) {
							value := http.CacheFacade.GetInt(key, 0)
							if value >= instance.MaxAttempts {
								expireSecond := http.CacheFacade.GetInt(timer, 0) + instance.DecayMinutes*60
								ctx.Response().Header("X-RateLimit-Reset", strconv.Itoa(expireSecond))
								ctx.Response().Header("Retry-After", strconv.Itoa(expireSecond-int(carbon.Now().Timestamp())))
								if instance.ResponseCallback != nil {
									instance.ResponseCallback(ctx)
									return
								} else {
									ctx.Request().AbortWithStatus(httpcontract.StatusTooManyRequests)
									return
								}
							} else {
								var err error
								if currentTimes, err = http.CacheFacade.Increment(key); err != nil {
									panic(err)
								}
							}
						} else {
							expireMinute := time.Duration(instance.DecayMinutes) * time.Minute

							err := http.CacheFacade.Put(timer, carbon.Now().Timestamp(), expireMinute)
							if err != nil {
								panic(err)
							}

							err = http.CacheFacade.Put(key, currentTimes, expireMinute)
							if err != nil {
								panic(err)
							}
						}

						// add the headers for the passed request
						ctx.Response().Header("X-RateLimit-Limit", strconv.Itoa(instance.MaxAttempts))
						ctx.Response().Header("X-RateLimit-Remaining", strconv.Itoa(instance.MaxAttempts-currentTimes))
					}
				}
			}
		}

		ctx.Request().Next()
	}
}

func key(ctx httpcontract.Context, limiter string, limit *httplimit.Limit) (string, string) {
	// if no key is set, use the path and ip address as the default key
	var key, timer string
	prefix := http.ConfigFacade.GetString("cache.prefix")
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
