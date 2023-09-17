package http

import (
	"bytes"
	"net/http"
)

type Json map[string]any

//go:generate mockery --name=Response
type Response interface {
	Render() error
}

//go:generate mockery --name=ContextResponse
type ContextResponse interface {
	// Data write the given data to the response.
	Data(code int, contentType string, data []byte) Response
	// Download initiates a file download by specifying the file path and the desired filename
	Download(filepath, filename string) Response
	// File serves a file located at the specified file path as the response.
	File(filepath string) Response
	// Header sets an HTTP header field with the given key and value.
	Header(key, value string) ContextResponse
	// Json sends a JSON response with the specified status code and data object.
	Json(code int, obj any) Response
	// Origin returns the ResponseOrigin
	Origin() ResponseOrigin
	// Redirect performs an HTTP redirect to the specified location with the given status code.
	Redirect(code int, location string) Response
	// String writes a string response with the specified status code and format.
	// The 'values' parameter can be used to replace placeholders in the format string.
	String(code int, format string, values ...any) Response
	// Success returns ResponseSuccess
	Success() ResponseSuccess
	// Status sets the HTTP response status code and returns the ResponseStatus.
	Status(code int) ResponseStatus
	// View returns ResponseView
	View() ResponseView
	// Writer returns the underlying http.ResponseWriter associated with the response.
	Writer() http.ResponseWriter
	// Flush flushes any buffered data to the client.
	Flush()
}

//go:generate mockery --name=ResponseStatus
type ResponseStatus interface {
	// Data write the given data to the Response.
	Data(contentType string, data []byte) Response
	// Json sends a JSON Response with the specified data object.
	Json(obj any) Response
	// String writes a string Response with the specified format and values.
	String(format string, values ...any) Response
}

//go:generate mockery --name=ResponseSuccess
type ResponseSuccess interface {
	// Data write the given data to the Response.
	Data(contentType string, data []byte) Response
	// Json sends a JSON Response with the specified data object.
	Json(obj any) Response
	// String writes a string Response with the specified format and values.
	String(format string, values ...any) Response
}

//go:generate mockery --name=ResponseOrigin
type ResponseOrigin interface {
	// Body returns the response's body content as a *bytes.Buffer.
	Body() *bytes.Buffer
	// Header returns the response's HTTP header.
	Header() http.Header
	// Size returns the size, in bytes, of the response's body content.
	Size() int
	// Status returns the HTTP status code of the response.
	Status() int
}

type ResponseView interface {
	// Make generates a Response for the specified view with optional data.
	Make(view string, data ...any) Response
	// First generates a response for the first available view from the provided list.
	First(views []string, data ...any) Response
}
