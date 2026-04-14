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

	return slices.Clone(r.messages)
}

func (r *conversation) Tools() []contractsai.Tool { return r.agent.Tools() }

func (r *conversation) Prompt(input string) (contractsai.Response, error) {
	tools := r.agent.Tools()

	// Snapshot the message length before we touch anything so we can roll back
	// cleanly on any error path.
	r.mu.Lock()
	snapshot := len(r.messages)
	r.messages = append(r.messages, contractsai.Message{Role: contractsai.RoleUser, Content: input})
	r.mu.Unlock()

	rollback := func() {
		r.mu.Lock()
		r.messages = r.messages[:snapshot]
		r.mu.Unlock()
	}

	// Build the initial prompt. The conversation acts as its own Agent so the
	// provider always sees the live runtime history.
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
			rollback()
			return nil, err
		}

		toolCalls := resp.ToolCalls()
		if len(toolCalls) == 0 {
			// Plain text response — we're done.
			break
		}

		if i == MaxToolCallIterations-1 {
			// About to exceed the limit — roll back everything and return error.
			rollback()
			return nil, fmt.Errorf("ai: tool call loop exceeded %d iterations", MaxToolCallIterations)
		}

		// Execute each requested tool and collect results.
		toolResults, execErr := r.executeTools(tools, toolCalls)
		if execErr != nil {
			rollback()
			return nil, execErr
		}

		// Append the assistant's tool-call message and all tool results to history.
		r.mu.Lock()
		r.messages = append(r.messages, contractsai.Message{
			Role:      contractsai.RoleAssistant,
			Content:   resp.Text(),
			ToolCalls: toolCalls,
		})
		r.messages = append(r.messages, toolResults...)
		r.mu.Unlock()

		// On the next iteration the input is empty — the model continues from
		// the tool results already in the history.
		agentPrompt = contractsai.AgentPrompt{
			Agent: r,
			Input: "",
			Model: r.model,
			Tools: tools,
		}
	}

	// Append the final assistant reply to history.
	r.mu.Lock()
	r.messages = append(r.messages, contractsai.Message{Role: contractsai.RoleAssistant, Content: resp.Text()})
	r.mu.Unlock()

	return resp, nil
}

func (r *conversation) Stream(input string) (contractsai.StreamableResponse, error) {
	stream, err := r.provider.Stream(r.ctx, contractsai.AgentPrompt{
		Agent: r,
		Input: input,
		Model: r.model,
		Tools: r.agent.Tools(),
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
