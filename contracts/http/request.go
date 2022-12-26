package http

import (
	"net/http"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/validation"
)

//go:generate mockery --name=Request
type Request interface {
	Header(key, defaultValue string) string
	Headers() http.Header
	Method() string
	Path() string
	Url() string
	FullUrl() string
	Ip() string

	//Input Retrieve  an input item from the request: /users/{id}
	Input(key string) string
	// Query Retrieve a query string item form the request: /users?id=1
	Query(key, defaultValue string) string
	// Form Retrieve a form string item form the post: /users POST:id=1
	Form(key, defaultValue string) string
	Bind(obj any) error
	File(name string) (filesystem.File, error)

	AbortWithStatus(code int)
	AbortWithStatusJson(code int, jsonObj interface{})

	Next()
	Origin() *http.Request
	Response() Response

	Validate(rules map[string]string, options ...validation.Option) (validation.Validator, error)
	ValidateRequest(request FormRequest) (validation.Errors, error)
}

type FormRequest interface {
	Authorize(ctx Context) error
	Rules() map[string]string
	Messages() map[string]string
	Attributes() map[string]string
	PrepareForValidation(data validation.Data)
}
