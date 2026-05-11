package ai

import (
	"context"
	"time"

	contractsai "github.com/goravel/framework/contracts/ai"
)

type imageRequest struct {
	ctx         context.Context
	app         *Application
	prompt      string
	provider    string
	model       string
	size        contractsai.ImageSize
	quality     contractsai.ImageQuality
	attachments []contractsai.Attachment
	timeout     time.Duration
}

func NewImageRequest(ctx context.Context, app *Application, prompt string) contractsai.ImageRequest {
	return &imageRequest{
		ctx:    ctx,
		app:    app,
		prompt: prompt,
	}
}

func (r *imageRequest) Model(model string) contractsai.ImageRequest {
	r.model = model
	return r
}

func (r *imageRequest) Provider(provider string) contractsai.ImageRequest {
	r.provider = provider
	return r
}

func (r *imageRequest) Square() contractsai.ImageRequest {
	r.size = contractsai.ImageSizeSquare
	return r
}

func (r *imageRequest) Portrait() contractsai.ImageRequest {
	r.size = contractsai.ImageSizePortrait
	return r
}

func (r *imageRequest) Landscape() contractsai.ImageRequest {
	r.size = contractsai.ImageSizeLandscape
	return r
}

func (r *imageRequest) Quality(quality contractsai.ImageQuality) contractsai.ImageRequest {
	r.quality = quality
	return r
}

func (r *imageRequest) Attachments(attachments ...contractsai.Attachment) contractsai.ImageRequest {
	r.attachments = append(r.attachments, filterNilAttachments(attachments)...)
	return r
}

func (r *imageRequest) Timeout(timeout time.Duration) contractsai.ImageRequest {
	r.timeout = timeout
	return r
}

func (r *imageRequest) Store(disk ...string) (string, error) {
	response, err := r.Generate()
	if err != nil {
		return "", err
	}

	return response.Store(disk...)
}

func (r *imageRequest) StoreAs(path string, disk ...string) (string, error) {
	response, err := r.Generate()
	if err != nil {
		return "", err
	}

	return response.StoreAs(path, disk...)
}

func (r *imageRequest) Generate() (contractsai.ImageResponse, error) {
	options := make([]contractsai.Option, 0, 2)
	if r.provider != "" {
		options = append(options, WithProvider(r.provider))
	}
	if r.model != "" {
		options = append(options, WithModel(r.model))
	}

	return r.app.image(r.ctx, contractsai.ImagePrompt{
		Prompt:      r.prompt,
		Model:       r.model,
		Size:        r.size,
		Quality:     r.quality,
		Attachments: filterNilAttachments(r.attachments),
		Timeout:     r.timeout,
	}, options...)
}
