package ai

import "context"

// Provider defines low-level model interactions.
type Provider interface {
	// Prompt executes a non-streaming model request.
	Prompt(ctx context.Context, agent Agent, messages []Message, options ...Option) (Response, error)
}
