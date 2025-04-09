package limit

import (
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http"
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
	// The store instance.
	Store contractshttp.Store
	// The response generator callback.
	ResponseCallback contractshttp.HandlerFunc
}

func NewLimit(maxAttempts, decayMinutes int) *Limit {
	instance := NewStore(http.CacheFacade, http.JsonFacade, uint64(maxAttempts), time.Duration(decayMinutes)*time.Minute)

	return &Limit{
		Store: instance,
		ResponseCallback: func(ctx contractshttp.Context) error {
			return ctx.Response().Status(contractshttp.StatusTooManyRequests).String(contractshttp.StatusText(contractshttp.StatusTooManyRequests))
		},
	}
}

func (r *Limit) By(key string) contractshttp.Limit {
	r.Key = key

	return r
}

func (r *Limit) Response(callback contractshttp.HandlerFunc) contractshttp.Limit {
	r.ResponseCallback = callback

	return r
}
