package client

import (
	"net/http"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/errors"
)

var _ client.Factory = (*Factory)(nil)

type Factory struct {
	client.Request

	json        foundation.Json
	config      *FactoryConfig
	clients     sync.Map
	activeState *FakeState
	mu          sync.RWMutex
}

func NewFactory(config *FactoryConfig, json foundation.Json) (*Factory, error) {
	if config == nil {
		return nil, errors.HttpClientConfigNotSet
	}

	factory := &Factory{
		config: config,
		json:   json,
	}

	if err := factory.refreshDefaultClient(); err != nil {
		return nil, err
	}

	return factory, nil
}

func (r *Factory) Client(names ...string) client.Request {
	name := r.config.Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}

	if name == r.config.Default && r.Request != nil {
		return r.Request
	}

	r.mu.RLock()
	state := r.activeState
	r.mu.RUnlock()

	httpClient, err := r.resolveClient(name, state)
	if err != nil {
		return newRequestWithError(err)
	}

	cfg := r.config.Clients[name]
	return NewRequest(httpClient, r.json, cfg.BaseUrl, name)
}

func (r *Factory) Fake(mocks map[string]any) client.Factory {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.activeState = NewFakeState(r.json, mocks)

	r.clients.Range(func(key, value any) bool {
		r.clients.Delete(key)
		return true
	})

	_ = r.refreshDefaultClientLocked()
	return r
}

func (r *Factory) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.activeState = nil

	r.clients.Range(func(key, value any) bool {
		r.clients.Delete(key)
		return true
	})

	_ = r.refreshDefaultClientLocked()
}

func (r *Factory) PreventStrayRequests() client.Factory {
	r.mu.RLock()
	if r.activeState != nil {
		defer r.mu.RUnlock()
		r.activeState.PreventStrayRequests()
		return r
	}
	r.mu.RUnlock()

	return r.Fake(nil).PreventStrayRequests()
}

func (r *Factory) AllowStrayRequests(patterns []string) client.Factory {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.activeState != nil {
		r.activeState.AllowStrayRequests(patterns)
	}
	return r
}

func (r *Factory) AssertSent(assertion func(client.Request) bool) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.activeState != nil && r.activeState.AssertSent(assertion)
}

func (r *Factory) AssertSentCount(count int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.activeState != nil {
		return r.activeState.AssertSentCount(count)
	}

	return count == 0
}

func (r *Factory) AssertNotSent(assertion func(client.Request) bool) bool {
	return !r.AssertSent(assertion)
}

func (r *Factory) AssertNothingSent() bool {
	return r.AssertSentCount(0)
}

func (r *Factory) Sequence() client.ResponseSequence {
	return NewResponseSequence(NewResponseFactory(r.json))
}

func (r *Factory) Response() client.ResponseFactory {
	return NewResponseFactory(r.json)
}

func (r *Factory) resolveClient(name string, state *FakeState) (*http.Client, error) {
	if name == "" {
		return nil, errors.HttpClientDefaultNotSet
	}

	if val, ok := r.clients.Load(name); ok {
		return val.(*http.Client), nil
	}

	cfg, ok := r.config.Clients[name]
	if !ok {
		return nil, errors.HttpClientConnectionNotFound.Args(name)
	}

	newClient := r.createHTTPClient(&cfg, state)
	actual, _ := r.clients.LoadOrStore(name, newClient)

	return actual.(*http.Client), nil
}

func (r *Factory) createHTTPClient(cfg *Config, state *FakeState) *http.Client {
	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	baseTransport.MaxIdleConns = cfg.MaxIdleConns
	baseTransport.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	baseTransport.MaxConnsPerHost = cfg.MaxConnsPerHost
	baseTransport.IdleConnTimeout = cfg.IdleConnTimeout

	var transport http.RoundTripper = baseTransport

	if state != nil {
		transport = NewFakeTransport(state, baseTransport, r.json)
	}

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}
}

func (r *Factory) refreshDefaultClientLocked() error {
	name := r.config.Default

	c, err := r.resolveClient(name, r.activeState)
	if err != nil {
		return err
	}

	cfg := r.config.Clients[name]
	r.Request = NewRequest(c, r.json, cfg.BaseUrl, name)

	return nil
}

func (r *Factory) refreshDefaultClient() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.refreshDefaultClientLocked()
}
