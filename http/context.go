package http

import (
	"context"

	"github.com/goravel/framework/contracts/http"
)

func Background() http.Context {
	return NewContext()
}

type Ctx context.Context

type Context struct {
	Ctx
}

func NewContext() *Context {
	return &Context{
		Ctx: context.Background(),
	}
}

func (r *Context) Context() context.Context {
	return r.Ctx
}

func (r *Context) WithValue(key string, value any) {
	//nolint:all
	r.Ctx = context.WithValue(r.Ctx, key, value)
}

func (r *Context) Request() http.ContextRequest {
	return nil
}

func (r *Context) Response() http.ContextResponse {
	return nil
}
