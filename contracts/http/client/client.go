package client

import (
	"net/http"
)

type Client interface {
	// Name returns the logical identifier of this client (e.g., "goravel", "downstream").
	Name() string

	// Config returns the configuration options used to initialize this client.
	// The returned struct is read-only.
	Config() *Config

	// HTTPClient exposes the underlying standard library *http.Client.
	//
	// Use this only for advanced scenarios, such as:
	//  1. Injecting the client into third-party SDKs.
	//  2. Mocking the RoundTripper in tests.
	//  3. Accessing low-level Transport details.
	//
	// For standard HTTP calls, prefer using NewRequest().
	HTTPClient() *http.Client

	// NewRequest creates a fluent Request builder scoped to this client.
	//
	// The returned Request object is mutable and not thread-safe.
	// It applies the client's BaseUrl and default headers automatically.
	NewRequest() Request
}
