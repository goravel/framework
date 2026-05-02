package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	goopenai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
	mocksai "github.com/goravel/framework/mocks/ai"
)

type capturedToolRequest struct {
	path          string
	authorization string
	model         string
	body          map[string]any
}

func newToolResponsesServer(t *testing.T, responses []string, captured chan<- capturedToolRequest) *httptest.Server {
	t.Helper()

	callCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		payload := decodeToolBodyMap(t, r)
		if payload != nil {
			model, _ := payload["model"].(string)
			select {
			case captured <- capturedToolRequest{path: r.URL.Path, authorization: r.Header.Get("Authorization"), model: model, body: payload}:
			default:
			}
		}

		if callCount >= len(responses) {
			t.Fatalf("unexpected extra response request: call %d exceeds configured responses %d", callCount+1, len(responses))
		}
		body := responses[callCount]
		callCount++

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/responses", handler)
	mux.HandleFunc("/v1/responses", handler)

	return httptest.NewServer(mux)
}

func TestProviderPromptWithTools(t *testing.T) {
	var mockAgent *mocksai.Agent

	beforeEach := func() {
		mockAgent = mocksai.NewAgent(t)
	}

	tests := []struct {
		name              string
		setup             func()
		tools             []contractsai.Tool
		input             string
		apiResponse       string
		expectText        string
		expectToolCallIDs []string
		expectToolNames   []string
		expectErr         bool
	}{
		{
			name:  "returns tool calls when model requests tool invocation",
			input: "What is the weather in London?",
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			tools: []contractsai.Tool{newStaticTool("get_weather", "Get the weather", nil)},
			apiResponse: responseBody(t, "", []map[string]any{{
				"id":        "fc_1",
				"type":      "function_call",
				"call_id":   "call_1",
				"name":      "get_weather",
				"arguments": `{"city":"London"}`,
				"status":    "completed",
			}}, 10, 5, 15),
			expectToolCallIDs: []string{"call_1"},
			expectToolNames:   []string{"get_weather"},
		},
		{
			name:  "returns plain text when no tool calls",
			input: "Hello",
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			tools:       []contractsai.Tool{newStaticTool("get_weather", "Get the weather", nil)},
			apiResponse: responseBody(t, "Hi there!", nil, 5, 3, 8),
			expectText:  "Hi there!",
		},
		{
			name:  "decodes tool call args into map",
			input: "Get weather",
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			tools: []contractsai.Tool{newStaticTool("get_weather", "Get the weather", nil)},
			apiResponse: responseBody(t, "", []map[string]any{{
				"id":        "fc_2",
				"type":      "function_call",
				"call_id":   "call_2",
				"name":      "get_weather",
				"arguments": `{"city":"Paris","units":"celsius"}`,
				"status":    "completed",
			}}, 8, 4, 12),
			expectToolCallIDs: []string{"call_2"},
			expectToolNames:   []string{"get_weather"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()

			captured := make(chan capturedToolRequest, 1)
			server := newToolResponsesServer(t, []string{tt.apiResponse}, captured)
			t.Cleanup(server.Close)

			provider := &Provider{
				client: goopenai.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("test-key")),
				config: contractsai.ProviderConfig{},
			}
			provider.config.Models.Text.Default = "gpt-default"

			prompt := contractsai.AgentPrompt{Agent: mockAgent, Input: tt.input, Tools: tt.tools}
			resp, err := provider.Prompt(context.Background(), prompt)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tt.expectText != "" {
				assert.Equal(t, tt.expectText, resp.Text())
				assert.Empty(t, resp.ToolCalls())
			}

			if len(tt.expectToolCallIDs) > 0 {
				require.Len(t, resp.ToolCalls(), len(tt.expectToolCallIDs))
				for i, tc := range resp.ToolCalls() {
					assert.Equal(t, tt.expectToolCallIDs[i], tc.ID)
					assert.Equal(t, tt.expectToolNames[i], tc.Name)
					assert.NotEmpty(t, tc.RawArgs)
				}
			}
		})
	}
}

func TestProviderBuildInput_ToolCallHistory(t *testing.T) {
	tests := []struct {
		name             string
		messages         []contractsai.Message
		input            string
		expectedMessages []normalizedMessage
	}{
		{
			name: "maps tool result message correctly",
			messages: []contractsai.Message{
				{Role: contractsai.RoleUser, Content: "What is the weather?"},
				{Role: contractsai.RoleAssistant, Content: "", ToolCalls: []contractsai.ToolCall{{ID: "call_1", Name: "get_weather", Args: map[string]any{"city": "London"}, RawArgs: `{"city":"London"}`}}},
				{Role: contractsai.RoleToolResult, Content: "Sunny, 25°C", ToolCallID: "call_1"},
			},
			input: "Thanks",
			expectedMessages: []normalizedMessage{
				{role: "user", content: "What is the weather?"},
				{role: "assistant", content: ""},
				{role: "tool", content: "Sunny, 25°C"},
				{role: "user", content: "Thanks"},
			},
		},
		{
			name: "assistant message with text and no tool calls",
			messages: []contractsai.Message{
				{Role: contractsai.RoleAssistant, Content: "Hello!"},
			},
			input: "How are you?",
			expectedMessages: []normalizedMessage{
				{role: "assistant", content: "Hello!"},
				{role: "user", content: "How are you?"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiResponse := responseBody(t, "ok", nil, 1, 1, 2)

			mockAgent := mocksai.NewAgent(t)
			mockAgent.EXPECT().Instructions().Return("").Once()
			mockAgent.EXPECT().Messages().Return(tt.messages).Once()

			captured := make(chan capturedToolRequest, 1)
			server := newToolResponsesServer(t, []string{apiResponse}, captured)
			t.Cleanup(server.Close)

			provider := &Provider{
				client: goopenai.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("test-key")),
				config: contractsai.ProviderConfig{},
			}
			provider.config.Models.Text.Default = "gpt-default"

			_, err := provider.Prompt(context.Background(), contractsai.AgentPrompt{Agent: mockAgent, Input: tt.input})
			require.NoError(t, err)

			req, ok := readCapturedToolRequest(t, captured)
			require.True(t, ok)
			assert.Equal(t, tt.expectedMessages, normalizeInputItems(inputItemsFromBody(req.body)))
		})
	}
}

func TestProviderBuildTools(t *testing.T) {
	tests := []struct {
		name          string
		tools         []contractsai.Tool
		expectInTools []string
	}{
		{
			name: "sends tool definitions to API",
			tools: []contractsai.Tool{
				newStaticTool("get_weather", "Get current weather", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"city": map[string]any{"type": "string"},
					},
					"required": []string{"city"},
				}),
				newStaticTool("send_email", "Send an email", nil),
			},
			expectInTools: []string{"get_weather", "send_email"},
		},
		{
			name:          "no tools sends no tools field",
			tools:         nil,
			expectInTools: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiResponse := responseBody(t, "ok", nil, 1, 1, 2)

			mockAgent := mocksai.NewAgent(t)
			mockAgent.EXPECT().Instructions().Return("").Once()
			mockAgent.EXPECT().Messages().Return(nil).Once()

			captured := make(chan capturedToolRequest, 1)
			server := newToolResponsesServer(t, []string{apiResponse}, captured)
			t.Cleanup(server.Close)

			provider := &Provider{
				client: goopenai.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("test-key")),
				config: contractsai.ProviderConfig{},
			}
			provider.config.Models.Text.Default = "gpt-default"

			_, err := provider.Prompt(context.Background(), contractsai.AgentPrompt{Agent: mockAgent, Input: "hello", Tools: tt.tools})
			require.NoError(t, err)

			req, ok := readCapturedToolRequest(t, captured)
			require.True(t, ok)

			tools := toolItemsFromBody(req.body)
			if tt.expectInTools == nil {
				assert.Empty(t, tools)
				return
			}

			require.Len(t, tools, len(tt.expectInTools))
			for i, tool := range tools {
				assert.Equal(t, "function", tool["type"])
				assert.Equal(t, tt.expectInTools[i], tool["name"])
				if tt.expectInTools[i] == "send_email" {
					_, hasStrict := tool["strict"]
					assert.False(t, hasStrict)
					_, hasParameters := tool["parameters"]
					assert.False(t, hasParameters)
				}
				if tt.expectInTools[i] == "get_weather" {
					assert.Equal(t, true, tool["strict"])
					assert.NotNil(t, tool["parameters"])
				}
			}
		})
	}
}

func TestConversationPromptUsesPreviousResponseIDForToolLoop(t *testing.T) {
	captured := make(chan capturedToolRequest, 2)
	server := newToolResponsesServer(t, []string{
		responseBody(t, "", []map[string]any{{
			"id":        "fc_1",
			"type":      "function_call",
			"call_id":   "call_1",
			"name":      "lookup_weather",
			"arguments": `{"city":"London"}`,
			"status":    "completed",
		}}, 6, 3, 9),
		responseBody(t, "It is sunny.", nil, 4, 2, 6),
	}, captured)
	t.Cleanup(server.Close)

	provider := &Provider{
		client: goopenai.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("test-key")),
		config: contractsai.ProviderConfig{},
	}
	provider.config.Models.Text.Default = "gpt-default"

	tool := newStaticTool("lookup_weather", "Lookup weather", nil)
	agent := &conversationAgentStub{tools: []contractsai.Tool{tool}}
	conversation := frameworkai.NewConversation(context.Background(), agent, provider, "gpt-default", nil)

	resp, err := conversation.Prompt("What is the weather in London?")
	require.NoError(t, err)
	assert.Equal(t, "It is sunny.", resp.Text())

	firstReq, ok := readCapturedToolRequest(t, captured)
	require.True(t, ok)
	secondReq, ok := readCapturedToolRequest(t, captured)
	require.True(t, ok)

	assert.Empty(t, firstReq.body["previous_response_id"])
	assert.Equal(t, "resp_1", secondReq.body["previous_response_id"])
	assert.Equal(t, []normalizedMessage{{role: "tool", content: "tool result"}}, normalizeInputItems(inputItemsFromBody(secondReq.body)))
}

func TestConversationStreamUsesPreviousResponseIDForToolLoop(t *testing.T) {
	responsesBodies := []string{
		stringsJoinLines(
			`data: {"type":"response.completed","sequence_number":1,"response":{"id":"resp_stream_1","object":"response","model":"gpt-test","output":[{"id":"fc_1","type":"function_call","call_id":"call_1","name":"lookup_weather","arguments":"{\"city\":\"London\"}","status":"completed"}],"usage":{"input_tokens":6,"input_tokens_details":{"cached_tokens":0},"output_tokens":3,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":9}}}`,
			``,
			`data: [DONE]`,
			``,
		),
		stringsJoinLines(
			`data: {"type":"response.output_text.delta","sequence_number":1,"item_id":"msg_1","output_index":0,"content_index":0,"delta":"It is sunny."}`,
			``,
			`data: {"type":"response.completed","sequence_number":2,"response":{"id":"resp_stream_2","object":"response","model":"gpt-test","output":[{"id":"msg_1","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"It is sunny.","annotations":[],"logprobs":[]}]}],"usage":{"input_tokens":4,"input_tokens_details":{"cached_tokens":0},"output_tokens":2,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":6}}}`,
			``,
			`data: [DONE]`,
			``,
		),
	}

	captured := make(chan capturedStreamRequest, 2)
	callCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		payload := decodeBodyMap(t, r)
		if payload != nil {
			model, _ := payload["model"].(string)
			instructions, _ := payload["instructions"].(string)
			stream, _ := payload["stream"].(bool)
			captured <- capturedStreamRequest{path: r.URL.Path, authorization: r.Header.Get("Authorization"), model: model, instructions: instructions, body: payload, stream: stream}
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responsesBodies[callCount]))
		callCount++
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/responses", handler)
	mux.HandleFunc("/v1/responses", handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	provider := &Provider{
		client: goopenai.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("test-key")),
		config: contractsai.ProviderConfig{},
	}
	provider.config.Models.Text.Default = "gpt-default"

	tool := newStaticTool("lookup_weather", "Lookup weather", nil)
	agent := &conversationAgentStub{tools: []contractsai.Tool{tool}}
	conversation := frameworkai.NewConversation(context.Background(), agent, provider, "gpt-default", nil)

	stream, err := conversation.Stream("What is the weather in London?")
	require.NoError(t, err)

	var events []contractsai.StreamEvent
	err = stream.Each(func(event contractsai.StreamEvent) error {
		events = append(events, event)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []normalizedStreamEvent{
		{eventType: contractsai.StreamEventTypeToolCall, toolCalls: []contractsai.ToolCall{{ID: "call_1", Name: "lookup_weather", Args: map[string]any{"city": "London"}, RawArgs: `{"city":"London"}`}}},
		{eventType: contractsai.StreamEventTypeTextDelta, delta: "It is sunny."},
		{eventType: contractsai.StreamEventTypeDone, usage: &streamUsageSnapshot{input: 4, output: 2, total: 6}},
	}, normalizeEmptyToolCalls(normalizeProviderStreamEvents(events)))

	firstReq, ok := readCapturedStreamRequest(t, captured)
	require.True(t, ok)
	secondReq, ok := readCapturedStreamRequest(t, captured)
	require.True(t, ok)

	assert.Empty(t, firstReq.body["previous_response_id"])
	assert.Equal(t, "resp_stream_1", secondReq.body["previous_response_id"])
	assert.Equal(t, []normalizedMessage{{role: "tool", content: "tool result"}}, normalizeInputItems(inputItemsFromBody(secondReq.body)))
}

type conversationAgentStub struct {
	tools []contractsai.Tool
}

func (a *conversationAgentStub) Instructions() string                 { return "" }
func (a *conversationAgentStub) Messages() []contractsai.Message      { return nil }
func (a *conversationAgentStub) Middleware() []contractsai.Middleware { return nil }
func (a *conversationAgentStub) Tools() []contractsai.Tool            { return a.tools }

func readCapturedToolRequest(t *testing.T, captured <-chan capturedToolRequest) (capturedToolRequest, bool) {
	t.Helper()
	select {
	case req := <-captured:
		return req, true
	default:
		return capturedToolRequest{}, false
	}
}

func decodeToolBodyMap(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	r.Body = io.NopCloser(bytes.NewReader(body))

	var payload map[string]any
	require.NoErrorf(t, json.Unmarshal(body, &payload), "failed to unmarshal request body: %s", string(body))

	return payload
}

func toolItemsFromBody(body map[string]any) []map[string]any {
	rawTools, ok := body["tools"].([]any)
	if !ok {
		return nil
	}

	tools := make([]map[string]any, 0, len(rawTools))
	for _, rawTool := range rawTools {
		tool, ok := rawTool.(map[string]any)
		if !ok {
			continue
		}
		tools = append(tools, tool)
	}

	return tools
}

func newStaticTool(name, description string, params map[string]any) contractsai.Tool {
	return &staticTool{name: name, description: description, params: params}
}

type staticTool struct {
	name        string
	description string
	params      map[string]any
}

func (t *staticTool) Name() string               { return t.name }
func (t *staticTool) Description() string        { return t.description }
func (t *staticTool) Parameters() map[string]any { return t.params }
func (t *staticTool) Execute(_ context.Context, _ map[string]any) (string, error) {
	return "tool result", nil
}
