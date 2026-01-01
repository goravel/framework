package client

import (
	"net/http"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	httperrors "github.com/goravel/framework/errors"
)

var _ client.Factory = (*Factory)(nil)

type Factory struct {
	// Request embeds the client.Request interface.
	// This allows the Factory to act directly as the default client proxy,
	// simplifying usage (e.g., Facades.Http().Get(...) works immediately).
	client.Request

	config *FactoryConfig
	json   foundation.Json

	// clients serves as a thread-safe, append-only cache for *http.Client instances.
	// sync.Map is preferred here over RWMutex because keys are written once and read frequently.
	clients sync.Map
}

func NewFactory(config *FactoryConfig, json foundation.Json) (*Factory, error) {
	if config == nil {
		return nil, httperrors.HttpClientConfigNotSet
	}

	f := &Factory{
		config: config,
		json:   json,
	}

	// Pre-resolve the default client to ensure immediate availability.
	defaultClient, err := f.resolveClient(config.Default)
	if err != nil {
		// Initialize a "zombie" request that returns the error lazily upon execution.
		f.Request = newRequestWithError(err)
	} else {
		cfg := config.Clients[config.Default]
		f.Request = NewRequest(defaultClient, json, cfg.BaseUrl)
	}

	return f, nil
}

// Client resolves or creates a specific client instance by name.
// If no name is provided, it returns the default client.
// Client switches the context to a specific client configuration.
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
	if key == f.config.Default && f.Request != nil {
		return f.Request
	}

	httpClient, err := f.resolveClient(key)
	if err != nil {
		return newRequestWithError(err)
	}

	cfg := f.config.Clients[key]
	return NewRequest(httpClient, f.json, cfg.BaseUrl)
}

// resolveClient retrieves a cached *http.Client or creates a new one using the Singleton pattern.
func (f *Factory) resolveClient(name string) (*http.Client, error) {
	if name == "" {
		return nil, httperrors.HttpClientDefaultNotSet
	}

	if val, ok := f.clients.Load(name); ok {
		return val.(*http.Client), nil
	}

	cfg, ok := f.config.Clients[name]
	if !ok {
		return nil, httperrors.HttpClientConnectionNotFound.Args(name)
	}

	newClient := f.createHTTPClient(&cfg)

	// LoadOrStore handles the race condition atomically.
	// If another goroutine created the client while we were working, actual will be theirs.
	actual, _ := f.clients.LoadOrStore(name, newClient)

	return actual.(*http.Client), nil
}

// createHTTPClient initializes the low-level transport with isolation settings.
func (f *Factory) createHTTPClient(cfg *Config) *http.Client {
	// Clone the default transport to ensure strict isolation between clients.
	// This prevents shared state (like global timeouts) from leaking between instances.
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
