package client

import (
	"fmt"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
)

type Factory struct {
	config *FactoryConfig
	json   foundation.Json

	mu      sync.RWMutex
	clients map[string]client.Client
}

func NewFactory(config *FactoryConfig, json foundation.Json) *Factory {
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
		panic("http client: default client is not configured")
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
	if !ok {
		panic(fmt.Sprintf("http client: connection [%s] is not configured", key))
	}

	newClient := NewClient(key, &cfg, r.json)
	r.clients[key] = newClient

	return newClient
}

func (r *Factory) Request(name ...string) client.Request {
	return r.Client(name...).NewRequest()
}
