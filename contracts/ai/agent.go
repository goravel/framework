package ai

// Agent encapsulates instructions and context.
type Agent interface {
	// Instructions returns the system instructions for the model.
	Instructions() string
	// Messages returns prior conversation messages to include as context.
	Messages() []Message
}
