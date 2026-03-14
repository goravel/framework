package ai

// Response exposes generated text and provider metadata.
type Response interface {
	Text() string
	Usage() Usage
}

// Usage contains token statistics for a response.
type Usage interface {
	InputTokens() int
	OutputTokens() int
	TotalTokens() int
}
