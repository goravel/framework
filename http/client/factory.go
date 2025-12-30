package client

import (
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/errors"
)

var _ client.Factory = (*Factory)(nil)

type Factory struct {
	config *FactoryConfig
	json   foundation.Json

	// mu guards the clients map to ensure thread safety during lazy initialization.
	mu sync.RWMutex
	// clients stores the standard library *http.Client instances (Connection Pool).
	clients map[string]*http.Client
}

func NewFactory(config *FactoryConfig, json foundation.Json) *Factory {
	if config == nil {
		config = &FactoryConfig{
			Clients: make(map[string]client.Config),
		}
	}

	return &Factory{
		config:  config,
		json:    json,
		clients: make(map[string]*http.Client),
	}
}

// Client switches the context to a specific client configuration.
//
// It resolves the underlying *http.Client from the pool (or creates it)
// and returns a fresh Request builder scoped to that client.
func (r *Factory) Client(name ...string) client.Request {
	key := r.config.Default
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}

	// If the key is empty, it indicates that no name was provided and
	// no "default_client" is defined in the configuration.
	if key == "" {
		return newRequestWithError(errors.HttpClientDefaultNotSet)
	}

	// Check if the client already exists in the pool using a read lock.
	r.mu.RLock()
	httpClient, exists := r.clients[key]
	r.mu.RUnlock()

	if exists {
		// We must fetch the config again to pass it to the Request.
		// Since the client exists, the config MUST exist, so we skip the 'ok' check.
		cfg := r.config.Clients[key]
		return NewRequest(httpClient, r.json, cfg.BaseUrl)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check to ensure another goroutine didn't create it while we waited.
	if httpClient, exists = r.clients[key]; exists {
		cfg := r.config.Clients[key]
		return NewRequest(httpClient, r.json, cfg.BaseUrl)
	}

	cfg, ok := r.config.Clients[key]
	if !ok {
		return newRequestWithError(errors.HttpClientConnectionNotFound.Args(key))
	}

	httpClient = r.createHTTPClient(&cfg)
	r.clients[key] = httpClient

	return NewRequest(httpClient, r.json, cfg.BaseUrl)
}

// createHTTPClient handles the low-level transport creation.
//
// It clones the default transport to ensure strict isolation between clients
// (e.g., preventing one client's timeout settings from affecting another).
func (r *Factory) createHTTPClient(cfg *client.Config) *http.Client {
	// We cast the global DefaultTransport to a concrete *http.Transport and clone it.
	// This gives us a fresh struct with populated Proxy, TLS, and DialContext settings.
	transport := http.DefaultTransport.(*http.Transport).Clone()

	transport.MaxIdleConns = cfg.MaxIdleConns
	transport.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	transport.MaxConnsPerHost = cfg.MaxConnsPerHost
	transport.IdleConnTimeout = cfg.IdleConnTimeout

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}
}

// ----------------------------------------------------------------------
// Proxy Methods
// ----------------------------------------------------------------------
// These methods allow the Factory to be used directly as a Request builder.
// They strictly delegate to the default client's Request methods.

func (r *Factory) Get(uri string) (client.Response, error) {
	return r.Client().Get(uri)
}

func (r *Factory) Post(uri string, body io.Reader) (client.Response, error) {
	return r.Client().Post(uri, body)
}

func (r *Factory) Put(uri string, body io.Reader) (client.Response, error) {
	return r.Client().Put(uri, body)
}

func (r *Factory) Patch(uri string, body io.Reader) (client.Response, error) {
	return r.Client().Patch(uri, body)
}

func (r *Factory) Delete(uri string, body io.Reader) (client.Response, error) {
	return r.Client().Delete(uri, body)
}

func (r *Factory) Head(uri string) (client.Response, error) {
	return r.Client().Head(uri)
}

func (r *Factory) Options(uri string) (client.Response, error) {
	return r.Client().Options(uri)
}

// ----------------------------------------------------------------------
// Configuration Proxy Methods
// ----------------------------------------------------------------------
// IMPORTANT: These return 'client.Request', allowing the chain to continue
// on the specific Request instance, not on the Factory itself.

func (r *Factory) Accept(contentType string) client.Request {
	return r.Client().Accept(contentType)
}

func (r *Factory) AcceptJSON() client.Request {
	return r.Client().AcceptJSON()
}

func (r *Factory) AsForm() client.Request {
	return r.Client().AsForm()
}

func (r *Factory) BaseUrl(url string) client.Request {
	return r.Client().BaseUrl(url)
}

func (r *Factory) HttpClient() *http.Client {
	return r.Client().HttpClient()
}

func (r *Factory) Clone() client.Request {
	return r.Client().Clone()
}

func (r *Factory) FlushHeaders() client.Request {
	return r.Client().FlushHeaders()
}

func (r *Factory) ReplaceHeaders(headers map[string]string) client.Request {
	return r.Client().ReplaceHeaders(headers)
}

func (r *Factory) WithBasicAuth(username, password string) client.Request {
	return r.Client().WithBasicAuth(username, password)
}

func (r *Factory) WithContext(ctx context.Context) client.Request {
	return r.Client().WithContext(ctx)
}

func (r *Factory) WithCookie(cookie *http.Cookie) client.Request {
	return r.Client().WithCookie(cookie)
}

func (r *Factory) WithCookies(cookies []*http.Cookie) client.Request {
	return r.Client().WithCookies(cookies)
}

func (r *Factory) WithHeader(key, value string) client.Request {
	return r.Client().WithHeader(key, value)
}

func (r *Factory) WithHeaders(headers map[string]string) client.Request {
	return r.Client().WithHeaders(headers)
}

func (r *Factory) WithQueryParameter(key, value string) client.Request {
	return r.Client().WithQueryParameter(key, value)
}

func (r *Factory) WithQueryParameters(params map[string]string) client.Request {
	return r.Client().WithQueryParameters(params)
}

func (r *Factory) WithQueryString(query string) client.Request {
	return r.Client().WithQueryString(query)
}

func (r *Factory) WithoutHeader(key string) client.Request {
	return r.Client().WithoutHeader(key)
}

func (r *Factory) WithToken(token string, ttype ...string) client.Request {
	return r.Client().WithToken(token, ttype...)
}

func (r *Factory) WithoutToken() client.Request {
	return r.Client().WithoutToken()
}

func (r *Factory) WithUrlParameter(key, value string) client.Request {
	return r.Client().WithUrlParameter(key, value)
}

func (r *Factory) WithUrlParameters(params map[string]string) client.Request {
	return r.Client().WithUrlParameters(params)
}
