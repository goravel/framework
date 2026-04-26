package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	goopenai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
	mocksai "github.com/goravel/framework/mocks/ai"
	mocksconfig "github.com/goravel/framework/mocks/config"
)

type capturedRequest struct {
	path          string
	authorization string
	model         string
	messages      []map[string]any
}

type normalizedCapturedRequest struct {
	path          string
	authorization string
	model         string
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
	messages      []map[string]any
	stream        bool
	includeUsage  bool
}

type normalizedCapturedStreamRequest struct {
	path          string
	authorization string
	model         string
	messages      []normalizedMessage
	stream        bool
	includeUsage  bool
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
			name:   "builds messages with default model",
			status: http.StatusOK,
			body:   `{"id":"cmpl_1","object":"chat.completion","created":1,"model":"gpt-test","choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"assistant reply","refusal":""}}],"usage":{"prompt_tokens":11,"completion_tokens":7,"total_tokens":18}}`,
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
				path:          "/chat/completions",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				messages: []normalizedMessage{
					{role: "system", content: "system rule"},
					{role: "user", content: "history user"},
					{role: "assistant", content: "history assistant"},
					{role: "user", content: "new input"},
				},
			},
		},
		{
			name:   "uses prompt model override",
			status: http.StatusOK,
			body:   `{"id":"cmpl_2","object":"chat.completion","created":1,"model":"gpt-test","choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"ok","refusal":""}}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`,
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			modelOverride: "gpt-override",
			input:         "hello",
			expectText:    "ok",
			expectUsage:   usageCheck{input: 1, output: 1, total: 2},
			expectRequest: normalizedCapturedRequest{
				path:          "/chat/completions",
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
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			input:            "hello",
			expectErr:        true,
			expectErrMessage: "boom",
			expectRequest: normalizedCapturedRequest{
				path:          "/chat/completions",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				messages: []normalizedMessage{
					{role: "user", content: "hello"},
				},
			},
		},
		{
			name:   "handles empty choices",
			status: http.StatusOK,
			body:   `{"id":"cmpl_3","object":"chat.completion","created":1,"model":"gpt-test","choices":[],"usage":{"prompt_tokens":3,"completion_tokens":0,"total_tokens":3}}`,
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			input:       "hello",
			expectText:  "",
			expectUsage: usageCheck{input: 3, output: 0, total: 3},
			expectRequest: normalizedCapturedRequest{
				path:          "/chat/completions",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				messages: []normalizedMessage{
					{role: "user", content: "hello"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()

			captured := make(chan capturedRequest, 1)
			server := newChatServer(t, tt.status, tt.body, captured)
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
				assert.Equal(t, tt.expectErrMessage, apiErr.Message)
				assert.ErrorContains(t, err, tt.expectErrMessage)
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

	successStreamBody := strings.Join([]string{
		`data: {"id":"chatcmpl_stream_1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"content":"hel"},"finish_reason":null}],"usage":null}`,
		``,
		`data: {"id":"chatcmpl_stream_1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"content":""},"finish_reason":null}],"usage":null}`,
		``,
		`data: {"id":"chatcmpl_stream_1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[],"usage":null}`,
		``,
		`data: {"id":"chatcmpl_stream_1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[],"usage":{"prompt_tokens":4,"completion_tokens":2,"total_tokens":6}}`,
		``,
		`data: {"id":"chatcmpl_stream_1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"content":"lo"},"finish_reason":null}],"usage":null}`,
		``,
		`data: [DONE]`,
		``,
	}, "\n")

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
				{
					eventType: contractsai.StreamEventTypeDone,
					usage:     &streamUsageSnapshot{input: 4, output: 2, total: 6},
				},
			},
			expectRequest: normalizedCapturedStreamRequest{
				path:          "/chat/completions",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				stream:        true,
				includeUsage:  true,
				messages: []normalizedMessage{
					{role: "system", content: "system rule"},
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
			body: strings.Join([]string{
				`data: {"id":"chatcmpl_stream_2","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"content":"ok"},"finish_reason":null}],"usage":null}`,
				``,
				`data: {"id":"chatcmpl_stream_2","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`,
				``,
				`data: [DONE]`,
				``,
			}, "\n"),
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
				{
					eventType: contractsai.StreamEventTypeDone,
					usage:     &streamUsageSnapshot{input: 1, output: 1, total: 2},
				},
			},
			expectRequest: normalizedCapturedStreamRequest{
				path:          "/chat/completions",
				authorization: "Bearer test-key",
				model:         "gpt-override",
				stream:        true,
				includeUsage:  true,
				messages: []normalizedMessage{
					{role: "user", content: "hello"},
				},
			},
		},
		{
			name:        "accumulates streaming tool call deltas across chunks",
			status:      http.StatusOK,
			contentType: "text/event-stream",
			body: strings.Join([]string{
				`data: {"id":"chatcmpl_stream_tc","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"get_weather","arguments":"{\"city\":"}}]},"finish_reason":null}],"usage":null}`,
				``,
				`data: {"id":"chatcmpl_stream_tc","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"type":"function","function":{"arguments":"\"London\",\"units\":\"celsius\"}"}}]},"finish_reason":null}],"usage":null}`,
				``,
				`data: {"id":"chatcmpl_stream_tc","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[],"usage":{"prompt_tokens":6,"completion_tokens":3,"total_tokens":9}}`,
				``,
				`data: [DONE]`,
				``,
			}, "\n"),
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
			expectEvents: []normalizedStreamEvent{{
				eventType: contractsai.StreamEventTypeDone,
				usage:     &streamUsageSnapshot{input: 6, output: 3, total: 9},
			}},
			expectRequest: normalizedCapturedStreamRequest{
				path:          "/chat/completions",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				stream:        true,
				includeUsage:  true,
				messages: []normalizedMessage{
					{role: "user", content: "hello"},
				},
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
				path:          "/chat/completions",
				authorization: "Bearer test-key",
				model:         "gpt-default",
				stream:        true,
				includeUsage:  true,
				messages: []normalizedMessage{
					{role: "user", content: "hello"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()

			captured := make(chan capturedStreamRequest, 1)
			server := newStreamingChatServer(t, tt.status, tt.contentType, tt.body, captured)
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
					thenUsage = usageCheck{
						input:  resp.Usage().Input(),
						output: resp.Usage().Output(),
						total:  resp.Usage().Total(),
					}
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
				assert.Equal(t, tt.expectErrMessage, apiErr.Message)
				assert.ErrorContains(t, eachErr, tt.expectErrMessage)
				require.Len(t, normalizedEvents, 1)
				assert.Equal(t, contractsai.StreamEventTypeError, normalizedEvents[0].eventType)
				assert.Contains(t, normalizedEvents[0].err, tt.expectErrMessage)
				assert.Equal(t, 0, thenCalled)
			} else {
				require.NoError(t, eachErr)
				assert.Equal(t, tt.expectEvents, normalizedEvents)
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

func newChatServer(t *testing.T, status int, body string, captured chan<- capturedRequest) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		var payload struct {
			Model    string           `json:"model"`
			Messages []map[string]any `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
			select {
			case captured <- capturedRequest{
				path:          r.URL.Path,
				authorization: r.Header.Get("Authorization"),
				model:         payload.Model,
				messages:      payload.Messages,
			}:
			default:
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chat/completions", handler)
	mux.HandleFunc("/v1/chat/completions", handler)

	return httptest.NewServer(mux)
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
		return strings.Join(parts, "")
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
	messages := make([]normalizedMessage, 0, len(req.messages))
	for _, message := range req.messages {
		role, _ := message["role"].(string)
		messages = append(messages, normalizedMessage{
			role:    role,
			content: messageText(message["content"]),
		})
	}

	return normalizedCapturedRequest{
		path:          req.path,
		authorization: req.authorization,
		model:         req.model,
		messages:      messages,
	}
}

func newStreamingChatServer(t *testing.T, status int, contentType string, body string, captured chan<- capturedStreamRequest) *httptest.Server {
	t.Helper()

	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		var payload struct {
			Model         string           `json:"model"`
			Messages      []map[string]any `json:"messages"`
			Stream        bool             `json:"stream"`
			StreamOptions struct {
				IncludeUsage bool `json:"include_usage"`
			} `json:"stream_options"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
			select {
			case captured <- capturedStreamRequest{
				path:          r.URL.Path,
				authorization: r.Header.Get("Authorization"),
				model:         payload.Model,
				messages:      payload.Messages,
				stream:        payload.Stream,
				includeUsage:  payload.StreamOptions.IncludeUsage,
			}:
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
	mux.HandleFunc("/chat/completions", handler)
	mux.HandleFunc("/v1/chat/completions", handler)

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
	messages := make([]normalizedMessage, 0, len(req.messages))
	for _, message := range req.messages {
		role, _ := message["role"].(string)
		messages = append(messages, normalizedMessage{
			role:    role,
			content: messageText(message["content"]),
		})
	}

	return normalizedCapturedStreamRequest{
		path:          req.path,
		authorization: req.authorization,
		model:         req.model,
		messages:      messages,
		stream:        req.stream,
		includeUsage:  req.includeUsage,
	}
}

func normalizeProviderStreamEvents(events []contractsai.StreamEvent) []normalizedStreamEvent {
	normalized := make([]normalizedStreamEvent, 0, len(events))
	for _, event := range events {
		entry := normalizedStreamEvent{
			eventType: event.Type,
			delta:     event.Delta,
			err:       event.Error,
			toolCalls: event.ToolCalls,
		}
		if event.Usage != nil {
			entry.usage = &streamUsageSnapshot{
				input:  event.Usage.Input(),
				output: event.Usage.Output(),
				total:  event.Usage.Total(),
			}
		}

		normalized = append(normalized, entry)
	}

	return normalized
}
