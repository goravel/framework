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

	mu      sync.RWMutex
	clients map[string]client.Client
}

func NewFactory(config *FactoryConfig, json foundation.Json) *Factory {
	if config == nil {
		config = &FactoryConfig{
			// Initialize the map so lookups don't crash either
			Clients: make(map[string]client.Config),
		}
	}

	return &Factory{
		config:  config,
		json:    json,
		clients: make(map[string]client.Client),
	}
}

func (r *Factory) Client(name ...string) client.Client {
	key := r.config.Default
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}

	// If the key is still empty, it means:
	//   a) The user called Client() without arguments.
	//   b) The config file does not have a "default_client" key set.
	// We cannot proceed because we don't know which connection to use.
	if key == "" {
		return newClientWithError(errors.HttpClientDefaultNotSet)
	}

	r.mu.RLock()
	c, exists := r.clients[key]
	r.mu.RUnlock()

	if exists {
		return c
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if c, exists = r.clients[key]; exists {
		return c
	}

	cfg, ok := r.config.Clients[key]
	// Returns a "Zombie Client" that holds the error until a request is sent.
	if !ok {
		return newClientWithError(errors.HttpClientConnectionNotFound.Args(key))
	}

	newClient := NewClient(key, &cfg, r.json)
	r.clients[key] = newClient

	return newClient
}

func (r *Factory) Request(name ...string) client.Request {
	return r.Client(name...).NewRequest()
}

//
// Proxy Methods (Backward Compatibility)
// These methods allow the Factory to be used directly as a Request builder.
// They strictly delegate to the default client's Request() method.
//

func (r *Factory) Get(uri string) (client.Response, error) {
	return r.Request().Get(uri)
}

func (r *Factory) Post(uri string, body io.Reader) (client.Response, error) {
	return r.Request().Post(uri, body)
}

func (r *Factory) Put(uri string, body io.Reader) (client.Response, error) {
	return r.Request().Put(uri, body)
}

func (r *Factory) Patch(uri string, body io.Reader) (client.Response, error) {
	return r.Request().Patch(uri, body)
}

func (r *Factory) Delete(uri string, body io.Reader) (client.Response, error) {
	return r.Request().Delete(uri, body)
}

func (r *Factory) Head(uri string) (client.Response, error) {
	return r.Request().Head(uri)
}

func (r *Factory) Options(uri string) (client.Response, error) {
	return r.Request().Options(uri)
}

func (r *Factory) Accept(contentType string) client.Request {
	return r.Request().Accept(contentType)
}

func (r *Factory) AcceptJSON() client.Request {
	return r.Request().AcceptJSON()
}

func (r *Factory) AsForm() client.Request {
	return r.Request().AsForm()
}

// Bind decodes the response body into the given variable.
//
// Deprecated: Do not use this method. It masks HTTP errors (like 500s) by
// parsing the body before checking the status code. Use Response.Bind() instead.
func (r *Factory) Bind(value any) client.Request {
	return r.Request().Bind(value)
}

func (r *Factory) Clone() client.Request {
	return r.Request().Clone()
}

func (r *Factory) FlushHeaders() client.Request {
	return r.Request().FlushHeaders()
}

func (r *Factory) ReplaceHeaders(headers map[string]string) client.Request {
	return r.Request().ReplaceHeaders(headers)
}

func (r *Factory) WithBasicAuth(username, password string) client.Request {
	return r.Request().WithBasicAuth(username, password)
}

func (r *Factory) WithContext(ctx context.Context) client.Request {
	return r.Request().WithContext(ctx)
}

func (r *Factory) WithCookie(cookie *http.Cookie) client.Request {
	return r.Request().WithCookie(cookie)
}

func (r *Factory) WithCookies(cookies []*http.Cookie) client.Request {
	return r.Request().WithCookies(cookies)
}

func (r *Factory) WithHeader(key, value string) client.Request {
	return r.Request().WithHeader(key, value)
}

func (r *Factory) WithHeaders(headers map[string]string) client.Request {
	return r.Request().WithHeaders(headers)
}

func (r *Factory) WithQueryParameter(key, value string) client.Request {
	return r.Request().WithQueryParameter(key, value)
}

func (r *Factory) WithQueryParameters(params map[string]string) client.Request {
	return r.Request().WithQueryParameters(params)
}

func (r *Factory) WithQueryString(query string) client.Request {
	return r.Request().WithQueryString(query)
}

func (r *Factory) WithoutHeader(key string) client.Request {
	return r.Request().WithoutHeader(key)
}

func (r *Factory) WithToken(token string, ttype ...string) client.Request {
	return r.Request().WithToken(token, ttype...)
}

func (r *Factory) WithoutToken() client.Request {
	return r.Request().WithoutToken()
}

func (r *Factory) WithUrlParameter(key, value string) client.Request {
	return r.Request().WithUrlParameter(key, value)
}

func (r *Factory) WithUrlParameters(params map[string]string) client.Request {
	return r.Request().WithUrlParameters(params)
}
