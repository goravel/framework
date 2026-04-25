package ai

import "context"

// Next executes the next prompt handler in the middleware chain.
type Next interface {
	Execute(ctx context.Context, prompt AgentPrompt) (Response, error)
}

// Middleware intercepts an agent prompt request.
type Middleware interface {
	Handle(ctx context.Context, prompt AgentPrompt, next Next) (Response, error)
}
