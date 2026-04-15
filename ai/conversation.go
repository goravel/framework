package ai

import (
	"context"
	"fmt"
	"slices"
	"sync"

	contractsai "github.com/goravel/framework/contracts/ai"
)

// MaxToolCallIterations is the maximum number of tool-call/re-prompt cycles
// that Prompt() will execute before returning an error. This guards against
// infinite loops when the model keeps requesting tool invocations.
const MaxToolCallIterations = 10

type conversation struct {
	ctx      context.Context
	agent    contractsai.Agent
	messages []contractsai.Message
	provider contractsai.Provider
	model    string
	mu       sync.RWMutex
	// promptMu serializes concurrent Prompt() calls so that a failure in one
	// call cannot corrupt history being written by another concurrent call.
	promptMu sync.Mutex
	// pending holds the in-progress message history during a Prompt call.
	// It is set before calling the provider and cleared on commit or rollback.
	// Messages() returns pending when it is non-nil.
	pending []contractsai.Message
}

func NewConversation(ctx context.Context, agent contractsai.Agent, provider contractsai.Provider, model string) *conversation {
	return &conversation{
		ctx:      ctx,
		agent:    agent,
		messages: slices.Clone(agent.Messages()),
		provider: provider,
		model:    model,
	}
}

func (r *conversation) Instructions() string { return r.agent.Instructions() }

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
		resp, err = r.provider.Prompt(r.ctx, agentPrompt)
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
			return nil, fmt.Errorf("ai: tool call loop exceeded %d iterations", MaxToolCallIterations)
		}

		// Execute each requested tool and collect results.
		toolResults, execErr := r.executeTools(tools, toolCalls)
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

func (r *conversation) Stream(input string) (contractsai.StreamableResponse, error) {
	stream, err := r.provider.Stream(r.ctx, contractsai.AgentPrompt{
		Agent: r,
		Input: input,
		Model: r.model,
	})
	if err != nil {
		return nil, err
	}

	return stream.Then(func(resp contractsai.Response) error {
		r.mu.Lock()
		r.messages = append(r.messages,
			contractsai.Message{Role: contractsai.RoleUser, Content: input},
			contractsai.Message{Role: contractsai.RoleAssistant, Content: resp.Text()},
		)
		r.mu.Unlock()

		return nil
	}), nil
}

func (r *conversation) Reset() {
	r.mu.Lock()
	r.messages = slices.Clone(r.agent.Messages())
	r.mu.Unlock()
}

// executeTools looks up each tool call by name and invokes it.
// Returns an error if a tool is not found or its execution fails.
func (r *conversation) executeTools(tools []contractsai.Tool, calls []contractsai.ToolCall) ([]contractsai.Message, error) {
	// Build a lookup map for O(1) access.
	index := make(map[string]contractsai.Tool, len(tools))
	for _, t := range tools {
		index[t.Name()] = t
	}

	results := make([]contractsai.Message, 0, len(calls))
	for _, call := range calls {
		tool, ok := index[call.Name]
		if !ok {
			return nil, fmt.Errorf("ai: tool %q not found", call.Name)
		}

		result, err := tool.Execute(r.ctx, call.Args)
		if err != nil {
			return nil, fmt.Errorf("ai: tool %q execution failed: %w", call.Name, err)
		}

		results = append(results, contractsai.Message{
			Role:       contractsai.RoleToolResult,
			Content:    result,
			ToolCallID: call.ID,
		})
	}

	return results, nil
}
