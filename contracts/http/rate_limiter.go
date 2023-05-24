package http

//go:generate mockery --name=RateLimiter
type RateLimiter interface {
	For(name string, callback func(ctx Context) Limit)
	ForWithLimits(name string, callback func(ctx Context) []Limit)
	Limiter(name string) func(ctx Context) []Limit
}

type Limit interface {
	By(key string) Limit
	Response(func(ctx Context)) Limit
}
