package client

import "net/http"

type Client interface {
	// NewRequest creates a new request builder.
	NewRequest() Request

	// Do executes a prepared request (advanced users).
	Do(req *http.Request) (*http.Response, error)

	// HTTPClient exposes the underlying *http.Client.
	// Intended for advanced / escape-hatch use cases.
	HTTPClient() *http.Client

	// Config returns the resolved client configuration.
	Config() *Config

	// Name returns the client name (e.g. "default", "payments").
	Name() string
}
