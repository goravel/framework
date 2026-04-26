package ai

import (
	"context"
	"slices"

	contractsai "github.com/goravel/framework/contracts/ai"
)

var _ contractsai.AI = (*Application)(nil)

type Application struct {
	ctx      context.Context
	config   contractsai.Config
	resolver *ProviderResolver
}

func NewApplication(ctx context.Context, config contractsai.Config) *Application {
	return &Application{
		ctx:      ctx,
		config:   config,
		resolver: NewProviderResolver(config),
	}
}

func (r *Application) Agent(agent contractsai.Agent, options ...contractsai.Option) (contractsai.Conversation, error) {
	opts := &contractsai.Options{}
	for _, option := range options {
		option(opts)
	}

	providerName := opts.Provider
	if providerName == "" {
		providerName = r.config.Default
	}

	provider, err := r.resolver.New(providerName)
	if err != nil {
		return nil, err
	}

	model := opts.Model
	middlewares := append(slices.Clone(agent.Middleware()), opts.Middlewares...)

	return NewConversation(r.ctx, agent, provider, model, middlewares), nil
}

func (r *Application) WithContext(ctx context.Context) contractsai.AI {
	return &Application{
		ctx:      ctx,
		config:   r.config,
		resolver: r.resolver,
	}
}
