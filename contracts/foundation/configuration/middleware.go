package configuration

import "github.com/goravel/framework/contracts/http"

type Middleware interface {
	Append(middleware ...http.Middleware) Middleware
	GetGlobalMiddleware() []http.Middleware
	GetRecover() func(ctx http.Context, err any)
	Prepend(middleware ...http.Middleware) Middleware
	Recover(fn func(ctx http.Context, err any)) Middleware
	Use(middleware ...http.Middleware) Middleware
}
