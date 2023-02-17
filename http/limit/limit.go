package limit

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
)

func PerMinute(maxAttempts int) contractshttp.Limit {
	return NewLimit(maxAttempts, 1)
}

func PerMinutes(decayMinutes, maxAttempts int) contractshttp.Limit {
	return NewLimit(maxAttempts, decayMinutes)
}

func PerHour(maxAttempts int) contractshttp.Limit {
	return NewLimit(maxAttempts, 60)
}

func PerHours(decayHours, maxAttempts int) contractshttp.Limit {
	return NewLimit(maxAttempts, 60*decayHours)
}

func PerDay(maxAttempts int) contractshttp.Limit {
	return NewLimit(maxAttempts, 60*24)
}

func PerDays(decayDays, maxAttempts int) contractshttp.Limit {
	return NewLimit(maxAttempts, 60*24*decayDays)
}

type Limit struct {
	// The rate limit signature key.
	Key string
	// The maximum number of attempts allowed within the given number of minutes.
	MaxAttempts int
	// The number of minutes until the rate limit is reset.
	DecayMinutes int
	// The response generator callback.
	ResponseCallback func(ctx contractshttp.Context)
}

func NewLimit(maxAttempts, decayMinutes int) *Limit {
	return &Limit{
		MaxAttempts:  maxAttempts,
		DecayMinutes: decayMinutes,
		ResponseCallback: func(ctx contractshttp.Context) {
			ctx.Request().AbortWithStatus(http.StatusTooManyRequests)
		},
	}
}

func (r *Limit) By(key string) contractshttp.Limit {
	r.Key = key

	return r
}

func (r *Limit) Response(callable func(ctx contractshttp.Context)) contractshttp.Limit {
	r.ResponseCallback = callable

	return r
}
