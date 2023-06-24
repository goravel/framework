package http

import (
	"context"
)

type Middleware func(Context)
type HandlerFunc func(Context)
type ResourceController interface {
	Index(Context)
	Show(Context)
	Store(Context)
	Update(Context)
	Destroy(Context)
}

//go:generate mockery --name=Context
type Context interface {
	context.Context
	Translation
	Context() context.Context
	WithValue(key string, value any)
	Request() Request
	Response() Response
}

//go:generate mockery --name=Translation
type Translation interface {
	Trans(key string, args ...interface{}) string
	GetLocale() string
	SetLocale(locale string)
}
