package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	goopenai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
	mocksai "github.com/goravel/framework/mocks/ai"
	mocksconfig "github.com/goravel/framework/mocks/config"
)

type capturedRequest struct {
	path          string
	authorization string
	model         string
	instructions  string
	body          map[string]any
}

type normalizedCapturedRequest struct {
	path          string
	authorization string
	model         string
	instructions  string
	messages      []normalizedMessage
}

type normalizedMessage struct {
	role    string
	content string
}

type capturedStreamRequest struct {
	path          string
	authorization string
	model         string
	instructions  string
	body          map[string]any
	stream        bool
}

type normalizedCapturedStreamRequest struct {
	path          string
	authorization string
	model         string
	instructions  string
	messages      []normalizedMessage
	stream        bool
}

type usageCheck struct {
	input  int
	output int
	total  int
}

type capturedFileUploadRequest struct {
	path          string
	authorization string
	filename      string
	mimeType      string
	purpose       string
	body          []byte
}

type capturedImageRequest struct {
	path          string
	authorization string
	contentType   string
	body          []byte
	formValues    map[string]string
	files         []capturedImageFile
}

type capturedAudioRequest struct {
	path          string
	authorization string
	contentType   string
	body          map[string]any
	accept        string
}

type capturedTranscriptionRequest struct {
	path          string
	authorization string
	contentType   string
	formValues    map[string]string
	fileName      string
	fileMimeType  string
	fileBody      []byte
}

type capturedImageFile struct {
	fieldName string
	fileName  string
	mimeType  string
	body      []byte
}

type nilNamedAttachment struct{}

type streamUsageSnapshot struct {
	input  int
	output int
	total  int
}

type normalizedStreamEvent struct {
	eventType contractsai.StreamEventType
	delta     string
	err       string
	usage     *streamUsageSnapshot
	toolCalls []contractsai.ToolCall
}

func TestNewOpenAIUnmarshalError(t *testing.T) {
	var mockConfig *mocksconfig.Config

	beforeEach := func() {
		mockConfig = mocksconfig.NewConfig(t)
	}

	tests := []struct {
		name         string
		setup        func()
		expectConfig *contractsai.ProviderConfig
		expectErr    error
	}{
		{
			name: "returns unmarshal error",
			setup: func() {
				mockConfig.EXPECT().UnmarshalKey("ai.providers.openai", new(contractsai.ProviderConfig)).Return(assert.AnError).Once()
			},
			expectErr: assert.AnError,
		},
		{
			name: "sets default text model",
			setup: func() {
				mockConfig.EXPECT().UnmarshalKey("ai.providers.openai", new(contractsai.ProviderConfig)).RunAndReturn(func(_ string, rawVal any) error {
					cfg := rawVal.(*contractsai.ProviderConfig)
					cfg.Key = "test-key"
					cfg.Url = "http://localhost:1234"
					return nil
				}).Once()
			},
			expectConfig: func() *contractsai.ProviderConfig {
				cfg := contractsai.ProviderConfig{Key: "test-key", Url: "http://localhost:1234"}
				cfg.Models.Text.Default = DefaultTextModel
				cfg.Models.Audio.Default = DefaultAudioModel
				cfg.Models.Transcription.Default = DefaultTranscriptionModel
				cfg.Models.Image.Default = DefaultImageModel
				return &cfg
			}(),
		},
		{
			name: "keeps configured default models",
			setup: func() {
				mockConfig.EXPECT().UnmarshalKey("ai.providers.openai", new(contractsai.ProviderConfig)).RunAndReturn(func(_ string, rawVal any) error {
					cfg := rawVal.(*contractsai.ProviderConfig)
					cfg.Key = "test-key"
					cfg.Models.Text.Default = "gpt-custom"
					cfg.Models.Image.Default = "gpt-image-custom"
					return nil
				}).Once()
			},
			expectConfig: func() *contractsai.ProviderConfig {
				cfg := contractsai.ProviderConfig{Key: "test-key"}
				cfg.Models.Text.Default = "gpt-custom"
				cfg.Models.Audio.Default = DefaultAudioModel
				cfg.Models.Transcription.Default = DefaultTranscriptionModel
				cfg.Models.Image.Default = "gpt-image-custom"
				return &cfg
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()

			provider, err := NewOpenAI(mockConfig, "openai")

			assert.Equal(t, tt.expectErr, err)
			if tt.expectErr != nil {
				assert.Nil(t, provider)
				return
			}
			require.NotNil(t, provider)
			assert.Equal(t, *tt.expectConfig, provider.config)
		})
	}
}

func TestProviderResolveTranscriptionModel(t *testing.T) {
	tests := []struct {
		name         string
		defaultModel string
		model        string
		diarize      bool
		expectModel  string
	}{
		{
			name:         "uses explicit model",
			defaultModel: DefaultTranscriptionModel,
			model:        "gpt-custom",
			diarize:      true,
			expectModel:  "gpt-custom",
		},
		{
			name:         "uses diarized fallback for zero config",
			defaultModel: DefaultTranscriptionModel,
			diarize:      true,
			expectModel:  DefaultDiarizedTranscriptionModel,
		},
		{
			name:         "uses custom configured default for diarized request",
			defaultModel: "gpt-transcription-custom",
			diarize:      true,
			expectModel:  "gpt-transcription-custom",
		},
		{
			name:        "uses default transcription model when unset",
			diarize:     false,
			expectModel: DefaultTranscriptionModel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{}
			provider.config.Models.Transcription.Default = tt.defaultModel

			assert.Equal(t, tt.expectModel, provider.resolveTranscriptionModel(tt.model, tt.diarize))
		})
	}
}

func TestProviderImage(t *testing.T) {
	tests := []struct {
		name          string
		prompt        contractsai.ImagePrompt
		response      string
		status        int
		expectError   error
		expectPath    string
		expectForm    map[string]string
		expectFiles   []capturedImageFile
		expectMime    string
		expectContent []byte
	}{
		{
			name: "generates image with defaults",
			prompt: contractsai.ImagePrompt{
				Prompt: "draw a cat",
			},
			status:     http.StatusOK,
			response:   imageResponseBody(t, "png", "image-bytes", 11, 7, 18),
			expectPath: "/images/generations",
			expectForm: map[string]string{
				"prompt": "draw a cat",
				"model":  "gpt-image-default",
			},
			expectMime:    "image/png",
			expectContent: []byte("image-bytes"),
		},
		{
			name: "uses explicit quality size and timeout",
			prompt: contractsai.ImagePrompt{
				Prompt:  "draw a cat",
				Model:   "gpt-image-override",
				Size:    contractsai.ImageSizeLandscape,
				Quality: contractsai.ImageQualityHigh,
				Timeout: 2 * time.Second,
			},
			status:     http.StatusOK,
			response:   imageResponseBody(t, "jpeg", "jpeg-bytes", 1, 2, 3),
			expectPath: "/images/generations",
			expectForm: map[string]string{
				"prompt":  "draw a cat",
				"model":   "gpt-image-override",
				"size":    "1536x1024",
				"quality": "high",
			},
			expectMime:    "image/jpeg",
			expectContent: []byte("jpeg-bytes"),
		},
		{
			name: "edits image when attachments provided",
			prompt: contractsai.ImagePrompt{
				Prompt:      "turn this into watercolor",
				Size:        contractsai.ImageSizePortrait,
				Quality:     contractsai.ImageQualityMedium,
				Attachments: []contractsai.Attachment{namedAttachment{kind: contractsai.AttachmentKindImage, filename: "photo.png", mimeType: "image/png", content: []byte("source-image")}},
			},
			status:     http.StatusOK,
			response:   imageResponseBody(t, "webp", "webp-bytes", 4, 5, 9),
			expectPath: "/images/edits",
			expectForm: map[string]string{
				"prompt":  "turn this into watercolor",
				"model":   "gpt-image-default",
				"size":    "1024x1536",
				"quality": "medium",
			},
			expectFiles:   []capturedImageFile{{fieldName: "image[]", fileName: "photo.png", mimeType: "image/png", body: []byte("source-image")}},
			expectMime:    "image/webp",
			expectContent: []byte("webp-bytes"),
		},
		{
			name:        "returns error for empty prompt",
			prompt:      contractsai.ImagePrompt{},
			expectError: errors.AIImagePromptRequired,
		},
		{
			name: "returns error for non image attachment",
			prompt: contractsai.ImagePrompt{
				Prompt:      "draw a cat",
				Attachments: []contractsai.Attachment{namedAttachment{kind: contractsai.AttachmentKindFile, filename: "report.txt", mimeType: "text/plain", content: []byte("report")}},
			},
			expectError: errors.AIImageAttachmentRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := make(chan capturedImageRequest, 1)
			server := newImagesServer(t, tt.status, tt.response, captured)
			defer server.Close()

			provider := &Provider{client: goopenai.NewClient(option.WithAPIKey("test-key"), option.WithBaseURL(server.URL))}
			provider.config.Models.Image.Default = "gpt-image-default"

			response, err := provider.Image(context.Background(), tt.prompt)
			assert.Equal(t, tt.expectError, err)
			if tt.expectError != nil {
				assert.Nil(t, response)
				return
			}

			require.NotNil(t, response)
			content, contentErr := response.Content()
			require.NoError(t, contentErr)
			assert.Equal(t, tt.expectContent, content)
			assert.Equal(t, tt.expectMime, response.MimeType())

			req, ok := readCapturedImageRequest(t, captured)
			require.True(t, ok, "expected image request payload")
			assert.Equal(t, tt.expectPath, req.path)
			assert.Equal(t, "Bearer test-key", req.authorization)
			assert.Equal(t, tt.expectForm, req.formValues)
			assert.Equal(t, tt.expectFiles, req.files)
		})
	}
}

func TestProviderAudio(t *testing.T) {
	tests := []struct {
		name          string
		prompt        contractsai.AudioPrompt
		status        int
		responseBody  string
		responseType  string
		expectError   error
		expectPath    string
		expectBody    map[string]any
		expectMime    string
		expectContent []byte
	}{
		{
			name: "generates audio with defaults",
			prompt: contractsai.AudioPrompt{
				Prompt: "welcome to goravel",
			},
			status:       http.StatusOK,
			responseBody: "audio-bytes",
			expectPath:   "/audio/speech",
			expectBody: map[string]any{
				"input":           "welcome to goravel",
				"model":           "gpt-audio-default",
				"voice":           "alloy",
				"response_format": "mp3",
			},
			expectMime:    "audio/mpeg",
			expectContent: []byte("audio-bytes"),
		},
		{
			name: "uses explicit voice instructions and timeout",
			prompt: contractsai.AudioPrompt{
				Prompt:       "welcome to goravel",
				Model:        "gpt-4o-mini-tts",
				Voice:        "default-male",
				Instructions: "Speak slowly",
				Timeout:      2 * time.Second,
			},
			status:       http.StatusOK,
			responseBody: "audio-bytes",
			responseType: "audio/mpeg; charset=utf-8",
			expectPath:   "/audio/speech",
			expectBody: map[string]any{
				"input":           "welcome to goravel",
				"model":           "gpt-4o-mini-tts",
				"voice":           "ash",
				"instructions":    "Speak slowly",
				"response_format": "mp3",
			},
			expectMime:    "audio/mpeg",
			expectContent: []byte("audio-bytes"),
		},
		{
			name:        "returns error for empty prompt",
			prompt:      contractsai.AudioPrompt{},
			expectError: errors.AIAudioPromptRequired,
		},
		{
			name: "returns error for empty response body",
			prompt: contractsai.AudioPrompt{
				Prompt: "welcome to goravel",
			},
			status:       http.StatusOK,
			responseBody: "",
			expectError:  errors.AIAudioResponseIsEmpty,
		},
		{
			name: "uses response format mime fallback",
			prompt: contractsai.AudioPrompt{
				Prompt: "welcome to goravel",
			},
			status:       http.StatusOK,
			responseBody: "audio-bytes",
			responseType: "",
			expectPath:   "/audio/speech",
			expectBody: map[string]any{
				"input":           "welcome to goravel",
				"model":           "gpt-audio-default",
				"voice":           "alloy",
				"response_format": "mp3",
			},
			expectMime:    "audio/mpeg",
			expectContent: []byte("audio-bytes"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := make(chan capturedAudioRequest, 1)
			server := newAudioServer(t, tt.status, tt.responseBody, tt.responseType, captured)
			defer server.Close()

			provider := &Provider{client: goopenai.NewClient(option.WithAPIKey("test-key"), option.WithBaseURL(server.URL))}
			provider.config.Models.Audio.Default = "gpt-audio-default"

			response, err := provider.Audio(context.Background(), tt.prompt)
			assert.Equal(t, tt.expectError, err)
			if tt.expectError != nil {
				assert.Nil(t, response)
				return
			}

			require.NotNil(t, response)
			content, contentErr := response.Content()
			require.NoError(t, contentErr)
			assert.Equal(t, tt.expectContent, content)
			assert.Equal(t, tt.expectMime, response.MimeType())

			req, ok := readCapturedAudioRequest(t, captured)
			require.True(t, ok, "expected audio request payload")
			assert.Equal(t, tt.expectPath, req.path)
			assert.Equal(t, "Bearer test-key", req.authorization)
			assert.Equal(t, "application/octet-stream", req.accept)
			assert.Equal(t, tt.expectBody, req.body)
		})
	}
}

func TestProviderTranscription(t *testing.T) {
	tests := []struct {
		name          string
		prompt        contractsai.TranscriptionPrompt
		status        int
		responseBody  string
		expectError   error
		expectPath    string
		expectForm    map[string]string
		expectFile    capturedImageFile
		expectText    string
		expectUsage   usageCheck
		expectSegment []contractsai.TranscriptionSegment
	}{
		{
			name: "transcribes audio with defaults",
			prompt: contractsai.TranscriptionPrompt{
				File: namedAttachment{kind: contractsai.AttachmentKindFile, filename: "call.mp3", mimeType: "audio/mpeg", content: []byte("audio")},
			},
			status:       http.StatusOK,
			responseBody: transcriptionResponseBody(t, "hello world", nil, map[string]any{"type": "tokens", "input_tokens": 3, "output_tokens": 4, "total_tokens": 7}),
			expectPath:   "/audio/transcriptions",
			expectForm: map[string]string{
				"model":           "gpt-transcription-default",
				"response_format": "json",
			},
			expectFile:  capturedImageFile{fieldName: "file", fileName: "call.mp3", mimeType: "audio/mpeg", body: []byte("audio")},
			expectText:  "hello world",
			expectUsage: usageCheck{input: 3, output: 4, total: 7},
		},
		{
			name: "uses diarized response format and language",
			prompt: contractsai.TranscriptionPrompt{
				File:     namedAttachment{kind: contractsai.AttachmentKindFile, filename: "call.mp3", mimeType: "audio/mpeg", content: []byte("audio")},
				Model:    "gpt-4o-transcribe-diarize",
				Language: "en",
				Diarize:  true,
				Timeout:  2 * time.Second,
			},
			status: http.StatusOK,
			responseBody: transcriptionResponseBody(t, "hello there", []map[string]any{{
				"speaker": "speaker_0",
				"start":   0.1,
				"end":     1.2,
				"text":    "hello there",
			}}, map[string]any{"type": "duration", "seconds": 1.2}),
			expectPath: "/audio/transcriptions",
			expectForm: map[string]string{
				"model":           "gpt-4o-transcribe-diarize",
				"language":        "en",
				"response_format": "diarized_json",
			},
			expectFile: capturedImageFile{fieldName: "file", fileName: "call.mp3", mimeType: "audio/mpeg", body: []byte("audio")},
			expectText: "hello there",
			expectSegment: []contractsai.TranscriptionSegment{{
				Speaker: "speaker_0",
				Start:   100 * time.Millisecond,
				End:     1200 * time.Millisecond,
				Text:    "hello there",
			}},
		},
		{
			name:        "returns error for nil file",
			prompt:      contractsai.TranscriptionPrompt{},
			expectError: errors.AITranscriptionFileRequired,
		},
		{
			name: "returns error for typed nil file",
			prompt: contractsai.TranscriptionPrompt{
				File: (*nilNamedAttachment)(nil),
			},
			expectError: errors.AITranscriptionFileRequired,
		},
		{
			name: "returns error for empty response",
			prompt: contractsai.TranscriptionPrompt{
				File: namedAttachment{kind: contractsai.AttachmentKindFile, filename: "call.mp3", mimeType: "audio/mpeg", content: []byte("audio")},
			},
			status:       http.StatusOK,
			responseBody: `{}`,
			expectError:  errors.AITranscriptionResponseIsEmpty,
		},
		{
			name: "allows empty transcript text",
			prompt: contractsai.TranscriptionPrompt{
				File: namedAttachment{kind: contractsai.AttachmentKindFile, filename: "call.mp3", mimeType: "audio/mpeg", content: []byte("audio")},
			},
			status:       http.StatusOK,
			responseBody: `{"text":""}`,
			expectPath:   "/audio/transcriptions",
			expectForm: map[string]string{
				"model":           "gpt-transcription-default",
				"response_format": "json",
			},
			expectFile: capturedImageFile{fieldName: "file", fileName: "call.mp3", mimeType: "audio/mpeg", body: []byte("audio")},
			expectText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := make(chan capturedTranscriptionRequest, 1)
			server := newTranscriptionServer(t, tt.status, tt.responseBody, captured)
			defer server.Close()

			provider := &Provider{client: goopenai.NewClient(option.WithAPIKey("test-key"), option.WithBaseURL(server.URL))}
			provider.config.Models.Transcription.Default = "gpt-transcription-default"

			response, err := provider.Transcription(context.Background(), tt.prompt)
			assert.Equal(t, tt.expectError, err)
			if tt.expectError != nil {
				assert.Nil(t, response)
				return
			}

			require.NotNil(t, response)
			assert.Equal(t, tt.expectText, response.Text())
			assert.Equal(t, tt.expectSegment, response.Segments())
			if response.Usage() != nil {
				assert.Equal(t, tt.expectUsage, usageCheck{input: response.Usage().Input(), output: response.Usage().Output(), total: response.Usage().Total()})
			}

			req, ok := readCapturedTranscriptionRequest(t, captured)
			require.True(t, ok, "expected transcription request payload")
			assert.Equal(t, tt.expectPath, req.path)
			assert.Equal(t, "Bearer test-key", req.authorization)
			assert.Equal(t, tt.expectForm, req.formValues)
			assert.Equal(t, tt.expectFile.fileName, req.fileName)
			assert.Equal(t, tt.expectFile.mimeType, req.fileMimeType)
			assert.Equal(t, tt.expectFile.body, req.fileBody)
		})
	}
}

func TestProviderPrompt(t *testing.T) {
	type usageCheck struct {
		input  int
		output int
		total  int
	}

	var mockAgent *mocksai.Agent

	beforeEach := func() {
		mockAgent = mocksai.NewAgent(t)
	}

	tests := []struct {
		name             string
		status           int
		body             string
		responses        []string
		repeatLastBody   bool
		setup            func()
		modelOverride    string
		input            string
		expectText       string
		expectUsage      usageCheck
		expectErr        bool
		expectErrMessage string
		expectRequest    normalizedCapturedRequest
	}{
		{
			name:   "builds input with default model",
			status: http.StatusOK,
			body:   responseBody(t, "assistant reply", nil, 11, 7, 18),
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("system rule").Once()
				mockAgent.EXPECT().Messages().Return([]contractsai.Message{
					{Role: contractsai.RoleUser, Content: "history user"},
					{Role: contractsai.RoleAssistant, Content: "history assistant"},
				}).Once()
			},
			input:       "new input",
			expectText:  "assistant reply",
			expectUsage: usageCheck{input: 11, output: 7, total: 18},
			expectRequest: normalizedCapturedRequest{
				path:          "/responses",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				instructions:  "system rule",
				messages: []normalizedMessage{
					{role: "user", content: "history user"},
					{role: "assistant", content: "history assistant"},
					{role: "user", content: "new input"},
				},
			},
		},
		{
			name:   "uses prompt model override",
			status: http.StatusOK,
			body:   responseBody(t, "ok", nil, 1, 1, 2),
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			modelOverride: "gpt-override",
			input:         "hello",
			expectText:    "ok",
			expectUsage:   usageCheck{input: 1, output: 1, total: 2},
			expectRequest: normalizedCapturedRequest{
				path:          "/responses",
				authorization: "Bearer test-key",
				model:         "gpt-override",
				messages: []normalizedMessage{
					{role: "user", content: "hello"},
				},
			},
		},
		{
			name:   "returns error when API fails",
			status: http.StatusInternalServerError,
			body:   `{"error":{"message":"boom","type":"server_error"}}`,
			responses: []string{
				`{"error":{"message":"boom","type":"server_error"}}`,
				`{"error":{"message":"boom","type":"server_error"}}`,
			},
			repeatLastBody: true,
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			input:            "hello",
			expectErr:        true,
			expectErrMessage: "boom",
			expectRequest: normalizedCapturedRequest{
				path:          "/responses",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				messages:      []normalizedMessage{{role: "user", content: "hello"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()

			captured := make(chan capturedRequest, 1)
			responses := tt.responses
			if responses == nil {
				responses = bodySequence(tt.body)
			}
			server := newResponsesServerWithOverflow(t, tt.status, responses, captured, tt.repeatLastBody)
			t.Cleanup(server.Close)

			provider := &Provider{
				client: goopenai.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("test-key")),
				config: contractsai.ProviderConfig{},
			}
			provider.config.Models.Text.Default = "gpt-default"

			tt.setup()

			prompt := contractsai.AgentPrompt{Agent: mockAgent, Input: tt.input}
			if tt.modelOverride != "" {
				prompt.Model = tt.modelOverride
			}

			resp, err := provider.Prompt(context.Background(), prompt)

			if tt.expectErr {
				assert.Nil(t, resp)
				require.Error(t, err)

				var apiErr *goopenai.Error
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, tt.status, apiErr.StatusCode)
				assert.Equal(t, tt.expectErrMessage, apiErr.Message)

				req, ok := readCapturedRequest(t, captured)
				require.True(t, ok, "expected request payload")
				assert.Equal(t, tt.expectRequest, normalizeCapturedRequest(req))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.expectText, resp.Text())
			require.NotNil(t, resp.Usage())
			assert.Equal(t, tt.expectUsage, usageCheck{
				input:  resp.Usage().Input(),
				output: resp.Usage().Output(),
				total:  resp.Usage().Total(),
			})

			req, ok := readCapturedRequest(t, captured)
			require.True(t, ok, "expected request payload")
			assert.Equal(t, tt.expectRequest, normalizeCapturedRequest(req))
		})
	}
}

func TestProviderBuildInputWithAttachments(t *testing.T) {
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Instructions().Return("").Once()
	mockAgent.EXPECT().Messages().Return([]contractsai.Message{{Role: contractsai.RoleUser, Content: "history"}}).Once()

	provider := &Provider{}
	input, _, _, err := provider.buildInput(context.Background(), contractsai.AgentPrompt{
		Agent: mockAgent,
		Input: "describe these",
		Attachments: []contractsai.Attachment{
			frameworkai.ImageFromByte([]byte("image"), frameworkai.WithMimeType("image/png")),
			namedAttachment{kind: contractsai.AttachmentKindFile, filename: "report.txt", mimeType: "text/plain", content: []byte("document")},
		},
	})
	require.NoError(t, err)
	require.Len(t, input, 2)

	content := marshalInputContent(t, input[1])
	require.Len(t, content, 3)
	assert.Equal(t, map[string]any{"text": "describe these", "type": "input_text"}, content[0])
	assert.Equal(t, "input_image", content[1]["type"])
	assert.Equal(t, "data:image/png;base64,aW1hZ2U=", content[1]["image_url"])
	assert.Equal(t, map[string]any{"text": "File: report.txt\n\ndocument", "type": "input_text"}, content[2])
}

func TestProviderBuildInputAttachesToActiveUserTurnOnFollowUp(t *testing.T) {
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Instructions().Return("").Once()
	mockAgent.EXPECT().Messages().Return([]contractsai.Message{
		{Role: contractsai.RoleUser, Content: "question"},
		{Role: contractsai.RoleAssistant, ToolCalls: []contractsai.ToolCall{{ID: "call-1", Name: "lookup"}}},
		{Role: contractsai.RoleToolResult, Content: "result", ToolCallID: "call-1"},
	}).Once()

	provider := &Provider{}
	input, _, _, err := provider.buildInput(context.Background(), contractsai.AgentPrompt{
		Agent:       mockAgent,
		Attachments: []contractsai.Attachment{namedAttachment{kind: contractsai.AttachmentKindFile, filename: "report.txt", content: []byte("document")}},
	})
	require.NoError(t, err)
	require.Len(t, input, 3)

	content := marshalInputContent(t, input[0])
	require.Len(t, content, 2)
	assert.Equal(t, map[string]any{"text": "question", "type": "input_text"}, content[0])
	assert.Equal(t, map[string]any{"text": "File: report.txt\n\ndocument", "type": "input_text"}, content[1])
}

func TestProviderBuildInputKeepsBinaryFileAttachmentsAsInputFiles(t *testing.T) {
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Instructions().Return("").Once()
	mockAgent.EXPECT().Messages().Return(nil).Once()

	provider := &Provider{}
	input, _, _, err := provider.buildInput(context.Background(), contractsai.AgentPrompt{
		Agent: mockAgent,
		Attachments: []contractsai.Attachment{
			namedAttachment{kind: contractsai.AttachmentKindFile, filename: "report.pdf", mimeType: "application/pdf", content: []byte("%PDF")},
		},
	})
	require.NoError(t, err)
	require.Len(t, input, 1)

	content := marshalInputContent(t, input[0])
	require.Len(t, content, 1)
	assert.Equal(t, "input_file", content[0]["type"])
	assert.Equal(t, "data:application/pdf;base64,JVBERg==", content[0]["file_data"])
	assert.Equal(t, "report.pdf", content[0]["filename"])
}

func TestProviderBuildInputUsesStoredFileIDReferences(t *testing.T) {
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Instructions().Return("").Once()
	mockAgent.EXPECT().Messages().Return(nil).Once()

	provider := &Provider{}
	input, _, _, err := provider.buildInput(context.Background(), contractsai.AgentPrompt{
		Agent: mockAgent,
		Attachments: []contractsai.Attachment{
			frameworkai.ImageFromID("image-123"),
			frameworkai.DocumentFromID("file-123"),
		},
	})
	require.NoError(t, err)
	require.Len(t, input, 1)

	content := marshalInputContent(t, input[0])
	require.Len(t, content, 2)
	assert.Equal(t, "input_image", content[0]["type"])
	assert.Equal(t, "image-123", content[0]["file_id"])
	assert.Equal(t, "input_file", content[1]["type"])
	assert.Equal(t, "file-123", content[1]["file_id"])
}

func TestProviderBuildInputDoesNotTreatEmptyAttachmentPromptAsToolContinuation(t *testing.T) {
	state := &providerStateStub{data: map[string]any{providerStateResponseID: "resp_prev"}}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Instructions().Return("").Once()
	mockAgent.EXPECT().Messages().Return([]contractsai.Message{{Role: contractsai.RoleUser, Content: "history"}}).Twice()

	provider := &Provider{}
	input, _, previousResponseID, err := provider.buildInput(context.Background(), contractsai.AgentPrompt{
		Agent:         mockAgent,
		Attachments:   []contractsai.Attachment{namedAttachment{kind: contractsai.AttachmentKindFile, filename: "report.txt", mimeType: "text/plain", content: []byte("document")}},
		ProviderState: state,
	})
	require.NoError(t, err)
	assert.Empty(t, previousResponseID)
	require.Len(t, input, 1)

	content := marshalInputContent(t, input[0])
	require.Len(t, content, 2)
	assert.Equal(t, map[string]any{"text": "history", "type": "input_text"}, content[0])
	assert.Equal(t, map[string]any{"text": "File: report.txt\n\ndocument", "type": "input_text"}, content[1])
}

func TestProviderBuildInputIgnoresAttachmentsDuringToolContinuation(t *testing.T) {
	state := &providerStateStub{data: map[string]any{providerStateResponseID: "resp_prev"}}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Instructions().Return("").Once()
	mockAgent.EXPECT().Messages().Return([]contractsai.Message{
		{Role: contractsai.RoleUser, Content: "question"},
		{Role: contractsai.RoleAssistant, ToolCalls: []contractsai.ToolCall{{ID: "call-1", Name: "lookup"}}},
		{Role: contractsai.RoleToolResult, Content: "result", ToolCallID: "call-1"},
	}).Once()

	provider := &Provider{}
	input, _, previousResponseID, err := provider.buildInput(context.Background(), contractsai.AgentPrompt{
		Agent:         mockAgent,
		Attachments:   []contractsai.Attachment{namedAttachment{kind: contractsai.AttachmentKindFile, filename: "report.txt", content: []byte("document")}},
		ProviderState: state,
	})
	require.NoError(t, err)
	assert.Equal(t, "resp_prev", previousResponseID)
	require.Len(t, input, 1)
	assert.Equal(t, []normalizedMessage{{role: "tool", content: "result"}}, normalizeInputItems([]map[string]any{marshalInputItem(t, input[0])}))
}

func TestProviderPromptAndStreamSerializeAttachmentsTheSameWay(t *testing.T) {
	attachments := []contractsai.Attachment{
		frameworkai.ImageFromByte([]byte("image"), frameworkai.WithMimeType("image/png")),
		namedAttachment{kind: contractsai.AttachmentKindFile, filename: "report.txt", mimeType: "text/plain", content: []byte("document")},
	}

	promptAgent := mocksai.NewAgent(t)
	promptAgent.EXPECT().Instructions().Return("").Once()
	promptAgent.EXPECT().Messages().Return(nil).Once()

	streamAgent := mocksai.NewAgent(t)
	streamAgent.EXPECT().Instructions().Return("").Once()
	streamAgent.EXPECT().Messages().Return(nil).Once()

	promptCaptured := make(chan capturedRequest, 1)
	promptServer := newResponsesServer(t, http.StatusOK, bodySequence(responseBody(t, "ok", nil, 1, 1, 2)), promptCaptured)
	t.Cleanup(promptServer.Close)

	provider := &Provider{
		client: goopenai.NewClient(option.WithBaseURL(promptServer.URL), option.WithAPIKey("test-key")),
		config: contractsai.ProviderConfig{},
	}
	provider.config.Models.Text.Default = "gpt-default"

	_, err := provider.Prompt(context.Background(), contractsai.AgentPrompt{
		Agent:       promptAgent,
		Input:       "describe these",
		Attachments: attachments,
	})
	require.NoError(t, err)

	promptReq, ok := readCapturedRequest(t, promptCaptured)
	require.True(t, ok)

	streamCaptured := make(chan capturedStreamRequest, 1)
	streamServer := newStreamingResponsesServer(t, http.StatusOK, "text/event-stream", stringsJoinLines(
		`data: {"type":"response.completed","sequence_number":1,"response":{"id":"resp_1","object":"response","model":"gpt-test","output":[{"id":"msg_1","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"ok","annotations":[],"logprobs":[]}]}],"usage":{"input_tokens":1,"input_tokens_details":{"cached_tokens":0},"output_tokens":1,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":2}}}`,
		``,
		`data: [DONE]`,
		``,
	), streamCaptured)
	t.Cleanup(streamServer.Close)

	provider.client = goopenai.NewClient(option.WithBaseURL(streamServer.URL), option.WithAPIKey("test-key"))

	stream, err := provider.Stream(context.Background(), contractsai.AgentPrompt{
		Agent:       streamAgent,
		Input:       "describe these",
		Attachments: attachments,
	})
	require.NoError(t, err)
	require.NotNil(t, stream)
	require.NoError(t, stream.Each(func(contractsai.StreamEvent) error { return nil }))

	streamReq, ok := readCapturedStreamRequest(t, streamCaptured)
	require.True(t, ok)

	assert.Equal(t, inputItemsFromBody(promptReq.body), inputItemsFromBody(streamReq.body))
	assert.True(t, streamReq.stream)
}

func TestProviderBuildInputUnsupportedAttachmentKind(t *testing.T) {
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Instructions().Return("").Once()
	mockAgent.EXPECT().Messages().Return(nil).Once()

	provider := &Provider{}
	input, _, _, err := provider.buildInput(context.Background(), contractsai.AgentPrompt{
		Agent:       mockAgent,
		Attachments: []contractsai.Attachment{unsupportedAttachment{}},
	})

	assert.Nil(t, input)
	assert.Equal(t, errors.AIUnsupportedAttachmentKind.Args(contractsai.AttachmentKind("audio")), err)
}

type unsupportedAttachment struct{}

type providerStateStub struct{ data map[string]any }

func (p *providerStateStub) Get(key string) any { return p.data[key] }
func (p *providerStateStub) Set(key string, value any) {
	if p.data == nil {
		p.data = make(map[string]any)
	}
	p.data[key] = value
}

type namedAttachment struct {
	kind     contractsai.AttachmentKind
	filename string
	mimeType string
	content  []byte
}

func (attachment namedAttachment) Kind() contractsai.AttachmentKind { return attachment.kind }
func (attachment namedAttachment) FileName() string                 { return attachment.filename }
func (attachment namedAttachment) MimeType() string                 { return attachment.mimeType }
func (attachment namedAttachment) Content(context.Context) ([]byte, error) {
	return attachment.content, nil
}
func (attachment namedAttachment) Put(context.Context, ...contractsai.Option) (contractsai.FileResponse, error) {
	return nil, nil
}

func (attachment *nilNamedAttachment) Kind() contractsai.AttachmentKind {
	return contractsai.AttachmentKindFile
}
func (attachment *nilNamedAttachment) FileName() string { return "" }
func (attachment *nilNamedAttachment) MimeType() string { return "" }
func (attachment *nilNamedAttachment) Content(context.Context) ([]byte, error) {
	return nil, nil
}
func (attachment *nilNamedAttachment) Put(context.Context, ...contractsai.Option) (contractsai.FileResponse, error) {
	return nil, nil
}

func (unsupportedAttachment) Kind() contractsai.AttachmentKind {
	return contractsai.AttachmentKind("audio")
}
func (unsupportedAttachment) FileName() string { return "audio.mp3" }
func (unsupportedAttachment) MimeType() string { return "audio/mpeg" }
func (unsupportedAttachment) Content(context.Context) ([]byte, error) {
	return []byte("audio"), nil
}
func (unsupportedAttachment) Put(context.Context, ...contractsai.Option) (contractsai.FileResponse, error) {
	return nil, nil
}

func TestProviderStream(t *testing.T) {
	type usageCheck struct {
		input  int
		output int
		total  int
	}

	var mockAgent *mocksai.Agent
	beforeEach := func() {
		mockAgent = mocksai.NewAgent(t)
	}

	successStreamBody := stringsJoinLines(
		`data: {"type":"response.output_text.delta","item_id":"msg_1","output_index":0,"content_index":0,"delta":"hel","logprobs":[],"sequence_number":1}`,
		``,
		`data: {"type":"response.output_text.delta","item_id":"msg_1","output_index":0,"content_index":0,"delta":"lo","logprobs":[],"sequence_number":2}`,
		``,
		`data: {"type":"response.completed","sequence_number":3,"response":{"id":"resp_1","object":"response","model":"gpt-test","output":[{"id":"msg_1","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"hello","annotations":[],"logprobs":[]}]}],"usage":{"input_tokens":4,"input_tokens_details":{"cached_tokens":0},"output_tokens":2,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":6}}}`,
		``,
		`data: [DONE]`,
		``,
	)

	toolStreamBody := stringsJoinLines(
		`data: {"type":"response.completed","sequence_number":1,"response":{"id":"resp_tc","object":"response","model":"gpt-test","output":[{"id":"fc_1","type":"function_call","call_id":"call_1","name":"get_weather","arguments":"{\"city\":\"London\",\"units\":\"celsius\"}","status":"completed"}],"usage":{"input_tokens":6,"input_tokens_details":{"cached_tokens":0},"output_tokens":3,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":9}}}`,
		``,
		`data: [DONE]`,
		``,
	)

	tests := []struct {
		name             string
		status           int
		contentType      string
		body             string
		setup            func()
		modelOverride    string
		input            string
		expectEachErr    bool
		expectErrMessage string
		expectText       string
		expectUsage      usageCheck
		expectToolCalls  []contractsai.ToolCall
		expectEvents     []normalizedStreamEvent
		expectRequest    normalizedCapturedStreamRequest
	}{
		{
			name:        "streams delta and done events with default model",
			status:      http.StatusOK,
			contentType: "text/event-stream",
			body:        successStreamBody,
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("system rule").Once()
				mockAgent.EXPECT().Messages().Return([]contractsai.Message{
					{Role: contractsai.RoleUser, Content: "history user"},
					{Role: contractsai.RoleAssistant, Content: "history assistant"},
				}).Once()
			},
			input:       "new input",
			expectText:  "hello",
			expectUsage: usageCheck{input: 4, output: 2, total: 6},
			expectEvents: []normalizedStreamEvent{
				{eventType: contractsai.StreamEventTypeTextDelta, delta: "hel"},
				{eventType: contractsai.StreamEventTypeTextDelta, delta: "lo"},
				{eventType: contractsai.StreamEventTypeDone, usage: &streamUsageSnapshot{input: 4, output: 2, total: 6}},
			},
			expectRequest: normalizedCapturedStreamRequest{
				path:          "/responses",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				instructions:  "system rule",
				stream:        true,
				messages: []normalizedMessage{
					{role: "user", content: "history user"},
					{role: "assistant", content: "history assistant"},
					{role: "user", content: "new input"},
				},
			},
		},
		{
			name:        "uses model override for stream",
			status:      http.StatusOK,
			contentType: "text/event-stream",
			body: stringsJoinLines(
				`data: {"type":"response.output_text.delta","item_id":"msg_1","output_index":0,"content_index":0,"delta":"ok","logprobs":[],"sequence_number":1}`,
				``,
				`data: {"type":"response.completed","sequence_number":2,"response":{"id":"resp_2","object":"response","model":"gpt-test","output":[{"id":"msg_1","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"ok","annotations":[],"logprobs":[]}]}],"usage":{"input_tokens":1,"input_tokens_details":{"cached_tokens":0},"output_tokens":1,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":2}}}`,
				``,
				`data: [DONE]`,
				``,
			),
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			modelOverride: "gpt-override",
			input:         "hello",
			expectText:    "ok",
			expectUsage:   usageCheck{input: 1, output: 1, total: 2},
			expectEvents: []normalizedStreamEvent{
				{eventType: contractsai.StreamEventTypeTextDelta, delta: "ok"},
				{eventType: contractsai.StreamEventTypeDone, usage: &streamUsageSnapshot{input: 1, output: 1, total: 2}},
			},
			expectRequest: normalizedCapturedStreamRequest{
				path:          "/responses",
				authorization: "Bearer test-key",
				model:         "gpt-override",
				stream:        true,
				messages:      []normalizedMessage{{role: "user", content: "hello"}},
			},
		},
		{
			name:        "returns tool calls from completed stream response",
			status:      http.StatusOK,
			contentType: "text/event-stream",
			body:        toolStreamBody,
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			input:       "hello",
			expectUsage: usageCheck{input: 6, output: 3, total: 9},
			expectToolCalls: []contractsai.ToolCall{{
				ID:      "call_1",
				Name:    "get_weather",
				Args:    map[string]any{"city": "London", "units": "celsius"},
				RawArgs: `{"city":"London","units":"celsius"}`,
			}},
			expectEvents: []normalizedStreamEvent{{eventType: contractsai.StreamEventTypeDone, usage: &streamUsageSnapshot{input: 6, output: 3, total: 9}}},
			expectRequest: normalizedCapturedStreamRequest{
				path:          "/responses",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				stream:        true,
				messages:      []normalizedMessage{{role: "user", content: "hello"}},
			},
		},
		{
			name:             "emits error event when stream request fails",
			status:           http.StatusInternalServerError,
			contentType:      "application/json",
			body:             `{"error":{"message":"boom","type":"server_error"}}`,
			expectEachErr:    true,
			expectErrMessage: "boom",
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			input: "hello",
			expectRequest: normalizedCapturedStreamRequest{
				path:          "/responses",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				stream:        true,
				messages:      []normalizedMessage{{role: "user", content: "hello"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()

			captured := make(chan capturedStreamRequest, 1)
			server := newStreamingResponsesServer(t, tt.status, tt.contentType, tt.body, captured)
			t.Cleanup(server.Close)

			provider := &Provider{
				client: goopenai.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("test-key")),
				config: contractsai.ProviderConfig{},
			}
			provider.config.Models.Text.Default = "gpt-default"

			tt.setup()

			prompt := contractsai.AgentPrompt{Agent: mockAgent, Input: tt.input}
			if tt.modelOverride != "" {
				prompt.Model = tt.modelOverride
			}

			stream, err := provider.Stream(context.Background(), prompt)
			require.NoError(t, err)
			require.NotNil(t, stream)

			thenCalled := 0
			var thenText string
			var thenUsage usageCheck
			var thenToolCalls []contractsai.ToolCall
			stream.Then(func(resp contractsai.AgentResponse) {
				thenCalled++
				thenText = resp.Text()
				thenToolCalls = resp.ToolCalls()
				if resp.Usage() != nil {
					thenUsage = usageCheck{input: resp.Usage().Input(), output: resp.Usage().Output(), total: resp.Usage().Total()}
				}
			})

			var events []contractsai.StreamEvent
			eachErr := stream.Each(func(event contractsai.StreamEvent) error {
				events = append(events, event)
				return nil
			})
			normalizedEvents := normalizeProviderStreamEvents(events)

			if tt.expectEachErr {
				require.Error(t, eachErr)

				var apiErr *goopenai.Error
				require.ErrorAs(t, eachErr, &apiErr)
				assert.Equal(t, tt.status, apiErr.StatusCode)
				require.Len(t, normalizedEvents, 1)
				assert.Equal(t, contractsai.StreamEventTypeError, normalizedEvents[0].eventType)
				assert.NotEmpty(t, normalizedEvents[0].err)
				assert.Equal(t, 0, thenCalled)
			} else {
				require.NoError(t, eachErr)
				assert.Equal(t, tt.expectEvents, normalizeEmptyToolCalls(normalizedEvents))
				assert.Equal(t, 1, thenCalled)
				assert.Equal(t, tt.expectText, thenText)
				assert.Equal(t, tt.expectUsage, thenUsage)
				assert.Equal(t, tt.expectToolCalls, thenToolCalls)
			}

			req, ok := readCapturedStreamRequest(t, captured)
			require.True(t, ok, "expected stream request payload")
			assert.Equal(t, tt.expectRequest, normalizeCapturedStreamRequest(req))
		})
	}
}

func TestProviderPutFile(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		mimeType       string
		content        []byte
		expectFileName string
		expectMimeType string
	}{
		{
			name:           "uses provided filename",
			fileName:       "report.txt",
			mimeType:       "text/plain",
			content:        []byte("report"),
			expectFileName: "report.txt",
			expectMimeType: "text/plain",
		},
		{
			name:           "uses default filename when empty",
			fileName:       "",
			mimeType:       "application/pdf",
			content:        []byte("%PDF"),
			expectFileName: "attachment.pdf",
			expectMimeType: "application/pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := make(chan capturedFileUploadRequest, 1)
			server := newFileUploadServer(t, `{"id":"file-123"}`, captured)
			defer server.Close()

			provider := &Provider{client: goopenai.NewClient(option.WithAPIKey("test-key"), option.WithBaseURL(server.URL))}
			file := namedAttachment{kind: contractsai.AttachmentKindFile, filename: tt.fileName, mimeType: tt.mimeType, content: tt.content}

			response, err := provider.PutFile(context.Background(), file)
			require.NoError(t, err)
			assert.Equal(t, "file-123", response.ID())
			assert.Empty(t, response.MimeType())
			content, err := response.Content(context.Background())
			require.NoError(t, err)
			assert.Nil(t, content)

			req, ok := readCapturedFileUploadRequest(t, captured)
			require.True(t, ok, "expected upload request payload")
			assert.Equal(t, "/files", req.path)
			assert.Equal(t, "Bearer test-key", req.authorization)
			assert.Equal(t, tt.expectFileName, req.filename)
			assert.Equal(t, tt.expectMimeType, req.mimeType)
			assert.Equal(t, "user_data", req.purpose)
			assert.Equal(t, tt.content, req.body)
		})
	}
}

func TestProviderGetFile(t *testing.T) {
	captured := make(chan capturedRequest, 2)
	server := newResponsesServerWithFileContent(t, captured)
	defer server.Close()

	provider := &Provider{client: goopenai.NewClient(option.WithAPIKey("test-key"), option.WithBaseURL(server.URL))}
	file, err := provider.GetFile(context.Background(), "file-123")
	require.NoError(t, err)
	assert.Equal(t, "file-123", file.ID())
	assert.Equal(t, "text/plain; charset=utf-8", file.MimeType())

	content, err := file.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
}

func TestProviderDeleteFile(t *testing.T) {
	captured := make(chan capturedRequest, 1)
	server := newFileDeleteServer(t, captured)
	defer server.Close()

	provider := &Provider{client: goopenai.NewClient(option.WithAPIKey("test-key"), option.WithBaseURL(server.URL))}
	require.NoError(t, provider.DeleteFile(context.Background(), "file-123"))

	req, ok := readCapturedRequest(t, captured)
	require.True(t, ok)
	assert.Equal(t, "/files/file-123", req.path)
	assert.Equal(t, "Bearer test-key", req.authorization)
}

func newResponsesServer(t *testing.T, status int, responses []string, captured chan<- capturedRequest) *httptest.Server {
	return newResponsesServerWithOverflow(t, status, responses, captured, false)
}

func newResponsesServerWithOverflow(t *testing.T, status int, responses []string, captured chan<- capturedRequest, repeatLastResponse bool) *httptest.Server {
	t.Helper()

	callCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		payload := decodeBodyMap(t, r)
		if payload != nil {
			model, _ := payload["model"].(string)
			instructions, _ := payload["instructions"].(string)
			select {
			case captured <- capturedRequest{path: r.URL.Path, authorization: r.Header.Get("Authorization"), model: model, instructions: instructions, body: payload}:
			default:
			}
		}

		if callCount >= len(responses) {
			if repeatLastResponse && len(responses) > 0 {
				body := responses[len(responses)-1]
				callCount++

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = w.Write([]byte(body))
				return
			}
			t.Fatalf("unexpected extra response request: call %d exceeds configured responses %d", callCount+1, len(responses))
		}
		body := responses[callCount]
		callCount++

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/responses", handler)
	mux.HandleFunc("/v1/responses", handler)

	return httptest.NewServer(mux)
}

func marshalInputContent(t *testing.T, message any) []map[string]any {
	t.Helper()

	data, err := json.Marshal(message)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	content, ok := raw["content"].([]any)
	require.True(t, ok)

	parts := make([]map[string]any, 0, len(content))
	for _, part := range content {
		item, ok := part.(map[string]any)
		require.True(t, ok)
		parts = append(parts, item)
	}

	return parts
}

func marshalInputItem(t *testing.T, item any) map[string]any {
	t.Helper()

	data, err := json.Marshal(item)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))

	return raw
}

func messageText(content any) string {
	switch val := content.(type) {
	case string:
		return val
	case []any:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			part, ok := item.(map[string]any)
			if !ok {
				continue
			}
			text, _ := part["text"].(string)
			if text != "" {
				parts = append(parts, text)
			}
		}
		return stringsJoin(parts)
	default:
		return ""
	}
}

func readCapturedRequest(t *testing.T, captured <-chan capturedRequest) (capturedRequest, bool) {
	t.Helper()
	select {
	case req := <-captured:
		return req, true
	default:
		return capturedRequest{}, false
	}
}

func normalizeCapturedRequest(req capturedRequest) normalizedCapturedRequest {
	return normalizedCapturedRequest{
		path:          req.path,
		authorization: req.authorization,
		model:         req.model,
		instructions:  req.instructions,
		messages:      normalizeInputItems(inputItemsFromBody(req.body)),
	}
}

func newStreamingResponsesServer(t *testing.T, status int, contentType string, body string, captured chan<- capturedStreamRequest) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		payload := decodeBodyMap(t, r)
		if payload != nil {
			model, _ := payload["model"].(string)
			instructions, _ := payload["instructions"].(string)
			stream, _ := payload["stream"].(bool)
			select {
			case captured <- capturedStreamRequest{path: r.URL.Path, authorization: r.Header.Get("Authorization"), model: model, instructions: instructions, body: payload, stream: stream}:
			default:
			}
		}

		resolvedContentType := contentType
		if resolvedContentType == "" {
			resolvedContentType = "text/event-stream"
		}
		w.Header().Set("Content-Type", resolvedContentType)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/responses", handler)
	mux.HandleFunc("/v1/responses", handler)

	return httptest.NewServer(mux)
}

func readCapturedStreamRequest(t *testing.T, captured <-chan capturedStreamRequest) (capturedStreamRequest, bool) {
	t.Helper()
	select {
	case req := <-captured:
		return req, true
	default:
		return capturedStreamRequest{}, false
	}
}

func normalizeCapturedStreamRequest(req capturedStreamRequest) normalizedCapturedStreamRequest {
	return normalizedCapturedStreamRequest{
		path:          req.path,
		authorization: req.authorization,
		model:         req.model,
		instructions:  req.instructions,
		messages:      normalizeInputItems(inputItemsFromBody(req.body)),
		stream:        req.stream,
	}
}

func newFileUploadServer(t *testing.T, response string, captured chan<- capturedFileUploadRequest) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		reader, err := r.MultipartReader()
		require.NoError(t, err)

		capturedRequest := capturedFileUploadRequest{path: r.URL.Path, authorization: r.Header.Get("Authorization")}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)

			body, readErr := io.ReadAll(part)
			require.NoError(t, readErr)

			switch part.FormName() {
			case "purpose":
				capturedRequest.purpose = string(body)
			case "file":
				capturedRequest.filename = part.FileName()
				capturedRequest.mimeType = part.Header.Get("Content-Type")
				capturedRequest.body = body
			}
		}

		select {
		case captured <- capturedRequest:
		default:
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/files", handler)
	mux.HandleFunc("/v1/files", handler)

	return httptest.NewServer(mux)
}

func newResponsesServerWithFileContent(t *testing.T, captured chan<- capturedRequest) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		select {
		case captured <- capturedRequest{path: r.URL.Path, authorization: r.Header.Get("Authorization")}:
		default:
		}

		switch r.URL.Path {
		case "/files/file-123", "/v1/files/file-123":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"file-123","filename":"report.txt","bytes":6,"created_at":1,"object":"file","purpose":"user_data","status":"processed"}`))
		case "/files/file-123/content", "/v1/files/file-123/content":
			w.Header().Set("Content-Type", "application/binary")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("report"))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/files/file-123", handler)
	mux.HandleFunc("/v1/files/file-123", handler)
	mux.HandleFunc("/files/file-123/content", handler)
	mux.HandleFunc("/v1/files/file-123/content", handler)

	return httptest.NewServer(mux)
}

func newFileDeleteServer(t *testing.T, captured chan<- capturedRequest) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		select {
		case captured <- capturedRequest{path: r.URL.Path, authorization: r.Header.Get("Authorization")}:
		default:
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"file-123","deleted":true,"object":"file"}`))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/files/file-123", handler)
	mux.HandleFunc("/v1/files/file-123", handler)

	return httptest.NewServer(mux)
}

func readCapturedFileUploadRequest(t *testing.T, captured <-chan capturedFileUploadRequest) (capturedFileUploadRequest, bool) {
	t.Helper()
	select {
	case req := <-captured:
		return req, true
	default:
		return capturedFileUploadRequest{}, false
	}
}

func newImagesServer(t *testing.T, status int, response string, captured chan<- capturedImageRequest) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		capturedRequest := capturedImageRequest{
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
			contentType:   r.Header.Get("Content-Type"),
		}

		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			reader, err := r.MultipartReader()
			require.NoError(t, err)

			capturedRequest.formValues = make(map[string]string)
			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				require.NoError(t, err)

				body, readErr := io.ReadAll(part)
				require.NoError(t, readErr)

				if part.FileName() != "" {
					capturedRequest.files = append(capturedRequest.files, capturedImageFile{
						fieldName: part.FormName(),
						fileName:  part.FileName(),
						mimeType:  part.Header.Get("Content-Type"),
						body:      body,
					})
					continue
				}

				capturedRequest.formValues[part.FormName()] = string(body)
			}
		} else {
			payload := decodeBodyMap(t, r)
			body, err := json.Marshal(payload)
			require.NoError(t, err)
			capturedRequest.body = body
			capturedRequest.formValues = make(map[string]string)
			for key, value := range payload {
				switch val := value.(type) {
				case string:
					capturedRequest.formValues[key] = val
				}
			}
		}

		select {
		case captured <- capturedRequest:
		default:
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/images/generations", handler)
	mux.HandleFunc("/v1/images/generations", handler)
	mux.HandleFunc("/images/edits", handler)
	mux.HandleFunc("/v1/images/edits", handler)

	return httptest.NewServer(mux)
}

func readCapturedImageRequest(t *testing.T, captured <-chan capturedImageRequest) (capturedImageRequest, bool) {
	t.Helper()
	select {
	case req := <-captured:
		return req, true
	default:
		return capturedImageRequest{}, false
	}
}

func newAudioServer(t *testing.T, status int, responseBody, responseType string, captured chan<- capturedAudioRequest) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		capturedRequest := capturedAudioRequest{
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
			contentType:   r.Header.Get("Content-Type"),
			body:          decodeBodyMap(t, r),
			accept:        r.Header.Get("Accept"),
		}

		select {
		case captured <- capturedRequest:
		default:
		}

		if responseType != "" {
			w.Header().Set("Content-Type", responseType)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(responseBody))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/audio/speech", handler)
	mux.HandleFunc("/v1/audio/speech", handler)

	return httptest.NewServer(mux)
}

func readCapturedAudioRequest(t *testing.T, captured <-chan capturedAudioRequest) (capturedAudioRequest, bool) {
	t.Helper()
	select {
	case req := <-captured:
		return req, true
	default:
		return capturedAudioRequest{}, false
	}
}

func newTranscriptionServer(t *testing.T, status int, responseBody string, captured chan<- capturedTranscriptionRequest) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		reader, err := r.MultipartReader()
		require.NoError(t, err)

		capturedRequest := capturedTranscriptionRequest{
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
			contentType:   r.Header.Get("Content-Type"),
			formValues:    make(map[string]string),
		}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)

			body, readErr := io.ReadAll(part)
			require.NoError(t, readErr)

			if part.FileName() != "" {
				capturedRequest.fileName = part.FileName()
				capturedRequest.fileMimeType = part.Header.Get("Content-Type")
				capturedRequest.fileBody = body
				continue
			}

			capturedRequest.formValues[part.FormName()] = string(body)
		}

		select {
		case captured <- capturedRequest:
		default:
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(responseBody))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/audio/transcriptions", handler)
	mux.HandleFunc("/v1/audio/transcriptions", handler)

	return httptest.NewServer(mux)
}

func readCapturedTranscriptionRequest(t *testing.T, captured <-chan capturedTranscriptionRequest) (capturedTranscriptionRequest, bool) {
	t.Helper()
	select {
	case req := <-captured:
		return req, true
	default:
		return capturedTranscriptionRequest{}, false
	}
}

func decodeBodyMap(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	r.Body = io.NopCloser(bytes.NewReader(body))

	var payload map[string]any
	require.NoErrorf(t, json.Unmarshal(body, &payload), "failed to unmarshal request body: %s", string(body))

	return payload
}

func inputItemsFromBody(body map[string]any) []map[string]any {
	rawItems, ok := body["input"].([]any)
	if !ok {
		return nil
	}

	items := make([]map[string]any, 0, len(rawItems))
	for _, rawItem := range rawItems {
		item, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		items = append(items, item)
	}

	return items
}

func normalizeInputItems(items []map[string]any) []normalizedMessage {
	messages := make([]normalizedMessage, 0, len(items))
	for _, item := range items {
		if role, ok := item["role"].(string); ok {
			messages = append(messages, normalizedMessage{role: role, content: messageText(item["content"])})
			continue
		}
		if _, ok := item["output"]; ok {
			messages = append(messages, normalizedMessage{role: "tool", content: messageText(item["output"])})
			continue
		}
		if _, ok := item["call_id"]; ok {
			messages = append(messages, normalizedMessage{role: "assistant", content: ""})
			continue
		}

		typ, _ := item["type"].(string)
		switch typ {
		case "message":
			role, _ := item["role"].(string)
			messages = append(messages, normalizedMessage{role: role, content: messageText(item["content"])})
		case "function_call_output":
			messages = append(messages, normalizedMessage{role: "tool", content: messageText(item["output"])})
		case "function_call":
			messages = append(messages, normalizedMessage{role: "assistant", content: ""})
		}
	}

	return messages
}

func normalizeProviderStreamEvents(events []contractsai.StreamEvent) []normalizedStreamEvent {
	normalized := make([]normalizedStreamEvent, 0, len(events))
	for _, event := range events {
		entry := normalizedStreamEvent{eventType: event.Type, delta: event.Delta, err: event.Error, toolCalls: event.ToolCalls}
		if event.Usage != nil {
			entry.usage = &streamUsageSnapshot{input: event.Usage.Input(), output: event.Usage.Output(), total: event.Usage.Total()}
		}
		normalized = append(normalized, entry)
	}

	return normalized
}

func normalizeEmptyToolCalls(events []normalizedStreamEvent) []normalizedStreamEvent {
	for i := range events {
		if len(events[i].toolCalls) == 0 {
			events[i].toolCalls = nil
		}
	}

	return events
}

func responseBody(t *testing.T, text string, output []map[string]any, inputTokens, outputTokens, totalTokens int) string {
	t.Helper()

	if output == nil {
		output = []map[string]any{{
			"id":     "msg_1",
			"type":   "message",
			"role":   "assistant",
			"status": "completed",
			"content": []map[string]any{{
				"type":        "output_text",
				"text":        text,
				"annotations": []any{},
				"logprobs":    []any{},
			}},
		}}
	}

	body := map[string]any{
		"id":     "resp_1",
		"object": "response",
		"model":  "gpt-test",
		"output": output,
		"usage": map[string]any{
			"input_tokens":          inputTokens,
			"input_tokens_details":  map[string]any{"cached_tokens": 0},
			"output_tokens":         outputTokens,
			"output_tokens_details": map[string]any{"reasoning_tokens": 0},
			"total_tokens":          totalTokens,
		},
	}

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	return string(encoded)
}

func imageResponseBody(t *testing.T, format, content string, inputTokens, outputTokens, totalTokens int) string {
	t.Helper()

	body := map[string]any{
		"created":       123,
		"output_format": format,
		"data": []map[string]any{{
			"b64_json": base64.StdEncoding.EncodeToString([]byte(content)),
		}},
		"usage": map[string]any{
			"input_tokens":         inputTokens,
			"input_tokens_details": map[string]any{"image_tokens": 0, "text_tokens": inputTokens},
			"output_tokens":        outputTokens,
			"total_tokens":         totalTokens,
		},
	}

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	return string(encoded)
}

func transcriptionResponseBody(t *testing.T, text string, segments []map[string]any, usage map[string]any) string {
	t.Helper()

	body := map[string]any{
		"text": text,
	}
	if segments != nil {
		body["segments"] = segments
	}
	if usage != nil {
		body["usage"] = usage
	}

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	return string(encoded)
}

func bodySequence(body string) []string {
	return []string{body}
}

func stringsJoin(parts []string) string {
	return strings.Join(parts, "")
}

func stringsJoinLines(parts ...string) string {
	return strings.Join(parts, "\n")
}
