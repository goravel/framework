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
	opts, resolvedProviders, err := r.resolveProviderChain(options)
	if err != nil {
		return nil, err
	}

	model := opts.Model
	middlewares := append(slices.Clone(agent.Middleware()), opts.Middlewares...)

	return NewConversation(r.ctx, agent, newFailoverProvider(resolvedProviders), model, middlewares), nil
}

func (r *Application) Audio(prompt string) contractsai.AudioRequest {
	return NewAudioRequest(r.ctx, r, prompt)
}

func (r *Application) Image(prompt string) contractsai.ImageRequest {
	return NewImageRequest(r.ctx, r, prompt)
}

func (r *Application) Transcription(file contractsai.StorableFile) contractsai.TranscriptionRequest {
	return NewTranscriptionRequest(r.ctx, r, file)
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
	opts, resolvedProviders, err := r.resolveProviderChain(options)
	if err != nil {
		return nil, err
	}
	if prompt.Model == "" {
		prompt.Model = opts.Model
	}

	var lastErr error
	for _, resolvedProvider := range resolvedProviders {
		audioProvider, ok := resolvedProvider.provider.(contractsai.AudioProvider)
		if !ok {
			return nil, errors.AIProviderDoesNotSupportAudio.Args(resolvedProvider.name)
		}

		response, err := audioProvider.Audio(ctx, prompt)
		if err == nil {
			return response, nil
		}
		if !isFailoverError(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, lastErr
}

func (r *Application) image(ctx context.Context, prompt contractsai.ImagePrompt, options ...contractsai.Option) (contractsai.ImageResponse, error) {
	opts, resolvedProviders, err := r.resolveProviderChain(options)
	if err != nil {
		return nil, err
	}
	if prompt.Model == "" {
		prompt.Model = opts.Model
	}

	var lastErr error
	for _, resolvedProvider := range resolvedProviders {
		imageProvider, ok := resolvedProvider.provider.(contractsai.ImageProvider)
		if !ok {
			return nil, errors.AIProviderDoesNotSupportImages.Args(resolvedProvider.name)
		}

		response, err := imageProvider.Image(ctx, prompt)
		if err == nil {
			return response, nil
		}
		if !isFailoverError(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, lastErr
}

func (r *Application) transcription(ctx context.Context, prompt contractsai.TranscriptionPrompt, options ...contractsai.Option) (contractsai.TranscriptionResponse, error) {
	opts, resolvedProviders, err := r.resolveProviderChain(options)
	if err != nil {
		return nil, err
	}
	if prompt.Model == "" {
		prompt.Model = opts.Model
	}

	var lastErr error
	for _, resolvedProvider := range resolvedProviders {
		transcriptionProvider, ok := resolvedProvider.provider.(contractsai.TranscriptionProvider)
		if !ok {
			return nil, errors.AIProviderDoesNotSupportTranscription.Args(resolvedProvider.name)
		}

		response, err := transcriptionProvider.Transcription(ctx, prompt)
		if err == nil {
			return response, nil
		}
		if !isFailoverError(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, lastErr
}

func (r *Application) resolveProvider(options []contractsai.Option) (*contractsai.Options, string, contractsai.Provider, error) {
	opts, resolvedProviders, err := r.resolveProviderChain(options)
	if err != nil {
		return nil, "", nil, err
	}

	return opts, resolvedProviders[0].name, resolvedProviders[0].provider, nil
}

func (r *Application) resolveProviderChain(options []contractsai.Option) (*contractsai.Options, []resolvedProvider, error) {
	opts := &contractsai.Options{}
	for _, option := range options {
		if option != nil {
			option(opts)
		}
	}

	providerNames := opts.ProviderChain
	if len(providerNames) == 0 {
		providerNames = []string{r.config.Default}
	}

	resolvedProviders := make([]resolvedProvider, 0, len(providerNames))
	for _, providerName := range providerNames {
		provider, err := r.resolver.New(providerName)
		if err != nil {
			return nil, nil, err
		}

		resolvedProviders = append(resolvedProviders, resolvedProvider{name: providerName, provider: provider})
	}

	return opts, resolvedProviders, nil
}

func (r *Application) WithContext(ctx context.Context) contractsai.AI {
	return &Application{
		ctx:      ctx,
		config:   r.config,
		resolver: r.resolver,
	}
}
