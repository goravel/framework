package limit

import (
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
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
	store contractshttp.Store
	// The response generator callback.
	response func(ctx contractshttp.Context)
	// The rate limit signature key.
	key string
}

func NewLimit(maxAttempts, decayMinutes int) *Limit {
	cacheFacade := http.App.MakeCache()
	jsonFacade := http.App.GetJson()

	if cacheFacade == nil {
		panic(errors.CacheFacadeNotSet)
	}

	if jsonFacade == nil {
		panic(errors.JSONParserNotSet)
	}

	instance := NewStore(cacheFacade, jsonFacade, uint64(maxAttempts), time.Duration(decayMinutes)*time.Minute)
	return &Limit{
		store: instance,
		response: func(ctx contractshttp.Context) {
			ctx.Request().Abort(contractshttp.StatusTooManyRequests)
		},
	}
}

func (r *Limit) By(key string) contractshttp.Limit {
	r.key = key

	return r
}

func (r *Limit) GetKey() string {
	return r.key
}

func (r *Limit) GetResponse() func(ctx contractshttp.Context) {
	return r.response
}

func (r *Limit) GetStore() contractshttp.Store {
	return r.store
}

func (r *Limit) Response(callable func(ctx contractshttp.Context)) contractshttp.Limit {
	r.response = callable

	return r
}
