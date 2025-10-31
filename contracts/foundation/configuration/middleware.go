package configuration

import "github.com/goravel/framework/contracts/http"

type Middleware interface {
	Append(middleware ...http.Middleware) Middleware
	GetGlobalMiddleware() []http.Middleware
	Prepend(middleware ...http.Middleware) Middleware
	Use(middleware ...http.Middleware) Middleware
}
