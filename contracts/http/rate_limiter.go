package http

//go:generate mockery --name=RateLimiter
type RateLimiter interface {
	// For Register a new rate limiter.
	For(name string, callback func(ctx Context) Limit)
	// ForWithLimits Register a new rate limiter with limits.
	ForWithLimits(name string, callback func(ctx Context) []Limit)
	// Limiter Get a rate limiter instance by name.
	Limiter(name string) func(ctx Context) []Limit
}

type Limit interface {
	// By Set the signature key name for the rate limiter.
	By(key string) Limit
	// Response Set the response callback that should be used.
	Response(func(ctx Context)) Limit
}
