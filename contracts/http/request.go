package http

import (
	"net/http"

	"github.com/goravel/framework/contracts/filesystem"
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
	Bind(obj interface{}) error
	File(name string) (filesystem.File, error)

	AbortWithStatus(code int)
	AbortWithStatusJson(code int, jsonObj interface{})

	Next()
	Origin() *http.Request
	Response() Response

	//Validate(ctx *gin.Context, request FormRequest) []error
}

type FormRequest interface {
	Messages() map[string]string
	Authorize() bool
}
