package ai

import (
	"context"
	"slices"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
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
	opts, _, provider, err := r.resolveProvider(options)
	if err != nil {
		return nil, err
	}

	model := opts.Model
	middlewares := append(slices.Clone(agent.Middleware()), opts.Middlewares...)

	return NewConversation(r.ctx, agent, provider, model, middlewares), nil
}

func (r *Application) putFile(file contractsai.StorableFile, options ...contractsai.Option) (contractsai.StoredFileResponse, error) {
	opts, providerName, provider, err := r.resolveProvider(options)
	if err != nil {
		return nil, err
	}

	fileProvider, ok := provider.(contractsai.FileProvider)
	if !ok {
		return nil, errors.AIProviderDoesNotSupportFiles.Args(providerName)
	}

	return fileProvider.PutFile(r.ctx, file, *opts)
}

func (r *Application) resolveProvider(options []contractsai.Option) (*contractsai.Options, string, contractsai.Provider, error) {
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
		return nil, "", nil, err
	}

	return opts, providerName, provider, nil
}

func (r *Application) WithContext(ctx context.Context) contractsai.AI {
	return &Application{
		ctx:      ctx,
		config:   r.config,
		resolver: r.resolver,
	}
}
