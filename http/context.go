package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"

	"github.com/goravel/framework/contracts/http"
)

func Background() http.Context {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	return NewContext(r, w)
}

type Ctx context.Context

// Context represents an HTTP request/response context
type Context struct {
	Ctx
	r *nethttp.Request
	w nethttp.ResponseWriter
}

func NewContext(r *nethttp.Request, w nethttp.ResponseWriter) *Context {
	return &Context{
		Ctx: Ctx(r.Context()),
		r:   r,
		w:   w,
	}
}

func (c *Context) Context() context.Context {
	return c.Ctx
}

func (c *Context) WithContext(ctx context.Context) {
	// Changing the request context to a new context
	c.Ctx = ctx
}

func (c *Context) WithValue(key any, value any) {
	// nolint:all
	c.Ctx = context.WithValue(c.Ctx, key, value)
}

func (c *Context) Request() http.ContextRequest {
	return NewContextRequest(c, LogFacade, ValidationFacade)
}

func (c *Context) Response() http.ContextResponse {
	return NewContextResponse(c.w, c.r, &ResponseOrigin{
		ResponseWriter: c.w,
	})
}
