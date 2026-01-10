package client

import (
	"io"
	"net/http"
)

type Response interface {
	// Bind unmarshalls the response body into the provided value.
	Bind(value any) error
	// Body returns the response body as a string.
	Body() (string, error)
	// ClientError determines if the response status code is >= 400 and < 500.
	ClientError() bool
	// Cookie retrieves a cookie by name from the response.
	Cookie(name string) *http.Cookie
	// Cookies returns all cookies from the response.
	Cookies() []*http.Cookie
	// Failed determines if the response status code is >= 400.
	Failed() bool
	// Header retrieves the first value of a given header field.
	Header(name string) string
	// Headers returns all response headers.
	Headers() http.Header
	// Json returns the response body parsed as a map[string]any.
	Json() (map[string]any, error)
	// Origin returns the underlying *http.Response instance.
	Origin() *http.Response
	// Redirect determines if the response status code is >= 300 and < 400.
	Redirect() bool
	// ServerError determines if the response status code is >= 500.
	ServerError() bool
	// Status returns the HTTP status code.
	Status() int
	// Stream returns the underlying reader to stream the response body.
	Stream() (io.ReadCloser, error)
	// Successful determines if the response status code is >= 200 and < 300.
	Successful() bool

	/* Status Code Helpers */

	OK() bool                  // 200 OK
	Created() bool             // 201 Created
	Accepted() bool            // 202 Accepted
	NoContent() bool           // 204 No Content
	MovedPermanently() bool    // 301 Moved Permanently
	Found() bool               // 302 Found
	BadRequest() bool          // 400 Bad Request
	Unauthorized() bool        // 401 Unauthorized
	PaymentRequired() bool     // 402 Payment Required
	Forbidden() bool           // 403 Forbidden
	NotFound() bool            // 404 Not Found
	RequestTimeout() bool      // 408 Request Timeout
	Conflict() bool            // 409 Conflict
	UnprocessableEntity() bool // 422 Unprocessable Entity
	TooManyRequests() bool     // 429 Too Many Requests
}
