package ai

import "context"

// MiddlewareNext executes the next prompt handler in the middleware chain.
type MiddlewareNext func(ctx context.Context, prompt AgentPrompt) (Response, error)

// Middleware intercepts an agent prompt request.
type Middleware interface {
	Handle(ctx context.Context, prompt AgentPrompt, next MiddlewareNext) (Response, error)
}
