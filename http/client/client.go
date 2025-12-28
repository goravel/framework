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
	return NewRequest(r, r.json)
}
