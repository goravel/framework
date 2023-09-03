package http

import (
	"context"
)

type Middleware func(Context)
type HandlerFunc func(Context) Response
type ResourceController interface {
	Index(Context) Response
	Show(Context) Response
	Store(Context) Response
	Update(Context) Response
	Destroy(Context) Response
}

//go:generate mockery --name=Context
type Context interface {
	context.Context
	Context() context.Context
	WithValue(key string, value any)
	Request() ContextRequest
	Response() ContextResponse
}
