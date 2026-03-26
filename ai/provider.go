package ai

import (
	"sync"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
)

type ProviderResolver struct {
	config    contractsai.Config
	providers map[string]contractsai.Provider
	mu        sync.RWMutex
}

func NewProviderResolver(config contractsai.Config) *ProviderResolver {
	return &ProviderResolver{
		config:    config,
		providers: make(map[string]contractsai.Provider),
	}
}

func (r *ProviderResolver) New(providerName string) (contractsai.Provider, error) {
	r.mu.RLock()
	if provider, ok := r.providers[providerName]; ok {
		r.mu.RUnlock()
		return provider, nil
	}
	r.mu.RUnlock()

	providerCfg, ok := r.config.Providers[providerName]
	if !ok {
		return nil, errors.AIProviderNotSupported.Args(providerName)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring the write lock to avoid TOCTOU races.
	if provider, ok := r.providers[providerName]; ok {
		return provider, nil
	}

	provider, err := r.resolve(providerName, providerCfg)
	if err != nil {
		return nil, err
	}
	if provider != nil {
		r.providers[providerName] = provider
	}

	return provider, nil
}

func (r *ProviderResolver) resolve(name string, config contractsai.ProviderConfig) (contractsai.Provider, error) {
	if p, ok := config.Via.(contractsai.Provider); ok {
		return p, nil
	}
	if fn, ok := config.Via.(func() (contractsai.Provider, error)); ok {
		return fn()
	}
	return nil, errors.AIProviderContractNotFulfilled.Args(name)
}
