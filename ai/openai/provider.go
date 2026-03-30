package openai

import (
	"context"

	goopenai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	contractsai "github.com/goravel/framework/contracts/ai"
	contractsconfig "github.com/goravel/framework/contracts/config"
)

// The OpenAI provider will be moved into a separate package in the future.

const DefaultTextModel = "gpt-5.4"

type Provider struct {
	client goopenai.Client
	config contractsai.ProviderConfig
}

func NewOpenAI(config contractsconfig.Config, provider string) (*Provider, error) {
	var providerConfig contractsai.ProviderConfig
	err := config.UnmarshalKey("ai.providers."+provider, &providerConfig)
	if err != nil {
		return nil, err
	}

	opts := []option.RequestOption{option.WithAPIKey(providerConfig.Key)}
	if providerConfig.Url != "" {
		opts = append(opts, option.WithBaseURL(providerConfig.Url))
	}
	if providerConfig.Models.Text.Default == "" {
		providerConfig.Models.Text.Default = DefaultTextModel
	}

	return &Provider{client: goopenai.NewClient(opts...), config: providerConfig}, nil
}

func (r *Provider) Prompt(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
	model := r.config.Models.Text.Default
	if prompt.Model != "" {
		model = prompt.Model
	}

	var messages []goopenai.ChatCompletionMessageParamUnion
	if instructions := prompt.Agent.Instructions(); instructions != "" {
		messages = append(messages, goopenai.SystemMessage(instructions))
	}
	for _, m := range prompt.Agent.Messages() {
		switch m.Role {
		case contractsai.RoleUser:
			messages = append(messages, goopenai.UserMessage(m.Content))
		case contractsai.RoleAssistant:
			messages = append(messages, goopenai.AssistantMessage(m.Content))
		}
	}
	messages = append(messages, goopenai.UserMessage(prompt.Input))

	completion, err := r.client.Chat.Completions.New(ctx, goopenai.ChatCompletionNewParams{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return nil, err
	}

	text := ""
	if len(completion.Choices) > 0 {
		text = completion.Choices[0].Message.Content
	}
	return &response{
		text: text,
		usage: &usage{
			input:  int(completion.Usage.PromptTokens),
			output: int(completion.Usage.CompletionTokens),
			total:  int(completion.Usage.TotalTokens),
		},
	}, nil
}
