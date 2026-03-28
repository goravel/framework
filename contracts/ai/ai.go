package ai

import "context"

// AI manages AI drivers and creates conversation sessions.
type AI interface {
	// Agent creates a conversation bound to the resolved driver.
	Agent(agent Agent, options ...Option) (Conversation, error)
	// WithContext returns a new AI instance that carries the provided context for all operations.
	WithContext(ctx context.Context) AI
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
