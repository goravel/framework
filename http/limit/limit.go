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
	// The store instance.
	Store contractshttp.Store
	// The response generator callback.
	ResponseCallback func(ctx contractshttp.Context)
	// The rate limit signature key.
	Key string
}

func NewLimit(maxAttempts, decayMinutes int) *Limit {
	instance := NewStore(http.CacheFacade, http.JsonFacade, uint64(maxAttempts), time.Duration(decayMinutes)*time.Minute)

	return &Limit{
		Store: instance,
		ResponseCallback: func(ctx contractshttp.Context) {
			ctx.Request().Abort(contractshttp.StatusTooManyRequests)
		},
	}
}

func (r *Limit) By(key string) contractshttp.Limit {
	r.Key = key

	return r
}

func (r *Limit) GetKey() string {
	return r.Key
}

func (r *Limit) GetResponseCallback() func(ctx contractshttp.Context) {
	return r.ResponseCallback
}

func (r *Limit) GetStore() contractshttp.Store {
	return r.Store
}

func (r *Limit) Response(callable func(ctx contractshttp.Context)) contractshttp.Limit {
	r.ResponseCallback = callable

	return r
}
