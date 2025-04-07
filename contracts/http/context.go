package http

import (
	"context"
)

// Deprecated: Use Handler instead.
type Middleware = Handler

type HandlerFunc func(ctx Context) Response

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(ctx Context) Response {
	return f(ctx)
}

type Handler interface {
	ServeHTTP(ctx Context) Response
}

type ResourceController interface {
	// Index method for controller
	Index(Context) Response
	// Show method for controller
	Show(Context) Response
	// Store method for controller
	Store(Context) Response
	// Update method for controller
	Update(Context) Response
	// Destroy method for controller
	Destroy(Context) Response
}

type Context interface {
	context.Context
	// Context returns the Context
	Context() context.Context
	// WithContext adds a new context to an existing one
	WithContext(ctx context.Context)
	// WithValue add value associated with key in context
	WithValue(key any, value any)
	// Request returns the ContextRequest
	Request() ContextRequest
	// Response returns the ContextResponse
	Response() ContextResponse
}
