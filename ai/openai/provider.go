package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"

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
const DefaultAudioModel = "gpt-4o-mini-tts"
const DefaultTranscriptionModel = "gpt-4o-mini-transcribe"
const DefaultDiarizedTranscriptionModel = "gpt-4o-transcribe-diarize"
const DefaultImageModel = "gpt-image-2"
const DefaultFemaleVoice = "alloy"
const DefaultMaleVoice = "ash"

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
	if providerConfig.Models.Audio.Default == "" {
		providerConfig.Models.Audio.Default = DefaultAudioModel
	}
	if providerConfig.Models.Transcription.Default == "" {
		providerConfig.Models.Transcription.Default = DefaultTranscriptionModel
	}
	if providerConfig.Models.Image.Default == "" {
		providerConfig.Models.Image.Default = DefaultImageModel
	}

	return &Provider{client: goopenai.NewClient(opts...), config: providerConfig}, nil
}

func (r *Provider) Audio(ctx context.Context, prompt contractsai.AudioPrompt) (contractsai.AudioResponse, error) {
	if prompt.Prompt == "" {
		return nil, errors.AIAudioPromptRequired
	}

	requestOptions := make([]option.RequestOption, 0, 1)
	if prompt.Timeout > 0 {
		requestOptions = append(requestOptions, option.WithRequestTimeout(prompt.Timeout))
	}

	params := goopenai.AudioSpeechNewParams{
		Input: prompt.Prompt,
		Model: goopenai.SpeechModel(r.resolveAudioModel(prompt.Model)),
		Voice: goopenai.AudioSpeechNewParamsVoiceUnion{
			OfString: param.NewOpt(r.resolveAudioVoice(prompt.Voice)),
		},
		ResponseFormat: goopenai.AudioSpeechNewParamsResponseFormatMP3,
	}
	if prompt.Instructions != "" {
		params.Instructions = param.NewOpt(prompt.Instructions)
	}

	response, err := r.client.Audio.Speech.New(ctx, params, requestOptions...)
	if err != nil {
		return nil, err
	}

	return r.parseAudioResponse(response, params.ResponseFormat)
}

func (r *Provider) Transcription(ctx context.Context, prompt contractsai.TranscriptionPrompt) (contractsai.TranscriptionResponse, error) {
	if isNilInterface(prompt.File) {
		return nil, errors.AITranscriptionFileRequired
	}

	content, err := prompt.File.Content(ctx)
	if err != nil {
		return nil, err
	}

	requestOptions := make([]option.RequestOption, 0, 1)
	if prompt.Timeout > 0 {
		requestOptions = append(requestOptions, option.WithRequestTimeout(prompt.Timeout))
	}

	params := goopenai.AudioTranscriptionNewParams{
		File:           goopenai.File(bytes.NewReader(content), r.uploadFilename(prompt.File), prompt.File.MimeType()),
		Model:          goopenai.AudioModel(r.resolveTranscriptionModel(prompt.Model, prompt.Diarize)),
		ResponseFormat: r.resolveTranscriptionResponseFormat(prompt.Diarize),
	}
	if prompt.Language != "" {
		params.Language = param.NewOpt(prompt.Language)
	}

	response, err := r.client.Audio.Transcriptions.New(ctx, params, requestOptions...)
	if err != nil {
		return nil, err
	}

	return r.parseTranscriptionResponse(response, prompt.Diarize)
}

func (r *Provider) Image(ctx context.Context, prompt contractsai.ImagePrompt) (contractsai.ImageResponse, error) {
	if prompt.Prompt == "" {
		return nil, errors.AIImagePromptRequired
	}
	for _, attachment := range prompt.Attachments {
		if attachment.Kind() != contractsai.AttachmentKindImage {
			return nil, errors.AIImageAttachmentRequired
		}
	}

	requestOptions := make([]option.RequestOption, 0, 1)
	if prompt.Timeout > 0 {
		requestOptions = append(requestOptions, option.WithRequestTimeout(prompt.Timeout))
	}

	if len(prompt.Attachments) == 0 {
		params := goopenai.ImageGenerateParams{
			Prompt: prompt.Prompt,
			Model:  goopenai.ImageModel(r.resolveImageModel(prompt.Model)),
		}
		if size := r.resolveImageGenerateSize(prompt.Size); size != "" {
			params.Size = size
		}
		if quality := r.resolveImageGenerateQuality(prompt.Quality); quality != "" {
			params.Quality = quality
		}

		response, err := r.client.Images.Generate(ctx, params, requestOptions...)
		if err != nil {
			return nil, err
		}

		return r.parseImageResponse(response)
	}

	images := make([]io.Reader, 0, len(prompt.Attachments))
	for _, attachment := range prompt.Attachments {
		content, err := attachment.Content(ctx)
		if err != nil {
			return nil, err
		}

		images = append(images, goopenai.File(bytes.NewReader(content), r.uploadFilename(attachment), attachment.MimeType()))
	}

	params := goopenai.ImageEditParams{
		Prompt: prompt.Prompt,
		Model:  goopenai.ImageModel(r.resolveImageModel(prompt.Model)),
		Image: goopenai.ImageEditParamsImageUnion{
			OfFileArray: images,
		},
	}
	if size := r.resolveImageEditSize(prompt.Size); size != "" {
		params.Size = size
	}
	if quality := r.resolveImageEditQuality(prompt.Quality); quality != "" {
		params.Quality = quality
	}

	response, err := r.client.Images.Edit(ctx, params, requestOptions...)
	if err != nil {
		return nil, err
	}

	return r.parseImageResponse(response)
}

func (r *Provider) Prompt(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.AgentResponse, error) {
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

	return frameworkai.NewTextResponse(text, r.parseUsage(completion.Usage), toolCalls), nil
}

func (r *Provider) Stream(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableAgentResponse, error) {
	params, err := r.buildRequest(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return frameworkai.NewStreamableResponse(ctx, func(streamCtx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.AgentResponse, error) {
		stream := r.client.Responses.NewStreaming(streamCtx, params)
		defer errors.Ignore(stream.Close)

		text := strings.Builder{}
		currentUsage := contractsai.Usage(frameworkai.NewUsage(0, 0, 0))
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

		return frameworkai.NewTextResponse(text.String(), currentUsage, toolCalls), nil
	}), nil
}

func (r *Provider) PutFile(ctx context.Context, file contractsai.StorableFile) (contractsai.FileResponse, error) {
	content, err := file.Content(ctx)
	if err != nil {
		return nil, err
	}

	params := goopenai.FileNewParams{
		File:    goopenai.File(bytes.NewReader(content), r.uploadFilename(file), file.MimeType()),
		Purpose: goopenai.FilePurposeUserData,
	}

	upload, err := r.client.Files.New(ctx, params)
	if err != nil {
		return nil, err
	}

	return frameworkai.NewFileResponse(upload.ID, "", nil), nil
}

func (r *Provider) GetFile(ctx context.Context, id string) (contractsai.FileResponse, error) {
	file, err := r.client.Files.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	response, err := r.client.Files.Content(ctx, id)
	if err != nil {
		return nil, err
	}
	defer errors.Ignore(response.Body.Close)

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	mimeType := mime.TypeByExtension(filepath.Ext(file.Filename))
	if mimeType == "" {
		mimeType = response.Header.Get("Content-Type")
	}

	return frameworkai.NewFileResponse(file.ID, mimeType, content), nil
}

func (r *Provider) DeleteFile(ctx context.Context, id string) error {
	_, err := r.client.Files.Delete(ctx, id)
	return err
}

func (r *Provider) uploadFilename(file contractsai.StorableFile) string {
	if fileName := file.FileName(); fileName != "" {
		return fileName
	}

	mediaType := file.MimeType()
	if parsed, _, err := mime.ParseMediaType(mediaType); err == nil {
		mediaType = parsed
	}

	extensions, err := mime.ExtensionsByType(mediaType)
	if err == nil && len(extensions) > 0 {
		return "attachment" + extensions[0]
	}

	return fmt.Sprintf("attachment%s", fallbackFileExtension(mediaType))
}

func fallbackFileExtension(mimeType string) string {
	switch strings.ToLower(mimeType) {
	case "text/plain", "text/plain; charset=utf-8":
		return ".txt"
	case "application/json":
		return ".json"
	case "application/pdf":
		return ".pdf"
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	default:
		return ".bin"
	}
}

func (r *Provider) resolveModel(model string) string {
	if model != "" {
		return model
	}

	return r.config.Models.Text.Default
}

func (r *Provider) resolveAudioModel(model string) string {
	if model != "" {
		return model
	}

	return r.config.Models.Audio.Default
}

func (r *Provider) resolveTranscriptionModel(model string, diarize bool) string {
	if model != "" {
		return model
	}
	if diarize {
		if r.config.Models.Transcription.Default != "" && r.config.Models.Transcription.Default != DefaultTranscriptionModel {
			return r.config.Models.Transcription.Default
		}

		return DefaultDiarizedTranscriptionModel
	}
	if r.config.Models.Transcription.Default == "" {
		return DefaultTranscriptionModel
	}

	return r.config.Models.Transcription.Default
}

func (r *Provider) resolveTranscriptionResponseFormat(diarize bool) goopenai.AudioResponseFormat {
	if diarize {
		return goopenai.AudioResponseFormatDiarizedJSON
	}

	return goopenai.AudioResponseFormatJSON
}

func (r *Provider) resolveAudioVoice(voice string) string {
	switch voice {
	case "", frameworkai.DefaultFemaleVoice:
		return string(goopenai.AudioSpeechNewParamsVoiceString2Alloy)
	case frameworkai.DefaultMaleVoice:
		return string(goopenai.AudioSpeechNewParamsVoiceString2Ash)
	default:
		return voice
	}
}

func (r *Provider) resolveImageModel(model string) string {
	if model != "" {
		return model
	}

	return r.config.Models.Image.Default
}

func (r *Provider) resolveImageGenerateSize(size contractsai.ImageSize) goopenai.ImageGenerateParamsSize {
	switch size {
	case contractsai.ImageSizeSquare:
		return goopenai.ImageGenerateParamsSize1024x1024
	case contractsai.ImageSizePortrait:
		return goopenai.ImageGenerateParamsSize1024x1536
	case contractsai.ImageSizeLandscape:
		return goopenai.ImageGenerateParamsSize1536x1024
	default:
		return ""
	}
}

func (r *Provider) resolveImageEditSize(size contractsai.ImageSize) goopenai.ImageEditParamsSize {
	switch size {
	case contractsai.ImageSizeSquare:
		return goopenai.ImageEditParamsSize1024x1024
	case contractsai.ImageSizePortrait:
		return goopenai.ImageEditParamsSize1024x1536
	case contractsai.ImageSizeLandscape:
		return goopenai.ImageEditParamsSize1536x1024
	default:
		return ""
	}
}

func (r *Provider) resolveImageGenerateQuality(quality contractsai.ImageQuality) goopenai.ImageGenerateParamsQuality {
	switch quality {
	case contractsai.ImageQualityLow:
		return goopenai.ImageGenerateParamsQualityLow
	case contractsai.ImageQualityMedium:
		return goopenai.ImageGenerateParamsQualityMedium
	case contractsai.ImageQualityHigh:
		return goopenai.ImageGenerateParamsQualityHigh
	default:
		return ""
	}
}

func (r *Provider) resolveImageEditQuality(quality contractsai.ImageQuality) goopenai.ImageEditParamsQuality {
	switch quality {
	case contractsai.ImageQualityLow:
		return goopenai.ImageEditParamsQualityLow
	case contractsai.ImageQualityMedium:
		return goopenai.ImageEditParamsQualityMedium
	case contractsai.ImageQualityHigh:
		return goopenai.ImageEditParamsQualityHigh
	default:
		return ""
	}
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
		if stored, ok := attachment.(contractsai.ProviderFile); ok && stored.ID() != "" {
			switch attachment.Kind() {
			case contractsai.AttachmentKindImage:
				parts = append(parts, responses.ResponseInputContentUnionParam{OfInputImage: &responses.ResponseInputImageParam{
					Detail: responses.ResponseInputImageDetailAuto,
					FileID: param.NewOpt(stored.ID()),
				}})
			case contractsai.AttachmentKindFile:
				parts = append(parts, responses.ResponseInputContentUnionParam{OfInputFile: &responses.ResponseInputFileParam{
					FileID: param.NewOpt(stored.ID()),
				}})
			default:
				return responses.ResponseInputItemUnionParam{}, errors.AIUnsupportedAttachmentKind.Args(attachment.Kind())
			}
			continue
		}

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

func (r *Provider) parseUsage(raw responses.ResponseUsage) contractsai.Usage {
	return frameworkai.NewUsage(int(raw.InputTokens), int(raw.OutputTokens), int(raw.TotalTokens))
}

func (r *Provider) parseImageResponse(response *goopenai.ImagesResponse) (contractsai.ImageResponse, error) {
	if response == nil || len(response.Data) == 0 {
		return nil, errors.AIImageResponseIsEmpty
	}

	content, err := r.resolveImageContent(response.Data[0])
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, errors.AIImageResponseIsEmpty
	}

	mimeType := r.resolveImageMimeType(response.OutputFormat)
	if mimeType == "" {
		mimeType = "image/png"
	}

	return frameworkai.NewImageResponse(
		content,
		mimeType,
		frameworkai.NewUsage(int(response.Usage.InputTokens), int(response.Usage.OutputTokens), int(response.Usage.TotalTokens)),
	), nil
}

func (r *Provider) resolveImageContent(image goopenai.Image) ([]byte, error) {
	if image.B64JSON == "" {
		return nil, errors.AIImageResponseIsEmpty
	}

	content, err := base64.StdEncoding.DecodeString(image.B64JSON)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (r *Provider) resolveImageMimeType(format goopenai.ImagesResponseOutputFormat) string {
	switch format {
	case goopenai.ImagesResponseOutputFormatJPEG:
		return "image/jpeg"
	case goopenai.ImagesResponseOutputFormatWebP:
		return "image/webp"
	case goopenai.ImagesResponseOutputFormatPNG:
		return "image/png"
	default:
		return ""
	}
}

func (r *Provider) parseAudioResponse(response *http.Response, format goopenai.AudioSpeechNewParamsResponseFormat) (contractsai.AudioResponse, error) {
	if response == nil || response.Body == nil {
		return nil, errors.AIAudioResponseIsEmpty
	}
	defer errors.Ignore(response.Body.Close)

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, errors.AIAudioResponseIsEmpty
	}

	mimeType := response.Header.Get("Content-Type")
	if mediaType, _, err := mime.ParseMediaType(mimeType); err == nil && mediaType != "" {
		mimeType = mediaType
	}
	if mimeType == "" || mimeType == "text/plain" || mimeType == "application/octet-stream" {
		mimeType = r.resolveAudioMimeType(format)
	}
	if mimeType == "" {
		mimeType = "audio/mpeg"
	}

	return frameworkai.NewAudioResponse(content, mimeType, frameworkai.NewUsage(0, 0, 0)), nil
}

func (r *Provider) parseTranscriptionResponse(response *goopenai.AudioTranscriptionNewResponseUnion, diarize bool) (contractsai.TranscriptionResponse, error) {
	if response == nil {
		return nil, errors.AITranscriptionResponseIsEmpty
	}

	segments, err := r.resolveTranscriptionSegments(response, diarize)
	if err != nil {
		return nil, err
	}
	if !response.JSON.Text.Valid() && !response.JSON.Segments.Valid() {
		return nil, errors.AITranscriptionResponseIsEmpty
	}

	return frameworkai.NewTranscriptionResponse(response.Text, segments, r.parseTranscriptionUsage(response.Usage)), nil
}

func (r *Provider) resolveTranscriptionSegments(response *goopenai.AudioTranscriptionNewResponseUnion, diarize bool) ([]contractsai.TranscriptionSegment, error) {
	if response == nil {
		return nil, nil
	}

	if diarize {
		type rawSegment struct {
			Speaker string  `json:"speaker"`
			Start   float64 `json:"start"`
			End     float64 `json:"end"`
			Text    string  `json:"text"`
		}
		type rawResponse struct {
			Segments []rawSegment `json:"segments"`
		}

		var raw rawResponse
		if err := json.Unmarshal([]byte(response.RawJSON()), &raw); err != nil {
			return nil, err
		}
		if len(raw.Segments) > 0 {
			segments := make([]contractsai.TranscriptionSegment, 0, len(raw.Segments))
			for _, segment := range raw.Segments {
				segments = append(segments, contractsai.TranscriptionSegment{
					Speaker: segment.Speaker,
					Start:   time.Duration(segment.Start * float64(time.Second)),
					End:     time.Duration(segment.End * float64(time.Second)),
					Text:    segment.Text,
				})
			}

			return segments, nil
		}
	}

	if len(response.Segments) == 0 {
		return nil, nil
	}

	segments := make([]contractsai.TranscriptionSegment, 0, len(response.Segments))
	for _, segment := range response.Segments {
		segments = append(segments, contractsai.TranscriptionSegment{
			Start: time.Duration(segment.Start * float64(time.Second)),
			End:   time.Duration(segment.End * float64(time.Second)),
			Text:  segment.Text,
		})
	}

	return segments, nil
}

func (r *Provider) parseTranscriptionUsage(raw goopenai.AudioTranscriptionNewResponseUnionUsage) contractsai.Usage {
	return frameworkai.NewUsage(int(raw.InputTokens), int(raw.OutputTokens), int(raw.TotalTokens))
}

func (r *Provider) resolveAudioMimeType(format goopenai.AudioSpeechNewParamsResponseFormat) string {
	switch format {
	case goopenai.AudioSpeechNewParamsResponseFormatWAV:
		return "audio/wav"
	case goopenai.AudioSpeechNewParamsResponseFormatFLAC:
		return "audio/flac"
	case goopenai.AudioSpeechNewParamsResponseFormatAAC:
		return "audio/aac"
	case goopenai.AudioSpeechNewParamsResponseFormatOpus:
		return "audio/opus"
	case goopenai.AudioSpeechNewParamsResponseFormatPCM:
		return "audio/pcm"
	case goopenai.AudioSpeechNewParamsResponseFormatMP3:
		fallthrough
	default:
		return "audio/mpeg"
	}
}

func isNilInterface(value any) bool {
	if value == nil {
		return true
	}

	reflectValue := reflect.ValueOf(value)
	switch reflectValue.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflectValue.IsNil()
	default:
		return false
	}
}
