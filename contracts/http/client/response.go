package client

import "net/http"

type Response interface {
	Body() (string, error)
	ClientError() bool
	Cookie(name string) *http.Cookie
	Cookies() []*http.Cookie
	Failed() bool
	Headers() http.Header
	Json() (map[string]any, error)
	Redirect() bool
	ServerError() bool
	Status() int
	Successful() bool
}
