package ai

import (
	"context"
	"slices"
	"sync"

	contractsai "github.com/goravel/framework/contracts/ai"
)

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

func (r *conversation) Prompt(input string) (contractsai.Response, error) {
	resp, err := r.provider.Prompt(r.ctx, contractsai.AgentPrompt{
		Agent: r,
		Input: input,
		Model: r.model,
	})
	if err != nil {
		return nil, err
	}

	text := resp.Text()

	r.mu.Lock()
	r.messages = append(r.messages,
		contractsai.Message{Role: contractsai.RoleUser, Content: input},
		contractsai.Message{Role: contractsai.RoleAssistant, Content: text},
	)
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
