package http

import (
	"net/http"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/validation"
)

//go:generate mockery --name=Request
type Request interface {
	Header(key string, defaultValue ...string) string
	Headers() http.Header
	Method() string
	Path() string
	Url() string
	FullUrl() string
	Ip() string

	// Route Retrieve  an input item from the request: /users/{id}
	Route(key string) string
	RouteInt(key string) int
	RouteInt64(key string) int64
	// Query Retrieve a query string item form the request: /users?id=1
	Query(key string, defaultValue ...string) string
	QueryInt(key string, defaultValue ...int) int
	QueryInt64(key string, defaultValue ...int64) int64
	QueryBool(key string, defaultValue ...bool) bool
	QueryArray(key string) []string
	QueryMap(key string) map[string]string
	// Form Retrieve a form string item from the post: /users POST:id=1
	Form(key string, defaultValue ...string) string
	// Json Retrieve data from the post: /users JSON:{"id": 1}
	Json(key string, defaultValue ...string) string
	// Bind Retrieve JSON and bind to obj
	Bind(obj any) error
	// Input Retrieve data by order: json, form, query, route
	Input(key string, defaultValue ...string) string
	InputInt(key string, defaultValue ...int) int
	InputInt64(key string, defaultValue ...int64) int64
	InputBool(key string, defaultValue ...bool) bool

	File(name string) (filesystem.File, error)

	AbortWithStatus(code int)
	AbortWithStatusJson(code int, jsonObj any)

	Next()
	Origin() *http.Request

	Validate(rules map[string]string, options ...validation.Option) (validation.Validator, error)
	ValidateRequest(request FormRequest) (validation.Errors, error)
}

type FormRequest interface {
	Authorize(ctx Context) error
	Rules(ctx Context) map[string]string
	Messages(ctx Context) map[string]string
	Attributes(ctx Context) map[string]string
	PrepareForValidation(ctx Context, data validation.Data) error
}
