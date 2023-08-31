package http

import (
	"bytes"
	"net/http"
)

type Json map[string]any

//go:generate mockery --name=Response
type Response interface {
	Data(code int, contentType string, data []byte) error
	Download(filepath, filename string) error
	File(filepath string) error
	Header(key, value string) Response
	Json(code int, obj any) error
	Origin() ResponseOrigin
	Redirect(code int, location string) error
	String(code int, format string, values ...any) error
	Success() ResponseSuccess
	Status(code int) ResponseStatus
	View() ResponseView
	Writer() http.ResponseWriter
	Flush()
}

//go:generate mockery --name=ResponseStatus
type ResponseStatus interface {
	Data(contentType string, data []byte) error
	Json(obj any) error
	String(format string, values ...any) error
}

//go:generate mockery --name=ResponseSuccess
type ResponseSuccess interface {
	Data(contentType string, data []byte) error
	Json(obj any) error
	String(format string, values ...any) error
}

//go:generate mockery --name=ResponseOrigin
type ResponseOrigin interface {
	Body() *bytes.Buffer
	Header() http.Header
	Size() int
	Status() int
}

type ResponseView interface {
	Make(view string, data ...any) error
	First(views []string, data ...any) error
}
