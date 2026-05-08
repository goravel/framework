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

func (r *Application) Audio(prompt string) contractsai.AudioRequest {
	return NewAudioRequest(r.ctx, r, prompt)
}

func (r *Application) Image(prompt string) contractsai.ImageRequest {
	return NewImageRequest(r.ctx, r, prompt)
}

func (r *Application) putFile(ctx context.Context, file contractsai.StorableFile, options ...contractsai.Option) (contractsai.FileResponse, error) {
	_, providerName, provider, err := r.resolveProvider(options)
	if err != nil {
		return nil, err
	}

	fileProvider, ok := provider.(contractsai.FileProvider)
	if !ok {
		return nil, errors.AIProviderDoesNotSupportFiles.Args(providerName)
	}

	return fileProvider.PutFile(ctx, file)
}

func (r *Application) getFile(ctx context.Context, id string, options ...contractsai.Option) (contractsai.FileResponse, error) {
	if id == "" {
		return nil, errors.AIStoredFileIDEmpty
	}

	_, providerName, provider, err := r.resolveProvider(options)
	if err != nil {
		return nil, err
	}

	fileProvider, ok := provider.(contractsai.FileProvider)
	if !ok {
		return nil, errors.AIProviderDoesNotSupportFiles.Args(providerName)
	}

	return fileProvider.GetFile(ctx, id)
}

func (r *Application) deleteFile(ctx context.Context, id string, options ...contractsai.Option) error {
	if id == "" {
		return errors.AIStoredFileIDEmpty
	}

	_, providerName, provider, err := r.resolveProvider(options)
	if err != nil {
		return err
	}

	fileProvider, ok := provider.(contractsai.FileProvider)
	if !ok {
		return errors.AIProviderDoesNotSupportFiles.Args(providerName)
	}

	return fileProvider.DeleteFile(ctx, id)
}

func (r *Application) audio(ctx context.Context, prompt contractsai.AudioPrompt, options ...contractsai.Option) (contractsai.AudioResponse, error) {
	opts, providerName, provider, err := r.resolveProvider(options)
	if err != nil {
		return nil, err
	}
	if prompt.Model == "" {
		prompt.Model = opts.Model
	}

	audioProvider, ok := provider.(contractsai.AudioProvider)
	if !ok {
		return nil, errors.AIProviderDoesNotSupportAudio.Args(providerName)
	}

	return audioProvider.Audio(ctx, prompt)
}

func (r *Application) image(ctx context.Context, prompt contractsai.ImagePrompt, options ...contractsai.Option) (contractsai.ImageResponse, error) {
	opts, providerName, provider, err := r.resolveProvider(options)
	if err != nil {
		return nil, err
	}
	if prompt.Model == "" {
		prompt.Model = opts.Model
	}

	imageProvider, ok := provider.(contractsai.ImageProvider)
	if !ok {
		return nil, errors.AIProviderDoesNotSupportImages.Args(providerName)
	}

	return imageProvider.Image(ctx, prompt)
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
