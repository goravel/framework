package client

import (
	"net/http"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
)

var _ client.Client = (*Client)(nil)

type Client struct {
	name       string
	config     *client.Config
	json       foundation.Json
	httpClient *http.Client

	// initError stores any error encountered during client creation (e.g., missing configuration).
	//
	// This allows us to support method chaining (Fluent API) without panicking or
	// forcing the user to handle errors during client retrieval. The error is
	// returned lazily when the request is eventually executed.
	initError error
}

func NewClient(name string, cfg *client.Config, json foundation.Json) *Client {
	if cfg == nil {
		// Copy the default value so we don't modify the global DefaultConfig variable
		cp := client.DefaultConfig
		cfg = &cp
	}

	// We cast the global DefaultTransport to a concrete *http.Transport and clone it.
	// This gives us a fresh struct with populated Proxy, TLS, and DialContext settings.
	transport := http.DefaultTransport.(*http.Transport).Clone()

	transport.MaxIdleConns = cfg.MaxIdleConns
	transport.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	transport.MaxConnsPerHost = cfg.MaxConnsPerHost
	transport.IdleConnTimeout = cfg.IdleConnTimeout

	return &Client{
		name:   name,
		config: cfg,
		json:   json,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
	}
}

// newClientWithError creates a client that will fail immediately when used.
//
// This is used internally by the Factory when a requested client (e.g., "github")
// is not found in the configuration. Instead of panicking, we return this
// "zombie" client which returns the error lazily when a request is made.
func newClientWithError(err error) *Client {
	return &Client{
		initError: err,
		// We provide an empty config to ensure internal methods do not panic
		// if they try to access config fields before the error is checked.
		config: &client.Config{},
	}
}

func (r *Client) Name() string {
	return r.name
}

func (r *Client) Config() *client.Config {
	return r.config
}

func (r *Client) HTTPClient() *http.Client {
	return r.httpClient
}

func (r *Client) NewRequest() client.Request {
	req := NewRequest(r, r.json)

	if r.initError != nil {
		req.setClientErr(r.initError)
	}

	return req
}
