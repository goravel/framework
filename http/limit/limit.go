package limit

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/http"
)

func PerMinute(maxAttempts int) http.Limit {
	return NewLimit(maxAttempts, 1)
}

func PerMinutes(decayMinutes, maxAttempts int) http.Limit {
	return NewLimit(maxAttempts, decayMinutes)
}

func PerHour(maxAttempts int) http.Limit {
	return NewLimit(maxAttempts, 60)
}

func PerHours(decayHours, maxAttempts int) http.Limit {
	return NewLimit(maxAttempts, 60*decayHours)
}

func PerDay(maxAttempts int) http.Limit {
	return NewLimit(maxAttempts, 60*24)
}

func PerDays(decayDays, maxAttempts int) http.Limit {
	return NewLimit(maxAttempts, 60*24*decayDays)
}

type Limit struct {
	// The rate limit signature key.
	Key string
	// The store instance.
	Store Store
	// The response generator callback.
	ResponseCallback func(ctx http.Context)
}

func NewLimit(maxAttempts, decayMinutes int) *Limit {
	instance, err := NewStore(uint64(maxAttempts), time.Duration(decayMinutes)*time.Minute)
	if err != nil {
		panic(fmt.Sprintf("failed to load rate limiter store: %v", err))
	}

	return &Limit{
		Store: instance,
		ResponseCallback: func(ctx http.Context) {
			ctx.Request().AbortWithStatus(http.StatusTooManyRequests)
		},
	}
}

func (r *Limit) By(key string) http.Limit {
	r.Key = key

	return r
}

func (r *Limit) Response(callable func(ctx http.Context)) http.Limit {
	r.ResponseCallback = callable

	return r
}
