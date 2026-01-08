package client

import (
	"context"
	"net/http"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/errors"
)

var _ client.Factory = (*Factory)(nil)

type Factory struct {
	client.Request
	json    foundation.Json
	config  *FactoryConfig
	clients sync.Map
	mu      sync.Mutex
	mock    *FakeTransport
}

func NewFactory(cfg *FactoryConfig, j foundation.Json) (*Factory, error) {
	if cfg == nil {
		return nil, errors.HttpClientConfigNotSet
	}

	f := &Factory{
		config: cfg,
		json:   j,
	}

	if err := f.refreshDefaultClient(); err != nil {
		return nil, err
	}

	return f, nil
}

func (r *Factory) Client(names ...string) client.Request {
	name := r.config.Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}

	if val, ok := r.clients.Load(name); ok {
		return NewRequest(val.(*http.Client), r.json, r.config.Clients[name].BaseUrl, name)
	}

	c, err := r.createAndCache(name)
	if err != nil {
		return newRequestWithError(err)
	}

	return NewRequest(c, r.json, r.config.Clients[name].BaseUrl, name)
}

func (r *Factory) Fake(mocks map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mock = NewFakeTransport(r.json, mocks)

	r.clients.Range(func(key, value any) bool {
		r.clients.Delete(key)
		return true
	})

	_ = r.refreshDefaultClientLocked()
}

func (r *Factory) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mock = nil

	r.clients.Range(func(key, value any) bool {
		r.clients.Delete(key)
		return true
	})

	_ = r.refreshDefaultClientLocked()
}

func (r *Factory) Sequence() client.ResponseSequence {
	return NewResponseSequence(NewResponseFactory(r.json))
}

func (r *Factory) Response() client.ResponseFactory {
	return NewResponseFactory(r.json)
}

func (r *Factory) AssertSent(assertion func(client.Request) bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.mock != nil {
		return r.mock.AssertSent(assertion)
	}

	return false
}

func (r *Factory) AssertSentCount(count int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.mock != nil {
		return r.mock.AssertSentCount(count)
	}

	return count == 0
}

func (r *Factory) AssertNotSent(assertion func(client.Request) bool) bool {
	return !r.AssertSent(assertion)
}

func (r *Factory) AssertNothingSent() bool {
	return r.AssertSentCount(0)
}

func (r *Factory) createAndCache(name string) (*http.Client, error) {
	cfg, ok := r.config.Clients[name]
	if !ok {
		return nil, errors.HttpClientConnectionNotFound.Args(name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if val, ok := r.clients.Load(name); ok {
		return val.(*http.Client), nil
	}

	c := r.buildClient(&cfg, r.mock)

	r.clients.Store(name, c)
	return c, nil
}

func (r *Factory) buildClient(cfg *Config, mock *FakeTransport) *http.Client {
	if mock != nil {
		return &http.Client{Timeout: cfg.Timeout, Transport: mock}
	}

	base := http.DefaultTransport.(*http.Transport).Clone()
	base.MaxIdleConns = cfg.MaxIdleConns
	base.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	base.MaxConnsPerHost = cfg.MaxConnsPerHost
	base.IdleConnTimeout = cfg.IdleConnTimeout

	return &http.Client{Timeout: cfg.Timeout, Transport: base}
}

func (r *Factory) refreshDefaultClientLocked() error {
	name := r.config.Default
	cfg, ok := r.config.Clients[name]
	if !ok {
		return errors.HttpClientDefaultNotSet
	}

	c := r.buildClient(&cfg, r.mock)
	r.clients.Store(name, c)

	ctx := context.WithValue(context.Background(), clientNameKey, name)
	r.Request = NewRequest(c, r.json, cfg.BaseUrl, name).WithContext(ctx)

	return nil
}

func (r *Factory) refreshDefaultClient() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.refreshDefaultClientLocked()
}
