package http

import (
	"bytes"
	"net/http"
)

type Json map[string]any

//go:generate mockery --name=Response
type Response interface {
	String(code int, format string, values ...any)
	Json(code int, obj any)
	File(filepath string)
	Download(filepath, filename string)
	Success() ResponseSuccess
	Header(key, value string) Response
	Origin() ResponseOrigin
}

//go:generate mockery --name=ResponseSuccess
type ResponseSuccess interface {
	String(format string, values ...any)
	Json(obj any)
}

//go:generate mockery --name=ResponseOrigin
type ResponseOrigin interface {
	Body() *bytes.Buffer
	Header() http.Header
	Size() int
	Status() int
}
