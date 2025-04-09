package http

import (
	"context"
)

type Middleware func(next Handler) Handler

type HandlerFunc func(ctx Context) error

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(ctx Context) error {
	return f(ctx)
}

type Handler interface {
	ServeHTTP(ctx Context) error
}

type ResourceController interface {
	// Index method for controller
	Index(Context) error
	// Show method for controller
	Show(Context) error
	// Store method for controller
	Store(Context) error
	// Update method for controller
	Update(Context) error
	// Destroy method for controller
	Destroy(Context) error
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
