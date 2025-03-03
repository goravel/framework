package client

import "net/http"

type Response interface {
	Body() (string, error)
	ClientError() bool
	Cookie(name string) *http.Cookie
	Cookies() []*http.Cookie
	Failed() bool
	Header(name string) string
	Headers() http.Header
	Json() (map[string]any, error)
	Redirect() bool
	ServerError() bool
	Status() int
	Successful() bool

	/* status code related methods */

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
