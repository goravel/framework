package http

import (
	"bytes"
	"net/http"
)

type Json map[string]any

//go:generate mockery --name=ContextResponse
type Response interface {
	Render()
}

//go:generate mockery --name=ContextResponse
type ContextResponse interface {
	Data(code int, contentType string, data []byte) Response
	Download(filepath, filename string) Response
	File(filepath string) Response
	Header(key, value string) ContextResponse
	Json(code int, obj any) Response
	Origin() ResponseOrigin
	Redirect(code int, location string) Response
	String(code int, format string, values ...any) Response
	Success() ResponseSuccess
	Status(code int) ResponseStatus
	View() ResponseView
	Writer() http.ResponseWriter
	Flush()
}

//go:generate mockery --name=ResponseStatus
type ResponseStatus interface {
	Data(contentType string, data []byte) Response
	Json(obj any) Response
	String(format string, values ...any) Response
}

//go:generate mockery --name=ResponseSuccess
type ResponseSuccess interface {
	Data(contentType string, data []byte) Response
	Json(obj any) Response
	String(format string, values ...any) Response
}

//go:generate mockery --name=ResponseOrigin
type ResponseOrigin interface {
	Body() *bytes.Buffer
	Header() http.Header
	Size() int
	Status() int
}

type ResponseView interface {
	Make(view string, data ...any) Response
	First(views []string, data ...any) Response
}
