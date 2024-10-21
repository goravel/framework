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

func (c *Context) WithContext(ctx context.Context) {
	// Changing the request context to a new context
	c.Ctx = ctx
}

func (r *Context) WithValue(key any, value any) {
	//nolint:all
	r.Ctx = context.WithValue(r.Ctx, key, value)
}

func (r *Context) Request() http.ContextRequest {
	return nil
}

func (r *Context) Response() http.ContextResponse {
	return nil
}
