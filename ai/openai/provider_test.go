package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
				return &cfg
			}(),
		},
		{
			name: "keeps configured default model",
			setup: func() {
				mockConfig.EXPECT().UnmarshalKey("ai.providers.openai", new(contractsai.ProviderConfig)).RunAndReturn(func(_ string, rawVal any) error {
					cfg := rawVal.(*contractsai.ProviderConfig)
					cfg.Key = "test-key"
					cfg.Models.Text.Default = "gpt-custom"
					return nil
				}).Once()
			},
			expectConfig: func() *contractsai.ProviderConfig {
				cfg := contractsai.ProviderConfig{Key: "test-key"}
				cfg.Models.Text.Default = "gpt-custom"
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
			body:   responseBody("assistant reply", nil, 11, 7, 18),
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
			body:   responseBody("ok", nil, 1, 1, 2),
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
			body:   `{"message":"boom","type":"server_error"}`,
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
			server := newResponsesServer(t, tt.status, bodySequence(tt.body), captured)
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
	promptServer := newResponsesServer(t, http.StatusOK, bodySequence(responseBody("ok", nil, 1, 1, 2)), promptCaptured)
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

func (unsupportedAttachment) Kind() contractsai.AttachmentKind {
	return contractsai.AttachmentKind("audio")
}
func (unsupportedAttachment) FileName() string { return "audio.mp3" }
func (unsupportedAttachment) MimeType() string { return "audio/mpeg" }
func (unsupportedAttachment) Content(context.Context) ([]byte, error) {
	return []byte("audio"), nil
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
			stream.Then(func(resp contractsai.Response) {
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

func newResponsesServer(t *testing.T, status int, responses []string, captured chan<- capturedRequest) *httptest.Server {
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

		body := ""
		if callCount < len(responses) {
			body = responses[callCount]
		}
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

func decodeBodyMap(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	r.Body = io.NopCloser(bytes.NewReader(body))

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

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

func responseBody(text string, output []map[string]any, inputTokens, outputTokens, totalTokens int) string {
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

	encoded, _ := json.Marshal(body)
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
