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
	// Request embeds the client.Request interface.
	// This allows the Factory to act directly as the default client proxy,
	// simplifying usage (e.g., Facades.Http().Get(...) works immediately).
	client.Request

	config *FactoryConfig
	json   foundation.Json

	// clients serves as a thread-safe cache for *http.Client instances.
	// sync.Map is preferred here because keys are written once and read frequently.
	clients sync.Map

	// mu protects the mocking state (mockTransport) and the embedded Request.
	// We use RWMutex to allow concurrent reads in createHTTPClient.
	mu            sync.RWMutex
	mockTransport *FakeTransport

	responseFactory *ResponseFactory
}

func NewFactory(config *FactoryConfig, json foundation.Json) (*Factory, error) {
	if config == nil {
		return nil, errors.HttpClientConfigNotSet
	}

	f := &Factory{
		config:          config,
		json:            json,
		responseFactory: NewResponseFactory(json),
	}

	// Resolve the default client immediately.
	// If the configuration is invalid (e.g., default client not found),
	// we fail fast and return the error to the caller (Service Provider).
	defaultClient, err := f.resolveClient(config.Default)
	if err != nil {
		return nil, err
	}

	cfg := config.Clients[config.Default]
	f.Request = NewRequest(defaultClient, json, cfg.BaseUrl, config.Default)

	return f, nil
}

func (f *Factory) Client(name ...string) client.Request {
	key := f.config.Default
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}

	// If the requested client is the default one,
	// return the embedded instance directly.
	//
	// Even if f.Request contains an error (e.g., configuration missing),
	// we return it as-is. This preserves the "Lazy Error" behavior
	// without re-allocating a new error object every time.
	f.mu.RLock()
	isDefault := key == f.config.Default
	embeddedReq := f.Request
	f.mu.RUnlock()

	if isDefault && embeddedReq != nil {
		return embeddedReq
	}

	httpClient, err := f.resolveClient(key)
	if err != nil {
		return newRequestWithError(err)
	}

	cfg := f.config.Clients[key]
	return NewRequest(httpClient, f.json, cfg.BaseUrl, key)
}

func (f *Factory) resolveClient(name string) (*http.Client, error) {
	if name == "" {
		return nil, errors.HttpClientDefaultNotSet
	}

	if val, ok := f.clients.Load(name); ok {
		return val.(*http.Client), nil
	}

	cfg, ok := f.config.Clients[name]
	if !ok {
		return nil, errors.HttpClientConnectionNotFound.Args(name)
	}

	newClient := f.createHTTPClient(&cfg)

	// LoadOrStore handles the race condition atomically.
	// If another goroutine created the client while we were working, actual will be theirs.
	actual, _ := f.clients.LoadOrStore(name, newClient)

	return actual.(*http.Client), nil
}

func (f *Factory) createHTTPClient(cfg *Config) *http.Client {
	f.mu.RLock()
	transport := f.mockTransport
	f.mu.RUnlock()

	// If the factory is in "Fake" mode, we inject the mock transport.
	// This bypasses the network entirely and uses the defined stubs.
	if transport != nil {
		return &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		}
	}

	// Clone the default transport to ensure strict isolation between clients.
	// This prevents shared state (like global timeouts) from leaking between instances.
	stdTransport := http.DefaultTransport.(*http.Transport).Clone()

	stdTransport.MaxIdleConns = cfg.MaxIdleConns
	stdTransport.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	stdTransport.MaxConnsPerHost = cfg.MaxConnsPerHost
	stdTransport.IdleConnTimeout = cfg.IdleConnTimeout

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: stdTransport,
	}
}

func (f *Factory) Fake(mocks map[string]any) {
	convertedMocks := make(map[string]func(client.Request) client.Response)

	for pattern, value := range mocks {
		var handler func(client.Request) client.Response

		switch v := value.(type) {
		case func(client.Request) client.Response:
			handler = v
		case client.Response:
			handler = func(_ client.Request) client.Response { return v }
		case string:
			handler = func(_ client.Request) client.Response { return f.Response().String(v, 200) }
		case int:
			handler = func(_ client.Request) client.Response { return f.Response().Status(v) }
		case *ResponseSequence:
			handler = func(_ client.Request) client.Response { return v.getNext() }
		}

		if handler != nil {
			convertedMocks[pattern] = handler
		}
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.mockTransport = NewFakeTransport(f.json, convertedMocks)

	// Clear the client cache safely by iterating keys.
	// This forces subsequent calls to resolveClient to create new http.Client instances,
	// which will now pick up the new mockTransport in createHTTPClient.
	f.clients.Range(func(key, value any) bool {
		f.clients.Delete(key)
		return true
	})

	// Re-initialize the default client immediately so the embedded Request uses the fake.
	// We manually inject the default client name into the context here since we are
	// replacing the embedded instance directly.
	if httpClient, err := f.resolveClient(f.config.Default); err == nil {
		cfg := f.config.Clients[f.config.Default]
		req := NewRequest(httpClient, f.json, cfg.BaseUrl, f.config.Default)
		ctx := context.WithValue(context.Background(), clientNameKey, f.config.Default)
		f.Request = req.WithContext(ctx)
	}
}

func (f *Factory) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.mockTransport = nil

	f.clients.Range(func(key, value any) bool {
		f.clients.Delete(key)
		return true
	})

	if httpClient, err := f.resolveClient(f.config.Default); err == nil {
		cfg := f.config.Clients[f.config.Default]
		f.Request = NewRequest(httpClient, f.json, cfg.BaseUrl, f.config.Default)
	}
}

func (f *Factory) Sequence() client.ResponseSequence {
	return NewResponseSequence(f.responseFactory)
}

func (f *Factory) Response() client.ResponseFactory {
	return f.responseFactory
}

func (f *Factory) AssertSent(assertion func(client.Request) bool) bool {
	f.mu.RLock()
	tr := f.mockTransport
	f.mu.RUnlock()

	if tr == nil {
		return false
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	for _, req := range tr.recorded {
		if assertion(req) {
			return true
		}
	}

	return false
}

func (f *Factory) AssertSentCount(count int) bool {
	f.mu.RLock()
	tr := f.mockTransport
	f.mu.RUnlock()

	if tr == nil {
		return count == 0
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	return len(tr.recorded) == count
}

func (f *Factory) AssertNotSent(assertion func(client.Request) bool) bool {
	return !f.AssertSent(assertion)
}

func (f *Factory) AssertNothingSent() bool {
	return f.AssertSentCount(0)
}
