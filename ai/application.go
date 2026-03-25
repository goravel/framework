package ai

import (
	"context"

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
	opts := make(map[string]any)
	for _, option := range options {
		option(opts)
	}

	providerName, _ := opts[contractsai.OptionProvider].(string)
	if providerName == "" {
		providerName = r.config.Default
	}

	provider, err := r.resolver.New(providerName)
	if err != nil {
		return nil, err
	}

	model, _ := opts[contractsai.OptionModel].(string)

	return NewConversation(r.ctx, agent, provider, model), nil
}

func (r *Application) WithContext(ctx context.Context) contractsai.AI {
	return &Application{
		ctx:      ctx,
		config:   r.config,
		resolver: r.resolver,
	}
}
