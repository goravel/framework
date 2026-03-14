package ai

import "context"

// Provider defines low-level model interactions.
type Provider interface {
	// Prompt executes a non-streaming model request.
	// TODO: Optimize the parameters when implementing a real provider.
	Prompt(ctx context.Context) (Response, error)
}
