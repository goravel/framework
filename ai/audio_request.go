package ai

import (
	"context"
	"time"

	contractsai "github.com/goravel/framework/contracts/ai"
)

type audioRequest struct {
	ctx          context.Context
	app          *Application
	prompt       string
	provider     string
	model        string
	voice        string
	instructions string
	timeout      time.Duration
}

func NewAudioRequest(ctx context.Context, app *Application, prompt string) contractsai.AudioRequest {
	return &audioRequest{
		ctx:    ctx,
		app:    app,
		prompt: prompt,
		voice:  DefaultFemaleVoice,
	}
}

func (r *audioRequest) Model(model string) contractsai.AudioRequest {
	r.model = model
	return r
}

func (r *audioRequest) Provider(provider string) contractsai.AudioRequest {
	r.provider = provider
	return r
}

func (r *audioRequest) Voice(voice string) contractsai.AudioRequest {
	r.voice = voice
	return r
}

func (r *audioRequest) Male() contractsai.AudioRequest {
	r.voice = DefaultMaleVoice
	return r
}

func (r *audioRequest) Female() contractsai.AudioRequest {
	r.voice = DefaultFemaleVoice
	return r
}

func (r *audioRequest) Instructions(instructions string) contractsai.AudioRequest {
	r.instructions = instructions
	return r
}

func (r *audioRequest) Timeout(timeout time.Duration) contractsai.AudioRequest {
	r.timeout = timeout
	return r
}

func (r *audioRequest) Store(disk ...string) (string, error) {
	response, err := r.Generate()
	if err != nil {
		return "", err
	}

	return response.Store(disk...)
}

func (r *audioRequest) StoreAs(path string, disk ...string) (string, error) {
	response, err := r.Generate()
	if err != nil {
		return "", err
	}

	return response.StoreAs(path, disk...)
}

func (r *audioRequest) Generate() (contractsai.AudioResponse, error) {
	options := make([]contractsai.Option, 0, 2)
	if r.provider != "" {
		options = append(options, WithProvider(r.provider))
	}
	if r.model != "" {
		options = append(options, WithModel(r.model))
	}

	return r.app.audio(r.ctx, contractsai.AudioPrompt{
		Prompt:       r.prompt,
		Model:        r.model,
		Voice:        r.voice,
		Instructions: r.instructions,
		Timeout:      r.timeout,
	}, options...)
}
