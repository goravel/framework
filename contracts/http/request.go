package http

import (
	"context"
	"net/http"
)

type Middleware func(Request)
type HandlerFunc func(Request)

type Request interface {
	//Input Retrieve  an input item from the request: /users/{id}
	Input(key string) string
	// Query Retrieve a query string item form the request: /users?id=1
	Query(key, defaultValue string) string
	// Form Retrieve a form string item form the post: /users POST:id=1
	Form(key, defaultValue string) string
	Bind(obj interface{}) error
	File(name string) (File, error)
	Header(key, defaultValue string) string
	Headers() http.Header
	WithContext(ctx context.Context) Request
	Context() context.Context
	Method() string
	AbortWithStatus(code int)
	Next()
	Path() string
	Ip() string
	//Validate(ctx *gin.Context, request FormRequest) []error
}

type File interface {
	Store(dst string) error
}
