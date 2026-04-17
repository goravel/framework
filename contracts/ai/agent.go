package ai

// Agent encapsulates instructions and context.
type Agent interface {
	// Instructions returns the system instructions for the model.
	Instructions() string
	// Messages returns prior conversation messages to include as context.
	Messages() []Message
	// Tools returns the tools the model may invoke during the conversation.
	// Return nil or an empty slice if no tools are available.
	Tools() []Tool
}
