package client

import (
	"context"
	"io"
	"net/http"
)

type Request interface {
	// Get sends a GET request to the specified URI.
	Get(uri string) (Response, error)
	// Post sends a POST request to the specified URI with the given body.
	Post(uri string, body io.Reader) (Response, error)
	// Put sends a PUT request to the specified URI with the given body.
	Put(uri string, body io.Reader) (Response, error)
	// Delete sends a DELETE request to the specified URI with the given body.
	Delete(uri string, body io.Reader) (Response, error)
	// Patch sends a PATCH request to the specified URI with the given body.
	Patch(uri string, body io.Reader) (Response, error)
	// Head sends a HEAD request to the specified URI.
	Head(uri string) (Response, error)

	// Options sends an OPTIONS request to the specified URI.
	Options(uri string) (Response, error)
	// Accept sets the Accept header to the specified content type.
	Accept(contentType string) Request
	// AcceptJSON sets the Accept header to "application/json".
	AcceptJSON() Request
	// AsForm sets the Content-Type header to "application/x-www-form-urlencoded".
	AsForm() Request
	// BaseUrl overrides the base URL defined in the configuration for this specific request chain.
	//
	// This allows you to hit a different domain than the one configured for the
	// client, useful for dynamic subdomains or runtime overrides.
	BaseUrl(url string) Request
	// Clone creates a deep copy of the request builder.
	// This is useful if you want to reuse a base request with shared headers/tokens
	// for multiple distinct API calls.
	Clone() Request
	// FlushHeaders clears all configured headers.
	FlushHeaders() Request
	// HttpClient returns the underlying standard library *http.Client.
	// Use this for advanced scenarios like injecting the client into third-party SDKs.
	HttpClient() *http.Client
	// ReplaceHeaders replaces all existing headers with the provided map.
	ReplaceHeaders(headers map[string]string) Request
	// WithBasicAuth sets the Authorization header using Basic Auth.
	WithBasicAuth(username, password string) Request
	// WithContext sets the context for the request.
	WithContext(ctx context.Context) Request
	// WithCookies adds the provided cookies to the request.
	WithCookies(cookies []*http.Cookie) Request
	// WithCookie adds a single cookie to the request.
	WithCookie(cookie *http.Cookie) Request
	// WithHeader sets a specific header key to the given value.
	WithHeader(key, value string) Request
	// WithHeaders adds multiple headers to the request.
	WithHeaders(map[string]string) Request
	// WithQueryParameter adds a query parameter to the URL.
	WithQueryParameter(key, value string) Request
	// WithQueryParameters adds multiple query parameters to the URL.
	WithQueryParameters(map[string]string) Request
	// WithQueryString parses and adds a raw query string (e.g., "foo=bar&baz=qux").
	WithQueryString(query string) Request
	// WithoutHeader removes a specific header by key.
	WithoutHeader(key string) Request
	// WithToken sets the Authorization header using a Bearer token.
	// You can optionally specify a custom token type (e.g., "Basic") as the second argument.
	WithToken(token string, ttype ...string) Request
	// WithoutToken removes the Authorization header.
	WithoutToken() Request
	// WithUrlParameter replaces a URL parameter placeholder (e.g., "{id}") with the given value.
	WithUrlParameter(key, value string) Request
	// WithUrlParameters replaces multiple URL parameter placeholders.
	WithUrlParameters(map[string]string) Request
}
