package ai

import (
	"context"
	"time"
)

// ProviderState stores provider-scoped conversation state across prompt calls.
type ProviderState interface {
	Get(key string) any
	Set(key string, value any)
}

// AgentPrompt carries all inputs the provider needs to call the model.
// Agent.Instructions() returns the system prompt; Agent.Messages() returns the runtime conversation history.
type AgentPrompt struct {
	Agent       Agent
	Input       string
	Model       string
	Attachments []Attachment
	// Tools lists the tools available to the model for this request.
	Tools []Tool
	// ProviderState carries provider-scoped state for this conversation.
	ProviderState ProviderState
}

type ImagePrompt struct {
	Prompt      string
	Model       string
	Size        ImageSize
	Quality     ImageQuality
	Attachments []Attachment
	Timeout     time.Duration
}

// Provider defines low-level model interactions (text generation).
// Future: extend with TextProvider, ImageProvider, AudioProvider, etc.
type Provider interface {
	// Prompt executes a non-streaming model request.
	Prompt(ctx context.Context, prompt AgentPrompt) (AgentResponse, error)
	// Stream executes a streaming model request and returns a streamable response.
	Stream(ctx context.Context, prompt AgentPrompt) (StreamableAgentResponse, error)
}

// ImageProvider is implemented by providers that support image generation.
type ImageProvider interface {
	// Image executes an image generation or edit request.
	Image(ctx context.Context, prompt ImagePrompt) (ImageResponse, error)
}

// FileProvider is implemented by providers that support storing files before they are referenced by prompts.
type FileProvider interface {
	// PutFile uploads the given file and returns the provider-managed file handle.
	// Providers may return an ID-only handle with empty MimeType()/Content().
	PutFile(ctx context.Context, file StorableFile) (FileResponse, error)
	// GetFile resolves a previously stored provider-managed file.
	GetFile(ctx context.Context, id string) (FileResponse, error)
	// DeleteFile removes a previously stored provider-managed file.
	DeleteFile(ctx context.Context, id string) error
}
