package http

import (
	"bytes"
	"net/http"
)

type Json map[string]any

//go:generate mockery --name=Response
type Response interface {
	Data(code int, contentType string, data []byte)
	Download(filepath, filename string)
	File(filepath string)
	Header(key, value string) Response
	Json(code int, obj any)
	Origin() ResponseOrigin
	Redirect(code int, location string)
	String(code int, format string, values ...any)
	Success() ResponseSuccess
	Status(code int) ResponseStatus
	Writer() http.ResponseWriter
}

//go:generate mockery --name=ResponseStatus
type ResponseStatus interface {
	Data(contentType string, data []byte)
	Json(obj any)
	String(format string, values ...any)
}

//go:generate mockery --name=ResponseSuccess
type ResponseSuccess interface {
	Data(contentType string, data []byte)
	Json(obj any)
	String(format string, values ...any)
}

//go:generate mockery --name=ResponseOrigin
type ResponseOrigin interface {
	Body() *bytes.Buffer
	Header() http.Header
	Size() int
	Status() int
}
