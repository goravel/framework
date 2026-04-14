package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	goopenai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
	mocksai "github.com/goravel/framework/mocks/ai"
)

// capturedToolRequest captures the tools field sent to the completions endpoint.
type capturedToolRequest struct {
	path          string
	authorization string
	model         string
	messages      []map[string]any
	tools         []map[string]any
}

func newToolChatServer(t *testing.T, responses []string, captured chan<- capturedToolRequest) *httptest.Server {
	t.Helper()

	callCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer errors.Ignore(r.Body.Close)

		var payload struct {
			Model    string           `json:"model"`
			Messages []map[string]any `json:"messages"`
			Tools    []map[string]any `json:"tools"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
			select {
			case captured <- capturedToolRequest{
				path:          r.URL.Path,
				authorization: r.Header.Get("Authorization"),
				model:         payload.Model,
				messages:      payload.Messages,
				tools:         payload.Tools,
			}:
			default:
			}
		}

		body := ""
		if callCount < len(responses) {
			body = responses[callCount]
		}
		callCount++

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chat/completions", handler)
	mux.HandleFunc("/v1/chat/completions", handler)

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
			apiResponse: `{
				"id":"cmpl_tc1","object":"chat.completion","created":1,"model":"gpt-default",
				"choices":[{"index":0,"finish_reason":"tool_calls","message":{
					"role":"assistant","content":"",
					"tool_calls":[{"id":"call_1","type":"function","function":{"name":"get_weather","arguments":"{\"city\":\"London\"}"}}]
				}}],
				"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}
			}`,
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
			tools: []contractsai.Tool{newStaticTool("get_weather", "Get the weather", nil)},
			apiResponse: `{
				"id":"cmpl_tc2","object":"chat.completion","created":1,"model":"gpt-default",
				"choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"Hi there!","refusal":""}}],
				"usage":{"prompt_tokens":5,"completion_tokens":3,"total_tokens":8}
			}`,
			expectText: "Hi there!",
		},
		{
			name:  "decodes tool call args into map",
			input: "Get weather",
			setup: func() {
				mockAgent.EXPECT().Instructions().Return("").Once()
				mockAgent.EXPECT().Messages().Return(nil).Once()
			},
			tools: []contractsai.Tool{newStaticTool("get_weather", "Get the weather", nil)},
			apiResponse: `{
				"id":"cmpl_tc3","object":"chat.completion","created":1,"model":"gpt-default",
				"choices":[{"index":0,"finish_reason":"tool_calls","message":{
					"role":"assistant","content":"",
					"tool_calls":[{"id":"call_2","type":"function","function":{"name":"get_weather","arguments":"{\"city\":\"Paris\",\"units\":\"celsius\"}"}}]
				}}],
				"usage":{"prompt_tokens":8,"completion_tokens":4,"total_tokens":12}
			}`,
			expectToolCallIDs: []string{"call_2"},
			expectToolNames:   []string{"get_weather"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()

			captured := make(chan capturedToolRequest, 1)
			server := newToolChatServer(t, []string{tt.apiResponse}, captured)
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

func TestProviderBuildMessages_ToolCallHistory(t *testing.T) {
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
				{
					Role:      contractsai.RoleAssistant,
					Content:   "",
					ToolCalls: []contractsai.ToolCall{{ID: "call_1", Name: "get_weather", Args: map[string]any{"city": "London"}, RawArgs: `{"city":"London"}`}},
				},
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
			apiResponse := `{"id":"cmpl_bm","object":"chat.completion","created":1,"model":"gpt-default","choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"ok","refusal":""}}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

			mockAgent := mocksai.NewAgent(t)
			mockAgent.EXPECT().Instructions().Return("").Once()
			mockAgent.EXPECT().Messages().Return(tt.messages).Once()

			captured := make(chan capturedToolRequest, 1)
			server := newToolChatServer(t, []string{apiResponse}, captured)
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

			got := make([]normalizedMessage, 0, len(req.messages))
			for _, m := range req.messages {
				role, _ := m["role"].(string)
				got = append(got, normalizedMessage{role: role, content: messageText(m["content"])})
			}
			assert.Equal(t, tt.expectedMessages, got)
		})
	}
}

func TestProviderBuildTools(t *testing.T) {
	tests := []struct {
		name          string
		tools         []contractsai.Tool
		expectInTools []string // expected tool names in the request
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
			apiResponse := `{"id":"cmpl_bt","object":"chat.completion","created":1,"model":"gpt-default","choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"ok","refusal":""}}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

			mockAgent := mocksai.NewAgent(t)
			mockAgent.EXPECT().Instructions().Return("").Once()
			mockAgent.EXPECT().Messages().Return(nil).Once()

			captured := make(chan capturedToolRequest, 1)
			server := newToolChatServer(t, []string{apiResponse}, captured)
			t.Cleanup(server.Close)

			provider := &Provider{
				client: goopenai.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("test-key")),
				config: contractsai.ProviderConfig{},
			}
			provider.config.Models.Text.Default = "gpt-default"

			_, err := provider.Prompt(context.Background(), contractsai.AgentPrompt{
				Agent: mockAgent, Input: "hello", Tools: tt.tools,
			})
			require.NoError(t, err)

			req, ok := readCapturedToolRequest(t, captured)
			require.True(t, ok)

			if tt.expectInTools == nil {
				assert.Empty(t, req.tools)
				return
			}

			require.Len(t, req.tools, len(tt.expectInTools))
			for i, tool := range req.tools {
				fn, _ := tool["function"].(map[string]any)
				require.NotNil(t, fn, "tool[%d] missing function field", i)
				assert.Equal(t, tt.expectInTools[i], fn["name"])
			}
		})
	}
}

func readCapturedToolRequest(t *testing.T, captured <-chan capturedToolRequest) (capturedToolRequest, bool) {
	t.Helper()
	select {
	case req := <-captured:
		return req, true
	default:
		return capturedToolRequest{}, false
	}
}

// newStaticTool returns a simple Tool implementation for tests.
func newStaticTool(name, description string, params map[string]any) contractsai.Tool {
	return &staticTool{name: name, description: description, params: params}
}

type staticTool struct {
	name        string
	description string
	params      map[string]any
}

func (t *staticTool) Name() string                   { return t.name }
func (t *staticTool) Description() string            { return t.description }
func (t *staticTool) Parameters() map[string]any     { return t.params }
func (t *staticTool) Execute(_ context.Context, _ map[string]any) (string, error) {
	return "tool result", nil
}
