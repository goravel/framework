package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime"
	"path/filepath"
	"strings"

	goopenai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"

	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/errors"
)

// The OpenAI provider will be moved into a separate package in the future.

const DefaultTextModel = "gpt-5.4"

const providerStateResponseID = "openai.response_id"

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
	params, err := r.buildRequest(ctx, prompt)
	if err != nil {
		return nil, err
	}

	completion, err := r.client.Responses.New(ctx, params)
	if err != nil {
		return nil, err
	}

	text, toolCalls := r.parseOutput(completion.Output)
	if completion.ID != "" && prompt.ProviderState != nil {
		prompt.ProviderState.Set(providerStateResponseID, completion.ID)
	}

	return &response{
		text:      text,
		toolCalls: toolCalls,
		usage:     r.parseUsage(completion.Usage),
	}, nil
}

func (r *Provider) Stream(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
	params, err := r.buildRequest(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return frameworkai.NewStreamableResponse(ctx, func(streamCtx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
		stream := r.client.Responses.NewStreaming(streamCtx, params)
		defer errors.Ignore(stream.Close)

		text := strings.Builder{}
		currentUsage := &usage{}
		responseID := ""
		var toolCalls []contractsai.ToolCall

		for stream.Next() {
			event := stream.Current()
			switch chunk := event.AsAny().(type) {
			case responses.ResponseTextDeltaEvent:
				text.WriteString(chunk.Delta)
				if err := emit(contractsai.StreamEvent{
					Type:  contractsai.StreamEventTypeTextDelta,
					Delta: chunk.Delta,
				}); err != nil {
					return nil, err
				}
			case responses.ResponseCompletedEvent:
				responseID = chunk.Response.ID
				toolText, parsedToolCalls := r.parseOutput(chunk.Response.Output)
				if text.Len() == 0 && toolText != "" {
					text.WriteString(toolText)
				}
				toolCalls = parsedToolCalls
				currentUsage = r.parseUsage(chunk.Response.Usage)
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
		if responseID != "" && prompt.ProviderState != nil {
			prompt.ProviderState.Set(providerStateResponseID, responseID)
		}

		return &response{
			text:      text.String(),
			toolCalls: toolCalls,
			usage:     currentUsage,
		}, nil
	}), nil
}

func (r *Provider) PutFile(ctx context.Context, file contractsai.StorableFile, _ contractsai.Options) (contractsai.StoredFileResponse, error) {
	content, err := file.Content(ctx)
	if err != nil {
		return nil, err
	}

	params := goopenai.FileNewParams{
		File:    goopenai.File(bytes.NewReader(content), file.FileName(), file.MimeType()),
		Purpose: goopenai.FilePurposeUserData,
	}

	upload, err := r.client.Files.New(ctx, params)
	if err != nil {
		return nil, err
	}

	return &storedFileResponse{id: upload.ID}, nil
}

func (r *Provider) resolveModel(model string) string {
	if model != "" {
		return model
	}

	return r.config.Models.Text.Default
}

func (r *Provider) buildRequest(ctx context.Context, prompt contractsai.AgentPrompt) (responses.ResponseNewParams, error) {
	input, instructions, previousResponseID, err := r.buildInput(ctx, prompt)
	if err != nil {
		return responses.ResponseNewParams{}, err
	}

	params := responses.ResponseNewParams{
		Model: shared.ResponsesModel(r.resolveModel(prompt.Model)),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: input,
		},
		ParallelToolCalls: param.NewOpt(true),
	}
	if instructions != "" {
		params.Instructions = param.NewOpt(instructions)
	}
	if previousResponseID != "" {
		params.PreviousResponseID = param.NewOpt(previousResponseID)
	}
	if len(prompt.Tools) > 0 {
		params.Tools = r.buildTools(prompt.Tools)
	}

	return params, nil
}

// buildInput converts the conversation history and current input into the
// Responses API input items that the API expects.
func (r *Provider) buildInput(ctx context.Context, prompt contractsai.AgentPrompt) ([]responses.ResponseInputItemUnionParam, string, string, error) {
	var previousResponseID string
	if prompt.ProviderState != nil {
		previousResponseID, _ = prompt.ProviderState.Get(providerStateResponseID).(string)
	}
	if previousResponseID != "" && prompt.Input == "" {
		if toolResultInput := r.buildToolResultInput(prompt.Agent.Messages()); len(toolResultInput) > 0 {
			return toolResultInput, prompt.Agent.Instructions(), previousResponseID, nil
		}
	}

	input := make([]responses.ResponseInputItemUnionParam, 0)
	instructions := prompt.Agent.Instructions()
	history := prompt.Agent.Messages()
	attachmentIndex := -1
	if prompt.Input == "" && len(prompt.Attachments) > 0 {
		for i := len(history) - 1; i >= 0; i-- {
			if history[i].Role == contractsai.RoleUser {
				attachmentIndex = i
				break
			}
		}
	}

	for i, m := range history {
		switch m.Role {
		case contractsai.RoleUser:
			attachments := []contractsai.Attachment(nil)
			if i == attachmentIndex {
				attachments = prompt.Attachments
			}

			message, err := r.buildUserInputItem(ctx, m.Content, attachments)
			if err != nil {
				return nil, "", "", err
			}
			input = append(input, message)
		case contractsai.RoleAssistant:
			if m.Content != "" || len(m.ToolCalls) == 0 {
				input = append(input, r.buildAssistantInputItem(m.Content))
			}
			for _, tc := range m.ToolCalls {
				callID := tc.ID
				if callID == "" {
					callID = tc.Name
				}
				input = append(input, responses.ResponseInputItemUnionParam{OfFunctionCall: &responses.ResponseFunctionToolCallParam{
					CallID:    callID,
					Name:      tc.Name,
					Arguments: tc.RawArgs,
					Status:    responses.ResponseFunctionToolCallStatusCompleted,
				}})
			}
		case contractsai.RoleToolResult:
			input = append(input, responses.ResponseInputItemUnionParam{OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
				CallID: m.ToolCallID,
				Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
					OfString: param.NewOpt(m.Content),
				},
			}})
		}
	}

	if prompt.Input != "" || (len(prompt.Attachments) > 0 && attachmentIndex == -1) {
		message, err := r.buildUserInputItem(ctx, prompt.Input, prompt.Attachments)
		if err != nil {
			return nil, "", "", err
		}
		input = append(input, message)
	}

	return input, instructions, "", nil
}

func (r *Provider) buildToolResultInput(history []contractsai.Message) []responses.ResponseInputItemUnionParam {
	input := make([]responses.ResponseInputItemUnionParam, 0)
	for i := len(history) - 1; i >= 0; i-- {
		message := history[i]
		if message.Role != contractsai.RoleToolResult {
			if len(input) > 0 {
				break
			}
			continue
		}

		input = append(input, responses.ResponseInputItemUnionParam{OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
			CallID: message.ToolCallID,
			Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
				OfString: param.NewOpt(message.Content),
			},
		}})
	}

	for left, right := 0, len(input)-1; left < right; left, right = left+1, right-1 {
		input[left], input[right] = input[right], input[left]
	}

	return input
}

func (r *Provider) buildUserInputItem(ctx context.Context, input string, attachments []contractsai.Attachment) (responses.ResponseInputItemUnionParam, error) {
	if len(attachments) == 0 {
		return responses.ResponseInputItemUnionParam{OfMessage: &responses.EasyInputMessageParam{
			Role: responses.EasyInputMessageRoleUser,
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: param.NewOpt(input),
			},
		}}, nil
	}

	parts := make([]responses.ResponseInputContentUnionParam, 0, len(attachments)+1)
	if input != "" {
		parts = append(parts, responses.ResponseInputContentUnionParam{OfInputText: &responses.ResponseInputTextParam{Text: input}})
	}
	for _, attachment := range attachments {
		switch attachment.Kind() {
		case contractsai.AttachmentKindImage:
			content, err := attachment.Content(ctx)
			if err != nil {
				return responses.ResponseInputItemUnionParam{}, err
			}

			parts = append(parts, responses.ResponseInputContentUnionParam{OfInputImage: &responses.ResponseInputImageParam{
				Detail:   responses.ResponseInputImageDetailAuto,
				ImageURL: param.NewOpt(r.dataURL(attachment.MimeType(), content)),
			}})
		case contractsai.AttachmentKindFile:
			content, err := attachment.Content(ctx)
			if err != nil {
				return responses.ResponseInputItemUnionParam{}, err
			}
			if r.shouldInlineFileAttachment(attachment.FileName(), attachment.MimeType()) {
				parts = append(parts, responses.ResponseInputContentUnionParam{OfInputText: &responses.ResponseInputTextParam{
					Text: r.inlineFileText(attachment.FileName(), content),
				}})
				continue
			}

			parts = append(parts, responses.ResponseInputContentUnionParam{OfInputFile: &responses.ResponseInputFileParam{
				FileData: param.NewOpt(r.dataURL(attachment.MimeType(), content)),
				Filename: param.NewOpt(attachment.FileName()),
			}})
		default:
			return responses.ResponseInputItemUnionParam{}, errors.AIUnsupportedAttachmentKind.Args(attachment.Kind())
		}
	}

	return responses.ResponseInputItemUnionParam{OfMessage: &responses.EasyInputMessageParam{
		Role: responses.EasyInputMessageRoleUser,
		Content: responses.EasyInputMessageContentUnionParam{
			OfInputItemContentList: parts,
		},
	}}, nil
}

func (r *Provider) buildAssistantInputItem(input string) responses.ResponseInputItemUnionParam {
	return responses.ResponseInputItemUnionParam{OfMessage: &responses.EasyInputMessageParam{
		Role: responses.EasyInputMessageRoleAssistant,
		Content: responses.EasyInputMessageContentUnionParam{
			OfString: param.NewOpt(input),
		},
	}}
}

func (r *Provider) dataURL(mimeType string, content []byte) string {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if mediaType, _, err := mime.ParseMediaType(mimeType); err == nil && mediaType != "" {
		mimeType = mediaType
	}

	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(content)
}

func (r *Provider) shouldInlineFileAttachment(fileName, mimeType string) bool {
	if mimeType != "" {
		mediaType, _, err := mime.ParseMediaType(mimeType)
		if err != nil {
			mediaType = mimeType
		}

		if strings.HasPrefix(mediaType, "text/") {
			return true
		}

		switch mediaType {
		case "application/json", "application/xml", "application/yaml", "application/x-yaml":
			return true
		}
	}

	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".txt", ".md", ".json", ".xml", ".yaml", ".yml", ".csv", ".html", ".htm":
		return true
	default:
		return false
	}
}

func (r *Provider) inlineFileText(fileName string, content []byte) string {
	if fileName == "" {
		return string(content)
	}

	return "File: " + fileName + "\n\n" + string(content)
}

// buildTools converts a slice of Tool definitions into OpenAI Responses tool params.
func (r *Provider) buildTools(tools []contractsai.Tool) []responses.ToolUnionParam {
	params := make([]responses.ToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		fn := responses.FunctionToolParam{
			Name: tool.Name(),
		}
		if desc := tool.Description(); desc != "" {
			fn.Description = param.NewOpt(desc)
		}
		if schema := tool.Parameters(); schema != nil {
			fn.Strict = param.NewOpt(true)
			fn.Parameters = schema
		}
		params = append(params, responses.ToolUnionParam{OfFunction: &fn})
	}
	return params
}

func (r *Provider) parseOutput(raw []responses.ResponseOutputItemUnion) (string, []contractsai.ToolCall) {
	text := strings.Builder{}
	toolCalls := make([]contractsai.ToolCall, 0)
	for _, item := range raw {
		switch value := item.AsAny().(type) {
		case responses.ResponseOutputMessage:
			for _, content := range value.Content {
				switch part := content.AsAny().(type) {
				case responses.ResponseOutputText:
					text.WriteString(part.Text)
				}
			}
		case responses.ResponseFunctionToolCall:
			args := make(map[string]any)
			if value.Arguments != "" {
				_ = json.Unmarshal([]byte(value.Arguments), &args)
			}
			toolCalls = append(toolCalls, contractsai.ToolCall{
				ID:      value.CallID,
				Name:    value.Name,
				Args:    args,
				RawArgs: value.Arguments,
			})
		}
	}
	if len(toolCalls) == 0 {
		return text.String(), nil
	}

	return text.String(), toolCalls
}

func (r *Provider) parseUsage(raw responses.ResponseUsage) *usage {
	return &usage{
		input:  int(raw.InputTokens),
		output: int(raw.OutputTokens),
		total:  int(raw.TotalTokens),
	}
}
