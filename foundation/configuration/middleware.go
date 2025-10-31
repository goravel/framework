package configuration

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/contracts/http"
)

type Middleware struct {
	middleware []http.Middleware
}

func NewMiddleware(middleware []http.Middleware) *Middleware {
	return &Middleware{
		middleware: middleware,
	}
}

func (r *Middleware) Append(middleware ...http.Middleware) configuration.Middleware {
	r.middleware = append(r.middleware, middleware...)

	return r
}

func (r *Middleware) GetGlobalMiddleware() []http.Middleware {
	return r.middleware
}

func (r *Middleware) Prepend(middleware ...http.Middleware) configuration.Middleware {
	r.middleware = append(middleware, r.middleware...)

	return r
}

func (r *Middleware) Use(middleware ...http.Middleware) configuration.Middleware {
	r.middleware = middleware

	return r
}
