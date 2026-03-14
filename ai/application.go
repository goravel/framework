package ai

import (
	"context"
	"sync"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/contracts/config"
)

var _ contractsai.AI = (*Application)(nil)

// Application is the AI manager implementation.
type Application struct {
	defaultDriver string
	config        config.Config
	drivers       map[string]contractsai.Provider
	ctx           context.Context
	provider      string
	model         string
	mu            *sync.RWMutex
}

func NewApplication(config config.Config) *Application {
	app := &Application{
		config:  config,
		drivers: make(map[string]contractsai.Provider),
		ctx:     context.Background(),
		mu:      &sync.RWMutex{},
	}
	if config != nil {
		app.defaultDriver = config.GetString("ai.default")
	}

	return app
}

func (r *Application) WithContext(ctx context.Context) contractsai.AI {
	return &Application{
		ctx:           ctx,
		config:        r.config,
		defaultDriver: r.defaultDriver,
		drivers:       r.drivers,
		provider:      r.provider,
		model:         r.model,
		mu:            r.mu,
	}
}

func (r *Application) WithProvider(provider string) contractsai.AI {
	return &Application{
		ctx:           r.ctx,
		config:        r.config,
		defaultDriver: r.defaultDriver,
		drivers:       r.drivers,
		provider:      provider,
		model:         r.model,
		mu:            r.mu,
	}
}

func (r *Application) WithModel(model string) contractsai.AI {
	return &Application{
		ctx:           r.ctx,
		config:        r.config,
		defaultDriver: r.defaultDriver,
		drivers:       r.drivers,
		provider:      r.provider,
		model:         model,
		mu:            r.mu,
	}
}

func (r *Application) Agent(agent contractsai.Agent, options ...contractsai.Option) (contractsai.Conversation, error) {
	return &conversation{}, nil
}
