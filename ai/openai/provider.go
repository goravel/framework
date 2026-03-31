package openai

import (
	"context"
	"strings"

	goopenai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/errors"
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
	model := r.resolveModel(prompt.Model)
	messages := r.buildMessages(prompt)

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

func (r *Provider) Stream(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
	model := r.resolveModel(prompt.Model)
	messages := r.buildMessages(prompt)

	return frameworkai.NewStreamableResponse(ctx, func(streamCtx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
		stream := r.client.Chat.Completions.NewStreaming(streamCtx, goopenai.ChatCompletionNewParams{
			Model:    model,
			Messages: messages,
			StreamOptions: goopenai.ChatCompletionStreamOptionsParam{
				IncludeUsage: goopenai.Bool(true),
			},
		})
		defer errors.Ignore(stream.Close)

		text := strings.Builder{}
		currentUsage := &usage{}

		for stream.Next() {
			chunk := stream.Current()
			if chunk.JSON.Usage.Valid() {
				currentUsage = &usage{
					input:  int(chunk.Usage.PromptTokens),
					output: int(chunk.Usage.CompletionTokens),
					total:  int(chunk.Usage.TotalTokens),
				}
			}

			if len(chunk.Choices) == 0 {
				continue
			}

			delta := chunk.Choices[0].Delta.Content
			if delta == "" {
				continue
			}

			text.WriteString(delta)
			if err := emit(contractsai.StreamEvent{
				Type:  contractsai.StreamEventTypeTextDelta,
				Delta: delta,
			}); err != nil {
				return nil, err
			}
		}

		if err := stream.Err(); err != nil {
			if streamCtx.Err() == nil {
				if emitErr := emit(contractsai.StreamEvent{
					Type:  contractsai.StreamEventTypeError,
					Error: err.Error(),
				}); emitErr != nil {
					return nil, emitErr
				}
			}

			return nil, err
		}

		if err := emit(contractsai.StreamEvent{
			Type:  contractsai.StreamEventTypeDone,
			Usage: currentUsage,
		}); err != nil {
			return nil, err
		}

		return &response{
			text:  text.String(),
			usage: currentUsage,
		}, nil
	}), nil
}

func (r *Provider) resolveModel(model string) string {
	if model != "" {
		return model
	}

	return r.config.Models.Text.Default
}

func (r *Provider) buildMessages(prompt contractsai.AgentPrompt) []goopenai.ChatCompletionMessageParamUnion {
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

	return messages
}
