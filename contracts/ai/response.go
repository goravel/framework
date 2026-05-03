package ai

// Response exposes generated text and provider metadata.
type Response interface {
	Text() string
	Usage() Usage
	// ToolCalls returns any tool invocations the model requested.
	// Returns nil or an empty slice when the model returns plain text.
	ToolCalls() []ToolCall
	// Then registers a callback against the resolved response.
	Then(callback func(Response)) Response
}

// Usage contains token statistics for a response.
type Usage interface {
	Input() int
	Output() int
	Total() int
}
