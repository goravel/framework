package http

//go:generate mockery --name=RateLimiter
type RateLimiter interface {
	// For register a new rate limiter.
	For(name string, callback func(ctx Context) Limit)
	// ForWithLimits register a new rate limiter with limits.
	ForWithLimits(name string, callback func(ctx Context) []Limit)
	// Limiter get a rate limiter instance by name.
	Limiter(name string) func(ctx Context) []Limit
}

type Limit interface {
	// By set the signature key name for the rate limiter.
	By(key string) Limit
	// Response set the response callback that should be used.
	Response(func(ctx Context)) Limit
}
