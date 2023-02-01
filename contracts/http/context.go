package http

import (
	"context"
)

type Middleware func(Context)
type HandlerFunc func(Context)

//go:generate mockery --name=Context
type Context interface {
	context.Context
	Context() context.Context
	WithValue(key string, value any)
	Request() Request
	Response() Response
}
