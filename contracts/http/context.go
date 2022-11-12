package http

import (
	"context"
)

type Middleware func(Context)
type HandlerFunc func(Context)

type Context interface {
	context.Context
	Context() context.Context
	WithValue(key string, value interface{})
	Request() Request
	Response() Response
}
