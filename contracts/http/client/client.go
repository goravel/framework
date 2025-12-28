package client

import "net/http"

type Client interface {
	// Name returns the logical name of this client (for logging/metrics).
	// Example: "default", "payments".
	Name() string

	// Config returns the resolved configuration for this client.
	// The returned Config should be treated as read-only by callers.
	Config() *Config

	// HTTPClient exposes the underlying *http.Client used to execute requests.
	//
	// This is an advanced escape hatch intended for:
	//  - injecting a custom RoundTripper in tests,
	//  - performing low-level operations not covered by the Request API.
	//
	// Prefer using NewRequest() for normal request construction/execution.
	HTTPClient() *http.Client

	// NewRequest returns a fresh request builder bound to this client.
	//
	// The returned Request is mutable and owned by the caller. It must not be
	// reused concurrently across goroutines. Call NewRequest() for each logical
	// request (or Clone() it if you need to branch a base builder).
	NewRequest() Request
}
