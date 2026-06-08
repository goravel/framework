package ai

import (
	"context"
	"time"

	contractsai "github.com/goravel/framework/contracts/ai"
)

type transcriptionRequest struct {
	ctx       context.Context
	app       *Application
	file      contractsai.StorableFile
	provider  string
	failovers []string
	model     string
	language  string
	diarize   bool
	timeout   time.Duration
}

func NewTranscriptionRequest(ctx context.Context, app *Application, file contractsai.StorableFile) contractsai.TranscriptionRequest {
	return &transcriptionRequest{
		ctx:  ctx,
		app:  app,
		file: file,
	}
}

func (r *transcriptionRequest) Model(model string) contractsai.TranscriptionRequest {
	r.model = model
	return r
}

func (r *transcriptionRequest) Provider(provider string, failovers ...string) contractsai.TranscriptionRequest {
	r.provider = provider
	r.failovers = providerChain("", failovers...)
	return r
}

func (r *transcriptionRequest) Language(language string) contractsai.TranscriptionRequest {
	r.language = language
	return r
}

func (r *transcriptionRequest) Diarize() contractsai.TranscriptionRequest {
	r.diarize = true
	return r
}

func (r *transcriptionRequest) Timeout(timeout time.Duration) contractsai.TranscriptionRequest {
	r.timeout = timeout
	return r
}

func (r *transcriptionRequest) Generate() (contractsai.TranscriptionResponse, error) {
	options := make([]contractsai.Option, 0, 2)
	if r.provider != "" || len(r.failovers) > 0 {
		options = append(options, WithProvider(r.provider, r.failovers...))
	}
	if r.model != "" {
		options = append(options, WithModel(r.model))
	}

	return r.app.transcription(r.ctx, contractsai.TranscriptionPrompt{
		File:     r.file,
		Model:    r.model,
		Language: r.language,
		Diarize:  r.diarize,
		Timeout:  r.timeout,
	}, options...)
}
