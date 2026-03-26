package ai

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
)

type testProvider struct {
	id string
}

func (t *testProvider) Prompt(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
	return nil, nil
}

func TestProviderResolver_New(t *testing.T) {
	direct := &testProvider{id: "direct"}

	tests := []struct {
		name         string
		config       contractsai.Config
		providerName string
		wantProvider contractsai.Provider
		wantErr      error
	}{
		{
			name:         "unsupported provider",
			config:       contractsai.Config{Providers: map[string]contractsai.ProviderConfig{}},
			providerName: "missing",
			wantErr:      errors.AIProviderNotSupported.Args("missing"),
		},
		{
			name: "direct provider via instance",
			config: contractsai.Config{Providers: map[string]contractsai.ProviderConfig{
				"direct": {Via: direct},
			}},
			providerName: "direct",
			wantProvider: direct,
		},
		{
			name: "contract not fulfilled",
			config: contractsai.Config{Providers: map[string]contractsai.ProviderConfig{
				"broken": {Via: "not-a-provider"},
			}},
			providerName: "broken",
			wantErr:      errors.AIProviderContractNotFulfilled.Args("broken"),
		},
		{
			name: "factory returns error",
			config: contractsai.Config{Providers: map[string]contractsai.ProviderConfig{
				"factory": {Via: func() (contractsai.Provider, error) {
					return nil, assert.AnError
				}},
			}},
			providerName: "factory",
			wantErr:      assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewProviderResolver(tt.config)

			got, err := resolver.New(tt.providerName)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantProvider, got)
		})
	}
}

func TestProviderResolver_NewCacheBehavior(t *testing.T) {
	tests := []struct {
		name              string
		setup             func() (*ProviderResolver, contractsai.Provider, func() int)
		wantFactoryCalled int
		wantErr           error
	}{
		{
			name: "successful factory provider is cached",
			setup: func() (*ProviderResolver, contractsai.Provider, func() int) {
				factoryCalled := 0
				want := &testProvider{id: "factory"}
				resolver := NewProviderResolver(contractsai.Config{Providers: map[string]contractsai.ProviderConfig{
					"factory": {Via: func() (contractsai.Provider, error) {
						factoryCalled++
						return want, nil
					}},
				}})
				return resolver, want, func() int { return factoryCalled }
			},
			wantFactoryCalled: 1,
		},
		{
			name: "failed factory provider is not cached",
			setup: func() (*ProviderResolver, contractsai.Provider, func() int) {
				factoryCalled := 0
				resolver := NewProviderResolver(contractsai.Config{Providers: map[string]contractsai.ProviderConfig{
					"factory": {Via: func() (contractsai.Provider, error) {
						factoryCalled++
						return nil, assert.AnError
					}},
				}})
				return resolver, nil, func() int { return factoryCalled }
			},
			wantFactoryCalled: 2,
			wantErr:           assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, want, getFactoryCalled := tt.setup()

			first, firstErr := resolver.New("factory")
			second, secondErr := resolver.New("factory")

			assert.Equal(t, tt.wantErr, firstErr)
			assert.Equal(t, tt.wantErr, secondErr)
			assert.Equal(t, want, first)
			assert.Equal(t, want, second)
			assert.Equal(t, tt.wantFactoryCalled, getFactoryCalled())
		})
	}
}

func TestProviderResolver_NewConcurrent(t *testing.T) {
	const goroutines = 50

	var factoryCalled int
	var mu sync.Mutex
	want := &testProvider{id: "concurrent"}

	resolver := NewProviderResolver(contractsai.Config{Providers: map[string]contractsai.ProviderConfig{
		"p": {Via: func() (contractsai.Provider, error) {
			mu.Lock()
			factoryCalled++
			mu.Unlock()
			return want, nil
		}},
	}})

	results := make([]contractsai.Provider, goroutines)
	errs := make([]error, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = resolver.New("p")
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 1, factoryCalled, "factory should be called exactly once")
	for i := range goroutines {
		assert.NoError(t, errs[i])
		assert.Equal(t, want, results[i])
	}
}
