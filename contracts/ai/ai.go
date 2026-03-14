package ai

import "context"

// AI manages AI drivers and creates conversation sessions.
type AI interface {
	// WithContext returns a request-scoped AI instance.
	WithContext(ctx context.Context) AI
	// WithProvider sets the provider to use for the conversation.
	WithProvider(provider string) AI
	// WithModel sets the model to use for the conversation.
	WithModel(model string) AI
	// Agent creates a conversation bound to the resolved driver.
	Agent(agent Agent, options ...Option) (Conversation, error)
}

// Conversation is a stateful chat session.
type Conversation interface {
	// Prompt sends a non-streaming input and updates the conversation history.
	Prompt(input string) (Response, error)
	// Messages returns current conversation history.
	Messages() []Message
	// Reset clears runtime history and restores initial agent messages.
	Reset()
}
