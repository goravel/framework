package ai

import "context"

// Response exposes generated text and provider metadata.
type Response interface {
	Text(ctx context.Context) string
	Usage(ctx context.Context) Usage
	Raw(ctx context.Context) any
}

// Usage contains token statistics for a response.
type Usage interface {
	InputTokens(ctx context.Context) int
	OutputTokens(ctx context.Context) int
	TotalTokens(ctx context.Context) int
}
