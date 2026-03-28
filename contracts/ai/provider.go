package ai

import "context"

// AgentPrompt carries all inputs the provider needs to call the model.
// Agent.Instructions() returns the system prompt; Agent.Messages() returns the runtime conversation history.
type AgentPrompt struct {
	Agent Agent
	Input string
	Model string
}

// Provider defines low-level model interactions (text generation).
// Future: extend with TextProvider, ImageProvider, AudioProvider, etc.
type Provider interface {
	// Prompt executes a non-streaming model request.
	Prompt(ctx context.Context, prompt AgentPrompt) (Response, error)
}
