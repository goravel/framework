package ai

import (
	"context"
	"slices"
	"sync"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
)

// MaxToolCallIterations is the maximum number of tool-call/re-prompt cycles
// that Prompt() will execute before returning an error. This guards against
// infinite loops when the model keeps requesting tool invocations.
const MaxToolCallIterations = 10

type conversation struct {
	ctx         context.Context
	agent       contractsai.Agent
	messages    []contractsai.Message
	provider    contractsai.Provider
	model       string
	middlewares []contractsai.Middleware
	mu          sync.RWMutex
	// promptMu serializes concurrent Prompt() calls so that a failure in one
	// call cannot corrupt history being written by another concurrent call.
	promptMu sync.Mutex
	// pending holds the in-progress message history during a Prompt call.
	// It is set before calling the provider and cleared on commit or rollback.
	// Messages() returns pending when it is non-nil.
	pending []contractsai.Message
}

func NewConversation(ctx context.Context, agent contractsai.Agent, provider contractsai.Provider, model string, middlewares []contractsai.Middleware) *conversation {
	return &conversation{
		ctx:         ctx,
		agent:       agent,
		messages:    slices.Clone(agent.Messages()),
		provider:    provider,
		model:       model,
		middlewares: filterNilMiddlewares(middlewares),
	}
}

func (r *conversation) Instructions() string { return r.agent.Instructions() }

func (r *conversation) Middleware() []contractsai.Middleware {
	return slices.Clone(r.middlewares)
}

func (r *conversation) Messages() []contractsai.Message {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.pending != nil {
		return slices.Clone(r.pending)
	}

	return slices.Clone(r.messages)
}

func (r *conversation) Tools() []contractsai.Tool { return r.agent.Tools() }

func (r *conversation) Prompt(input string) (contractsai.Response, error) {
	// Serialize concurrent calls to prevent one failure from corrupting
	// history written by another in-flight Prompt.
	r.promptMu.Lock()
	defer r.promptMu.Unlock()

	tools := r.agent.Tools()

	// Build an isolated working copy of the message history for this call.
	// We do not touch r.messages until the entire call succeeds.
	// The user message is NOT added to working yet — buildMessages sees the
	// current history via Messages() and appends Input itself on the first
	// call, avoiding a duplicate.
	r.mu.RLock()
	working := slices.Clone(r.messages)
	r.mu.RUnlock()

	// pending is nil for the first provider call: Messages() returns the
	// committed history and buildMessages appends Input.  After the first
	// tool-call round we flip pending to the expanded working copy so
	// subsequent iterations see the full in-progress history while Input is "".

	clearPending := func() {
		r.mu.Lock()
		r.pending = nil
		r.mu.Unlock()
	}

	agentPrompt := contractsai.AgentPrompt{
		Agent: r,
		Input: input,
		Model: r.model,
		Tools: tools,
	}

	var (
		resp contractsai.Response
		err  error
	)

	for i := range MaxToolCallIterations {
		resp, err = r.prompt(r.ctx, agentPrompt)
		if err != nil {
			clearPending()
			return nil, err
		}

		toolCalls := resp.ToolCalls()
		if len(toolCalls) == 0 {
			// Plain text response — we're done.
			break
		}

		if i == MaxToolCallIterations-1 {
			clearPending()
			return nil, errors.AIToolCallLoopExceeded.Args(MaxToolCallIterations)
		}

		// Execute each requested tool and collect results.
		toolResults, execErr := r.executeTools(r.ctx, tools, toolCalls)
		if execErr != nil {
			clearPending()
			return nil, execErr
		}

		// Extend the working copy with the user message (if this is the first
		// tool round), the assistant tool-call message and tool results, then
		// expose it via pending so the provider sees the full history on the
		// next iteration (where Input will be "").
		if i == 0 {
			working = append(working, contractsai.Message{Role: contractsai.RoleUser, Content: input})
		}
		working = append(working, contractsai.Message{
			Role:      contractsai.RoleAssistant,
			Content:   resp.Text(),
			ToolCalls: toolCalls,
		})
		working = append(working, toolResults...)

		r.mu.Lock()
		r.pending = working
		r.mu.Unlock()

		// On the next iteration Input is empty — the model continues from the
		// tool results already in the pending history.
		agentPrompt = contractsai.AgentPrompt{
			Agent: r,
			Input: "",
			Model: r.model,
			Tools: tools,
		}
	}

	// Commit: append the user message (when no tool calls were made — in the
	// tool-call path it was already added to working during the loop) and the
	// final assistant reply, then replace r.messages with the complete history.
	if r.pending == nil {
		// No tool calls occurred; working still equals the pre-call snapshot.
		// Prepend the user message before the assistant reply.
		working = append(working, contractsai.Message{Role: contractsai.RoleUser, Content: input})
	}
	working = append(working, contractsai.Message{Role: contractsai.RoleAssistant, Content: resp.Text()})

	r.mu.Lock()
	r.messages = working
	r.pending = nil
	r.mu.Unlock()

	return resp, nil
}

func (r *conversation) prompt(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
	return r.runMiddlewarePipeline(ctx, prompt, contractsai.Next(func(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
		response, err := r.provider.Prompt(ctx, prompt)
		if err != nil {
			return nil, err
		}

		return newResolvedMiddlewareResponse(response), nil
	}))
}

func (r *conversation) Stream(input string) (contractsai.StreamableResponse, error) {
	tools := r.agent.Tools()
	initialPrompt := contractsai.AgentPrompt{
		Agent: r,
		Input: input,
		Model: r.model,
		Tools: tools,
	}
	clearPending := func() {
		r.mu.Lock()
		r.pending = nil
		r.mu.Unlock()
	}

	initialCtx, cancelInitial := context.WithCancel(r.ctx)
	resolvedInitialPrompt, initialMiddlewareResponse, shortCircuitResponse, err := r.prepareStreamPrompt(initialCtx, initialPrompt)
	if err != nil {
		cancelInitial()
		clearPending()
		return nil, err
	}

	if shortCircuitResponse != nil {
		cancelInitial()
		return NewStreamableResponse(r.ctx, func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			defer clearPending()

			r.mu.RLock()
			working := slices.Clone(r.messages)
			r.mu.RUnlock()

			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeDone}); err != nil {
				return nil, err
			}

			r.commitConversation(working, input, shortCircuitResponse)
			return shortCircuitResponse, nil
		}), nil
	}

	r.promptMu.Lock()
	initialStream, err := r.provider.Stream(initialCtx, resolvedInitialPrompt)
	r.promptMu.Unlock()
	if err != nil {
		cancelInitial()
		clearPending()
		return nil, err
	}

	return NewStreamableResponse(r.ctx, func(streamCtx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
		defer clearPending()

		r.promptMu.Lock()
		defer r.promptMu.Unlock()

		initialDone := make(chan struct{})
		defer close(initialDone)
		go func() {
			select {
			case <-streamCtx.Done():
				cancelInitial()
			case <-initialDone:
			}
		}()

		r.mu.RLock()
		working := slices.Clone(r.messages)
		r.mu.RUnlock()

		agentPrompt := initialPrompt
		useInitialStream := true

		for i := range MaxToolCallIterations {
			var (
				innerStream        contractsai.StreamableResponse
				middlewareResponse *middlewareResponse
			)

			if useInitialStream {
				innerStream = initialStream
				middlewareResponse = initialMiddlewareResponse
				useInitialStream = false
			} else {
				resolvedPrompt, nextMiddlewareResponse, shortCircuitResponse, err := r.prepareStreamPrompt(streamCtx, agentPrompt)
				if err != nil {
					return nil, err
				}
				if shortCircuitResponse != nil {
					if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeDone}); err != nil {
						return nil, err
					}

					r.commitConversation(working, input, shortCircuitResponse)
					return shortCircuitResponse, nil
				}

				middlewareResponse = nextMiddlewareResponse
				innerStream, err = r.provider.Stream(streamCtx, resolvedPrompt)
				if err != nil {
					return nil, err
				}
			}

			var resp contractsai.Response
			innerStream.Then(func(res contractsai.Response) {
				resp = res
			})

			var doneEvent *contractsai.StreamEvent

			// Forward all events from this iteration, suppressing intermediate
			// done events — only the final iteration's done event is forwarded.
			if err := innerStream.Each(func(event contractsai.StreamEvent) error {
				if event.Type == contractsai.StreamEventTypeDone {
					event := event
					doneEvent = &event
					return nil
				}
				return emit(event)
			}); err != nil {
				return nil, err
			}

			if resp == nil {
				return nil, errors.AIResponseIsNil
			}

			toolCalls := resp.ToolCalls()
			if len(toolCalls) == 0 {
				finalResp := finalizeMiddlewareResponse(middlewareResponse, resp)

				if doneEvent != nil {
					if err := emit(*doneEvent); err != nil {
						return nil, err
					}
				}

				r.commitConversation(working, input, finalResp)

				return finalResp, nil
			}

			if i == MaxToolCallIterations-1 {
				return nil, errors.AIToolCallLoopExceeded.Args(MaxToolCallIterations)
			}

			if err := emit(contractsai.StreamEvent{
				Type:      contractsai.StreamEventTypeToolCall,
				ToolCalls: toolCalls,
			}); err != nil {
				return nil, err
			}

			toolResults, execErr := r.executeTools(streamCtx, tools, toolCalls)
			if execErr != nil {
				return nil, execErr
			}

			if i == 0 {
				working = append(working, contractsai.Message{Role: contractsai.RoleUser, Content: input})
			}
			working = append(working, contractsai.Message{
				Role:      contractsai.RoleAssistant,
				Content:   resp.Text(),
				ToolCalls: toolCalls,
			})
			working = append(working, toolResults...)

			r.mu.Lock()
			r.pending = working
			r.mu.Unlock()

			agentPrompt = contractsai.AgentPrompt{
				Agent: r,
				Input: "",
				Model: r.model,
				Tools: tools,
			}
		}

		return nil, errors.AIToolCallLoopExceeded.Args(MaxToolCallIterations)
	}), nil
}

func (r *conversation) prepareStreamPrompt(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.AgentPrompt, *middlewareResponse, contractsai.Response, error) {
	resolvedPrompt := prompt
	deferredResponse := newDeferredMiddlewareResponse()
	calledNext := false

	response, err := r.runMiddlewarePipeline(ctx, prompt, contractsai.Next(func(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
		calledNext = true
		resolvedPrompt = prompt
		return deferredResponse, nil
	}))
	if err != nil {
		return contractsai.AgentPrompt{}, nil, nil, err
	}

	if !calledNext {
		if response == nil {
			return contractsai.AgentPrompt{}, nil, nil, errors.AIResponseIsNil
		}

		return contractsai.AgentPrompt{}, nil, response, nil
	}

	return resolvedPrompt, deferredResponse, nil, nil
}

func (r *conversation) runMiddlewarePipeline(ctx context.Context, prompt contractsai.AgentPrompt, destination contractsai.Next) (contractsai.Response, error) {
	next := destination

	for i := len(r.middlewares) - 1; i >= 0; i-- {
		middleware := r.middlewares[i]
		nextHandler := next
		next = contractsai.Next(func(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
			return middleware.Handle(ctx, prompt, nextHandler)
		})
	}

	response, err := next(ctx, prompt)
	if err != nil {
		return nil, err
	}

	if middlewareResponse, ok := response.(*middlewareResponse); ok {
		return middlewareResponse.Unwrap(), nil
	}

	return response, nil
}

func finalizeMiddlewareResponse(middlewareResponse *middlewareResponse, response contractsai.Response) contractsai.Response {
	if middlewareResponse == nil {
		return response
	}

	middlewareResponse.Resolve(response)

	return middlewareResponse.Unwrap()
}

func (r *conversation) commitConversation(working []contractsai.Message, input string, response contractsai.Response) {
	if r.pending == nil {
		working = append(working, contractsai.Message{Role: contractsai.RoleUser, Content: input})
	}
	working = append(working, contractsai.Message{Role: contractsai.RoleAssistant, Content: response.Text()})

	r.mu.Lock()
	r.messages = working
	r.pending = nil
	r.mu.Unlock()
}

func (r *conversation) Reset() {
	r.mu.Lock()
	r.messages = slices.Clone(r.agent.Messages())
	r.mu.Unlock()
}

// executeTools looks up each tool call by name and invokes it.
// Returns an error if a tool is not found or its execution fails.
func (r *conversation) executeTools(ctx context.Context, tools []contractsai.Tool, calls []contractsai.ToolCall) ([]contractsai.Message, error) {
	// Build a lookup map for O(1) access.
	index := make(map[string]contractsai.Tool, len(tools))
	for _, t := range tools {
		index[t.Name()] = t
	}

	results := make([]contractsai.Message, 0, len(calls))
	for _, call := range calls {
		tool, ok := index[call.Name]
		if !ok {
			return nil, errors.AIToolNotFound.Args(call.Name)
		}

		result, err := tool.Execute(ctx, call.Args)
		if err != nil {
			return nil, errors.AIToolExecutionFailed.Args(call.Name, err)
		}

		results = append(results, contractsai.Message{
			Role:       contractsai.RoleToolResult,
			Content:    result,
			ToolCallID: call.ID,
		})
	}

	return results, nil
}
