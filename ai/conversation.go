package ai

import (
	"context"
	"slices"

	contractsai "github.com/goravel/framework/contracts/ai"
)

// runtimeAgent wraps the user's Agent and tracks accumulated conversation history.
type runtimeAgent struct {
	agent    contractsai.Agent
	messages []contractsai.Message
}

func newRuntimeAgent(agent contractsai.Agent) *runtimeAgent {
	return &runtimeAgent{
		agent:    agent,
		messages: slices.Clone(agent.Messages()),
	}
}

func (r *runtimeAgent) Instructions() string            { return r.agent.Instructions() }
func (r *runtimeAgent) Messages() []contractsai.Message { return r.messages }

type conversation struct {
	ctx      context.Context
	agent    *runtimeAgent
	provider contractsai.Provider
	model    string
}

func NewConversation(ctx context.Context, agent contractsai.Agent, provider contractsai.Provider, model string) *conversation {
	return &conversation{
		ctx:      ctx,
		agent:    newRuntimeAgent(agent),
		provider: provider,
		model:    model,
	}
}

func (r *conversation) Prompt(input string) (contractsai.Response, error) {
	resp, err := r.provider.Prompt(r.ctx, contractsai.AgentPrompt{
		Agent: r.agent,
		Input: input,
		Model: r.model,
	})
	if err != nil {
		return nil, err
	}

	r.agent.messages = append(r.agent.messages,
		contractsai.Message{Role: contractsai.RoleUser, Content: input},
		contractsai.Message{Role: contractsai.RoleAssistant, Content: resp.Text()},
	)

	return resp, nil
}

func (r *conversation) Messages() []contractsai.Message { return r.agent.Messages() }

func (r *conversation) Reset() { r.agent.messages = slices.Clone(r.agent.agent.Messages()) }
