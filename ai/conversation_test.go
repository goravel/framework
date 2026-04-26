package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
	mocksai "github.com/goravel/framework/mocks/ai"
)

type ConversationTestSuite struct {
	suite.Suite
}

func TestConversationTestSuite(t *testing.T) {
	suite.Run(t, &ConversationTestSuite{})
}

func (s *ConversationTestSuite) TestPrompt() {
	ctx := context.Background()

	var (
		mockProvider *mocksai.Provider
		mockAgent    *mocksai.Agent
		conv         *conversation
	)

	beforeEach := func(initial []contractsai.Message, model string) {
		mockAgent = mocksai.NewAgent(s.T())
		mockAgent.EXPECT().Messages().Return(initial).Once()
		mockProvider = mocksai.NewProvider(s.T())
		conv = NewConversation(ctx, mockAgent, mockProvider, model, nil)
	}

	tests := []struct {
		name           string
		initial        []contractsai.Message
		model          string
		input          string
		setup          func() contractsai.Response
		expectMessages []contractsai.Message
		expectError    error
	}{
		{
			name:    "appends messages on success",
			initial: []contractsai.Message{{Role: contractsai.RoleUser, Content: "system"}},
			model:   "model-x",
			input:   "hello",
			setup: func() contractsai.Response {
				mockAgent.EXPECT().Tools().Return(nil).Once()
				mockResponse := mocksai.NewResponse(s.T())
				mockResponse.EXPECT().ToolCalls().Return(nil).Once()
				mockResponse.EXPECT().Text().Return("got it").Once()
				mockProvider.EXPECT().Prompt(ctx, contractsai.AgentPrompt{Agent: conv, Input: "hello", Model: "model-x"}).Return(mockResponse, nil).Once()
				return mockResponse
			},
			expectMessages: []contractsai.Message{
				{Role: contractsai.RoleUser, Content: "system"},
				{Role: contractsai.RoleUser, Content: "hello"},
				{Role: contractsai.RoleAssistant, Content: "got it"},
			},
		},
		{
			name:    "does not append on error",
			initial: []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "init"}},
			model:   "model-y",
			input:   "fail",
			setup: func() contractsai.Response {
				mockAgent.EXPECT().Tools().Return(nil).Once()
				mockProvider.EXPECT().Prompt(ctx, contractsai.AgentPrompt{Agent: conv, Input: "fail", Model: "model-y"}).Return(nil, assert.AnError).Once()
				return nil
			},
			expectMessages: []contractsai.Message{
				{Role: contractsai.RoleAssistant, Content: "init"},
			},
			expectError: assert.AnError,
		},
		{
			name:    "appends empty input and empty response",
			initial: []contractsai.Message{},
			model:   "model-empty",
			input:   "",
			setup: func() contractsai.Response {
				mockAgent.EXPECT().Tools().Return(nil).Once()
				mockResponse := mocksai.NewResponse(s.T())
				mockResponse.EXPECT().ToolCalls().Return(nil).Once()
				mockResponse.EXPECT().Text().Return("").Once()
				mockProvider.EXPECT().Prompt(ctx, contractsai.AgentPrompt{Agent: conv, Input: "", Model: "model-empty"}).Return(mockResponse, nil).Once()
				return mockResponse
			},
			expectMessages: []contractsai.Message{
				{Role: contractsai.RoleUser, Content: ""},
				{Role: contractsai.RoleAssistant, Content: ""},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			beforeEach(tt.initial, tt.model)
			expectResp := tt.setup()

			resp, err := conv.Prompt(tt.input)
			s.Equal(tt.expectError, err)
			s.Equal(expectResp, resp)
			s.Equal(tt.expectMessages, conv.Messages())
		})
	}
}

func (s *ConversationTestSuite) TestReset() {
	ctx := context.Background()

	var (
		mockProvider *mocksai.Provider
		mockAgent    *mocksai.Agent
		conv         *conversation
	)

	beforeEach := func(initial []contractsai.Message) {
		mockAgent = mocksai.NewAgent(s.T())
		mockAgent.EXPECT().Messages().Return(initial).Times(2)
		mockProvider = mocksai.NewProvider(s.T())
		conv = NewConversation(ctx, mockAgent, mockProvider, "model-z", nil)
	}

	tests := []struct {
		name         string
		initial      []contractsai.Message
		input        string
		promptBefore bool
		setup        func()
		expectBefore []contractsai.Message
		expectAfter  []contractsai.Message
	}{
		{
			name:         "restores initial messages after prompt",
			initial:      []contractsai.Message{{Role: contractsai.RoleToolResult, Content: "keep"}},
			input:        "append",
			promptBefore: true,
			setup: func() {
				mockAgent.EXPECT().Tools().Return(nil).Once()
				response := mocksai.NewResponse(s.T())
				mockProvider.EXPECT().Prompt(ctx, contractsai.AgentPrompt{Agent: conv, Input: "append", Model: "model-z"}).Return(response, nil).Once()
				response.EXPECT().ToolCalls().Return(nil).Once()
				response.EXPECT().Text().Return("done").Once()
			},
			expectBefore: []contractsai.Message{
				{Role: contractsai.RoleToolResult, Content: "keep"},
				{Role: contractsai.RoleUser, Content: "append"},
				{Role: contractsai.RoleAssistant, Content: "done"},
			},
			expectAfter: []contractsai.Message{{Role: contractsai.RoleToolResult, Content: "keep"}},
		},
		{
			name:         "keeps same messages when reset without prompt",
			initial:      []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}},
			promptBefore: false,
			setup:        func() {},
			expectBefore: []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}},
			expectAfter:  []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			beforeEach(tt.initial)
			tt.setup()

			if tt.promptBefore {
				_, err := conv.Prompt(tt.input)
				s.Require().NoError(err)
			}

			s.Equal(tt.expectBefore, conv.Messages())
			conv.Reset()
			resetMessages := conv.Messages()
			s.Equal(tt.expectAfter, resetMessages)
			s.NotSame(&tt.initial[0], &resetMessages[0])
		})
	}
}

type conversationProviderStub struct {
	streamFn func(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error)
}

func (r *conversationProviderStub) Prompt(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
	return nil, nil
}

func (r *conversationProviderStub) Stream(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
	if r.streamFn == nil {
		return nil, nil
	}

	return r.streamFn(ctx, prompt)
}

// agentStub is a minimal Agent implementation for stream tests that avoids mock expectation tracking.
type agentStub struct {
	messages   []contractsai.Message
	middleware []contractsai.Middleware
	tools      []contractsai.Tool
}

func (a *agentStub) Instructions() string                 { return "" }
func (a *agentStub) Middleware() []contractsai.Middleware { return a.middleware }
func (a *agentStub) Messages() []contractsai.Message      { return a.messages }
func (a *agentStub) Tools() []contractsai.Tool            { return a.tools }

func (s *ConversationTestSuite) TestStream() {
	ctx := context.Background()
	model := "stream-model"

	s.Run("returns provider error immediately without mutating messages", func() {
		initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}
		agent := &agentStub{messages: initial}

		var conv *conversation
		provider := &conversationProviderStub{
			streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				s.Equal(contractsai.AgentPrompt{Agent: conv, Input: "hello", Model: model, Tools: nil}, prompt)
				return nil, assert.AnError
			},
		}
		conv = NewConversation(ctx, agent, provider, model, nil)

		stream, err := conv.Stream("hello")

		s.Equal(assert.AnError, err)
		s.Nil(stream)
		s.Equal(initial, conv.Messages())
	})

	s.Run("appends messages after successful stream completion", func() {
		initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}
		agent := &agentStub{messages: initial}

		var conv *conversation
		provider := &conversationProviderStub{
			streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				s.Equal(contractsai.AgentPrompt{Agent: conv, Input: "hi", Model: model, Tools: nil}, prompt)
				return NewStreamableResponse(gotCtx, func(_ context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
					if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "partial"}); err != nil {
						return nil, err
					}
					if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeDone}); err != nil {
						return nil, err
					}

					return &streamableTestResponse{text: "assistant reply"}, nil
				}), nil
			},
		}
		conv = NewConversation(ctx, agent, provider, model, nil)

		stream, err := conv.Stream("hi")
		s.Require().NoError(err)
		s.Require().NotNil(stream)
		s.Equal(initial, conv.Messages())

		var events []contractsai.StreamEvent
		err = stream.Each(func(event contractsai.StreamEvent) error {
			events = append(events, event)
			return nil
		})

		s.NoError(err)
		s.Equal(normalizeStreamEvents([]contractsai.StreamEvent{
			{Type: contractsai.StreamEventTypeTextDelta, Delta: "partial"},
			{Type: contractsai.StreamEventTypeDone},
		}), normalizeStreamEvents(events))
		s.Equal([]contractsai.Message{
			{Role: contractsai.RoleAssistant, Content: "seed"},
			{Role: contractsai.RoleUser, Content: "hi"},
			{Role: contractsai.RoleAssistant, Content: "assistant reply"},
		}, conv.Messages())
	})

	s.Run("does not append messages when stream completes with error", func() {
		initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}
		agent := &agentStub{messages: initial}

		var conv *conversation
		provider := &conversationProviderStub{
			streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				s.Equal(contractsai.AgentPrompt{Agent: conv, Input: "hi", Model: model, Tools: nil}, prompt)
				return NewStreamableResponse(gotCtx, func(_ context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
					if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "partial"}); err != nil {
						return nil, err
					}

					return nil, assert.AnError
				}), nil
			},
		}
		conv = NewConversation(ctx, agent, provider, model, nil)

		stream, err := conv.Stream("hi")
		s.Require().NoError(err)
		s.Require().NotNil(stream)

		err = stream.Each(nil)
		s.Equal(assert.AnError, err)
		s.Equal(initial, conv.Messages())
	})
}

func (s *ConversationTestSuite) TestMessagesClone() {
	ctx := context.Background()
	initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}

	agent := &agentStub{messages: initial}
	conv := NewConversation(ctx, agent, &conversationProviderStub{}, "model", nil)

	initial[0].Content = "mutated"

	got := conv.Messages()
	s.Equal([]contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}, got)

	got[0].Content = "changed"
	s.Equal([]contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}, conv.Messages())
}

func (s *ConversationTestSuite) TestPromptToolInvocationLoop() {
	ctx := context.Background()

	// A provider stub that returns responses from a queue.
	type promptCall struct {
		response contractsai.Response
		err      error
	}

	type promptRecord struct {
		prompt contractsai.AgentPrompt
	}

	var (
		calls   []promptCall
		callIdx int
		records []promptRecord
	)

	provider := &conversationToolProviderStub{
		promptFn: func(_ context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
			records = append(records, promptRecord{prompt: prompt})
			resp := calls[callIdx]
			callIdx++
			return resp.response, resp.err
		},
	}

	s.Run("executes tool and re-prompts until plain text response", func() {
		calls = []promptCall{
			{response: &stubResponse{toolCalls: []contractsai.ToolCall{{ID: "c1", Name: "get_weather", Args: map[string]any{"city": "London"}, RawArgs: `{"city":"London"}`}}}},
			{response: &stubResponse{text: "The weather is sunny."}},
		}
		callIdx = 0
		records = nil

		tool := &stubTool{name: "get_weather", result: "Sunny, 25°C"}
		agent := &agentStub{tools: []contractsai.Tool{tool}}
		conv := NewConversation(ctx, agent, provider, "m", nil)

		resp, err := conv.Prompt("What's the weather in London?")
		s.Require().NoError(err)
		s.Equal("The weather is sunny.", resp.Text())

		// History: user + assistant(tool_calls) + tool_result + assistant(final)
		msgs := conv.Messages()
		s.Require().Len(msgs, 4)
		s.Equal(contractsai.RoleUser, msgs[0].Role)
		s.Equal("What's the weather in London?", msgs[0].Content)
		s.Equal(contractsai.RoleAssistant, msgs[1].Role)
		s.Len(msgs[1].ToolCalls, 1)
		s.Equal("c1", msgs[1].ToolCalls[0].ID)
		s.Equal(contractsai.RoleToolResult, msgs[2].Role)
		s.Equal("Sunny, 25°C", msgs[2].Content)
		s.Equal("c1", msgs[2].ToolCallID)
		s.Equal(contractsai.RoleAssistant, msgs[3].Role)
		s.Equal("The weather is sunny.", msgs[3].Content)

		// Second prompt had empty input (continuation after tool result).
		s.Equal("", records[1].prompt.Input)
		s.Equal(1, tool.callCount)
	})

	s.Run("returns error when tool is not found", func() {
		calls = []promptCall{
			{response: &stubResponse{toolCalls: []contractsai.ToolCall{{ID: "c2", Name: "unknown_tool", Args: nil}}}},
		}
		callIdx = 0
		records = nil

		agent := &agentStub{tools: []contractsai.Tool{}}
		conv := NewConversation(ctx, agent, provider, "m", nil)

		initialLen := len(conv.Messages())
		_, err := conv.Prompt("do something")
		s.ErrorContains(err, `tool "unknown_tool" not found`)
		// Message history must not grow on error.
		s.Len(conv.Messages(), initialLen)
	})

	s.Run("returns error when tool execution fails", func() {
		calls = []promptCall{
			{response: &stubResponse{toolCalls: []contractsai.ToolCall{{ID: "c3", Name: "fail_tool", Args: nil}}}},
		}
		callIdx = 0
		records = nil

		tool := &stubTool{name: "fail_tool", execErr: assert.AnError}
		agent := &agentStub{tools: []contractsai.Tool{tool}}
		conv := NewConversation(ctx, agent, provider, "m", nil)

		initialLen := len(conv.Messages())
		_, err := conv.Prompt("do something")
		s.ErrorContains(err, `tool "fail_tool" execution failed`)
		s.Len(conv.Messages(), initialLen)
	})

	s.Run("returns error when max iterations exceeded", func() {
		// Build MaxToolCallIterations responses that all return tool calls.
		calls = make([]promptCall, MaxToolCallIterations)
		for i := range calls {
			calls[i] = promptCall{
				response: &stubResponse{toolCalls: []contractsai.ToolCall{{ID: "cx", Name: "loop_tool", Args: nil}}},
			}
		}
		callIdx = 0
		records = nil

		tool := &stubTool{name: "loop_tool", result: "result"}
		agent := &agentStub{tools: []contractsai.Tool{tool}}
		conv := NewConversation(ctx, agent, provider, "m", nil)

		initialLen := len(conv.Messages())
		_, err := conv.Prompt("loop forever")
		s.ErrorContains(err, "exceeded")
		s.Len(conv.Messages(), initialLen)
	})
}

func (s *ConversationTestSuite) TestStreamToolInvocationLoop() {
	ctx := context.Background()

	type streamCall struct {
		response contractsai.Response
		events   []contractsai.StreamEvent
		err      error
	}

	makeProvider := func(calls []streamCall) (*conversationToolProviderStub, *int) {
		idx := 0
		p := &conversationToolProviderStub{
			promptFn: func(_ context.Context, _ contractsai.AgentPrompt) (contractsai.Response, error) {
				return nil, nil
			},
			streamFn: func(streamCtx context.Context, _ contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				c := calls[idx]
				idx++
				if c.err != nil {
					return nil, c.err
				}
				return NewStreamableResponse(streamCtx, func(_ context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
					for _, ev := range c.events {
						if err := emit(ev); err != nil {
							return nil, err
						}
					}
					return c.response, nil
				}), nil
			},
		}
		return p, &idx
	}

	s.Run("executes tool and re-prompts until plain text response", func() {
		toolCall := contractsai.ToolCall{ID: "c1", Name: "get_weather", Args: map[string]any{"city": "London"}, RawArgs: `{"city":"London"}`}
		provider, _ := makeProvider([]streamCall{
			{
				response: &stubResponse{toolCalls: []contractsai.ToolCall{toolCall}},
				events:   []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeDone}},
			},
			{
				response: &stubResponse{text: "The weather is sunny."},
				events: []contractsai.StreamEvent{
					{Type: contractsai.StreamEventTypeTextDelta, Delta: "The weather is sunny."},
					{Type: contractsai.StreamEventTypeDone},
				},
			},
		})

		tool := &stubTool{name: "get_weather", result: "Sunny, 25°C"}
		agent := &agentStub{tools: []contractsai.Tool{tool}}
		conv := NewConversation(ctx, agent, provider, "m", nil)

		stream, err := conv.Stream("What's the weather in London?")
		s.Require().NoError(err)

		var gotTypes []contractsai.StreamEventType
		var gotToolCalls []contractsai.ToolCall
		eachErr := stream.Each(func(event contractsai.StreamEvent) error {
			gotTypes = append(gotTypes, event.Type)
			if event.Type == contractsai.StreamEventTypeToolCall {
				gotToolCalls = append(gotToolCalls, event.ToolCalls...)
			}
			return nil
		})
		s.NoError(eachErr)

		s.Contains(gotTypes, contractsai.StreamEventTypeToolCall)
		s.Contains(gotTypes, contractsai.StreamEventTypeTextDelta)
		s.Contains(gotTypes, contractsai.StreamEventTypeDone)
		s.Equal([]contractsai.ToolCall{toolCall}, gotToolCalls)
		s.Equal(1, tool.callCount)

		// History: user + assistant(tool_calls) + tool_result + assistant(final)
		msgs := conv.Messages()
		s.Require().Len(msgs, 4)
		s.Equal(contractsai.RoleUser, msgs[0].Role)
		s.Equal("What's the weather in London?", msgs[0].Content)
		s.Equal(contractsai.RoleAssistant, msgs[1].Role)
		s.Equal(toolCall, msgs[1].ToolCalls[0])
		s.Equal(contractsai.RoleToolResult, msgs[2].Role)
		s.Equal("Sunny, 25°C", msgs[2].Content)
		s.Equal("c1", msgs[2].ToolCallID)
		s.Equal(contractsai.RoleAssistant, msgs[3].Role)
		s.Equal("The weather is sunny.", msgs[3].Content)
	})

	s.Run("returns provider stream error immediately without mutating messages", func() {
		provider, _ := makeProvider([]streamCall{
			{err: assert.AnError},
		})

		agent := &agentStub{tools: nil}
		conv := NewConversation(ctx, agent, provider, "m", nil)
		initial := conv.Messages()

		stream, err := conv.Stream("hello")
		s.Equal(assert.AnError, err)
		s.Nil(stream)
		s.Equal(initial, conv.Messages())
	})

	s.Run("returns error when stream finishes without a response", func() {
		provider, _ := makeProvider([]streamCall{{
			events: []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeDone}},
		}})

		conv := NewConversation(ctx, &agentStub{}, provider, "m", nil)

		stream, err := conv.Stream("hello")
		s.NoError(err)

		eachErr := stream.Each(nil)
		s.Equal(errors.AIResponseIsNil, eachErr)
		s.Empty(conv.Messages())
	})

	s.Run("returns error when tool is not found", func() {
		toolCall := contractsai.ToolCall{ID: "cx", Name: "missing_tool", Args: map[string]any{}}
		provider, _ := makeProvider([]streamCall{
			{
				response: &stubResponse{toolCalls: []contractsai.ToolCall{toolCall}},
				events:   []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeDone}},
			},
		})

		tool := &stubTool{name: "known_tool"}
		agent := &agentStub{tools: []contractsai.Tool{tool}}
		conv := NewConversation(ctx, agent, provider, "m", nil)

		stream, err := conv.Stream("hello")
		s.NoError(err)

		eachErr := stream.Each(nil)
		s.ErrorContains(eachErr, `tool "missing_tool" not found`)
		s.Len(conv.Messages(), 0)
	})

	s.Run("returns error when max iterations exceeded", func() {
		toolCall := contractsai.ToolCall{ID: "cx", Name: "loop_tool", Args: map[string]any{}}
		calls := make([]streamCall, MaxToolCallIterations)
		for i := range calls {
			calls[i] = streamCall{
				response: &stubResponse{toolCalls: []contractsai.ToolCall{toolCall}},
				events:   []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeDone}},
			}
		}
		provider, _ := makeProvider(calls)

		tool := &stubTool{name: "loop_tool", result: "result"}
		agent := &agentStub{tools: []contractsai.Tool{tool}}
		conv := NewConversation(ctx, agent, provider, "m", nil)

		stream, err := conv.Stream("loop forever")
		s.NoError(err)

		eachErr := stream.Each(nil)
		s.ErrorContains(eachErr, "exceeded")
		s.Len(conv.Messages(), 0)
	})
}

// conversationToolProviderStub is a Provider stub that delegates Prompt to a func.
type conversationToolProviderStub struct {
	promptFn func(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error)
	streamFn func(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error)
}

func (r *conversationToolProviderStub) Prompt(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
	return r.promptFn(ctx, prompt)
}

func (r *conversationToolProviderStub) Stream(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
	if r.streamFn == nil {
		return nil, nil
	}
	return r.streamFn(ctx, prompt)
}

func (s *ConversationTestSuite) TestExecuteTools_UsesProvidedContext() {
	ctx := context.Background()
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	tool := &stubTool{name: "get_weather", result: "Sunny"}
	conv := NewConversation(ctx, &agentStub{tools: []contractsai.Tool{tool}}, &conversationProviderStub{}, "m", nil)

	results, err := conv.executeTools(streamCtx, []contractsai.Tool{tool}, []contractsai.ToolCall{{ID: "c1", Name: "get_weather", Args: map[string]any{"city": "London"}}})

	s.NoError(err)
	s.Equal([]contractsai.Message{{
		Role:       contractsai.RoleToolResult,
		Content:    "Sunny",
		ToolCallID: "c1",
	}}, results)
	s.Same(streamCtx, tool.lastCtx)
}

func (s *ConversationTestSuite) TestPromptMiddleware() {
	ctx := context.Background()

	s.Run("mutates prompt before provider call", func() {
		provider := &conversationToolProviderStub{
			promptFn: func(_ context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
				s.Equal("hello from middleware", prompt.Input)
				return &stubResponse{text: "done"}, nil
			},
		}

		middleware := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
			prompt.Input = prompt.Input + " from middleware"
			return next(ctx, prompt)
		})

		conv := NewConversation(ctx, &agentStub{}, provider, "m", []contractsai.Middleware{middleware})

		resp, err := conv.Prompt("hello")
		s.NoError(err)
		s.Equal("done", resp.Text())
	})

	s.Run("mutates response after provider call", func() {
		provider := &conversationToolProviderStub{
			promptFn: func(_ context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
				s.Equal("hello", prompt.Input)
				return &stubResponse{text: "done"}, nil
			},
		}

		middleware := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
			response, err := next(ctx, prompt)
			if err != nil {
				return nil, err
			}

			return &stubResponse{text: response.Text() + " after middleware"}, nil
		})

		conv := NewConversation(ctx, &agentStub{}, provider, "m", []contractsai.Middleware{middleware})

		resp, err := conv.Prompt("hello")
		s.NoError(err)
		s.Equal("done after middleware", resp.Text())
		s.Equal([]contractsai.Message{
			{Role: contractsai.RoleUser, Content: "hello"},
			{Role: contractsai.RoleAssistant, Content: "done after middleware"},
		}, conv.Messages())
	})

	s.Run("runs middleware in order", func() {
		var order []string
		provider := &conversationToolProviderStub{
			promptFn: func(_ context.Context, _ contractsai.AgentPrompt) (contractsai.Response, error) {
				order = append(order, "provider")
				return &stubResponse{text: "done"}, nil
			},
		}

		first := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
			order = append(order, "first-before")
			response, err := next(ctx, prompt)
			order = append(order, "first-after")
			return response, err
		})
		second := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
			order = append(order, "second-before")
			response, err := next(ctx, prompt)
			order = append(order, "second-after")
			return response, err
		})

		conv := NewConversation(ctx, &agentStub{}, provider, "m", []contractsai.Middleware{first, second})

		_, err := conv.Prompt("hello")
		s.NoError(err)
		s.Equal([]string{"first-before", "second-before", "provider", "second-after", "first-after"}, order)
	})

	s.Run("can short circuit provider", func() {
		calledProvider := false
		provider := &conversationToolProviderStub{
			promptFn: func(_ context.Context, _ contractsai.AgentPrompt) (contractsai.Response, error) {
				calledProvider = true
				return &stubResponse{text: "provider"}, nil
			},
		}

		middleware := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
			return &stubResponse{text: "short circuit"}, nil
		})

		conv := NewConversation(ctx, &agentStub{}, provider, "m", []contractsai.Middleware{middleware})

		resp, err := conv.Prompt("hello")
		s.NoError(err)
		s.False(calledProvider)
		s.Equal("short circuit", resp.Text())
	})

	s.Run("propagates provider errors through middleware", func() {
		provider := &conversationToolProviderStub{
			promptFn: func(_ context.Context, _ contractsai.AgentPrompt) (contractsai.Response, error) {
				return nil, assert.AnError
			},
		}

		middleware := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
			return next(ctx, prompt)
		})

		conv := NewConversation(ctx, &agentStub{}, provider, "m", []contractsai.Middleware{middleware})

		resp, err := conv.Prompt("hello")
		s.Equal(assert.AnError, err)
		s.Nil(resp)
		s.Empty(conv.Messages())
	})

	s.Run("skips typed nil middleware", func() {
		provider := &conversationToolProviderStub{
			promptFn: func(_ context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
				s.Equal("hello", prompt.Input)
				return &stubResponse{text: "done"}, nil
			},
		}

		var nilMiddleware *conversationNilTestMiddleware
		middleware := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
			return next(ctx, prompt)
		})

		conv := NewConversation(ctx, &agentStub{}, provider, "m", []contractsai.Middleware{nilMiddleware, middleware})

		resp, err := conv.Prompt("hello")
		s.NoError(err)
		s.Equal("done", resp.Text())
	})
}

func (s *ConversationTestSuite) TestPromptThen() {
	ctx := context.Background()
	provider := &conversationToolProviderStub{
		promptFn: func(_ context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
			s.Equal("hello", prompt.Input)
			return &stubResponse{text: "done"}, nil
		},
	}

	conv := NewConversation(ctx, &agentStub{}, provider, "m", nil)

	resp, err := conv.Prompt("hello")
	s.Require().NoError(err)

	called := 0
	returned := resp.Then(func(response contractsai.Response) {
		called++
		s.Equal("done", response.Text())
	})

	s.Equal("done", returned.Text())
	s.Equal(1, called)
}

func (s *ConversationTestSuite) TestSharedMiddlewareAcrossPromptAndStream() {
	ctx := context.Background()

	var (
		promptInputs []string
		streamInputs []string
		finalized    []string
	)

	middleware := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
		prompt.Input += " via middleware"

		response, err := next(ctx, prompt)
		if err != nil {
			return nil, err
		}

		return response.Then(func(response contractsai.Response) {
			finalized = append(finalized, response.Text())
		}), nil
	})

	promptProvider := &conversationToolProviderStub{
		promptFn: func(_ context.Context, prompt contractsai.AgentPrompt) (contractsai.Response, error) {
			promptInputs = append(promptInputs, prompt.Input)
			return &stubResponse{text: "prompt done"}, nil
		},
	}
	promptConv := NewConversation(ctx, &agentStub{}, promptProvider, "m", []contractsai.Middleware{middleware})

	resp, err := promptConv.Prompt("hello")
	s.Require().NoError(err)
	s.Equal("prompt done", resp.Text())
	s.Equal([]string{"hello via middleware"}, promptInputs)
	s.Equal([]string{"prompt done"}, finalized)

	streamProvider := &conversationToolProviderStub{
		streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
			streamInputs = append(streamInputs, prompt.Input)
			return NewStreamableResponse(gotCtx, func(_ context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
				if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "partial"}); err != nil {
					return nil, err
				}
				s.Equal([]string{"prompt done"}, finalized)
				if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeDone}); err != nil {
					return nil, err
				}

				return &stubResponse{text: "stream done"}, nil
			}), nil
		},
	}
	streamConv := NewConversation(ctx, &agentStub{}, streamProvider, "m", []contractsai.Middleware{middleware})

	stream, err := streamConv.Stream("hello")
	s.Require().NoError(err)
	s.Equal([]string{"prompt done"}, finalized)

	err = stream.Each(func(event contractsai.StreamEvent) error {
		return nil
	})
	s.NoError(err)
	s.Equal([]string{"hello via middleware"}, streamInputs)
	s.Equal([]string{"prompt done", "stream done"}, finalized)
}

func (s *ConversationTestSuite) TestStreamShortCircuitMiddlewareCommitsMessages() {
	ctx := context.Background()
	providerCalled := false
	provider := &conversationToolProviderStub{
		streamFn: func(context.Context, contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
			providerCalled = true
			return nil, nil
		},
	}

	middleware := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
		return &stubResponse{text: "short stream"}, nil
	})

	conv := NewConversation(ctx, &agentStub{}, provider, "m", []contractsai.Middleware{middleware})

	stream, err := conv.Stream("hello")
	s.Require().NoError(err)

	var events []contractsai.StreamEvent
	err = stream.Each(func(event contractsai.StreamEvent) error {
		events = append(events, event)
		return nil
	})

	s.NoError(err)
	s.False(providerCalled)
	s.Equal([]contractsai.StreamEvent{{Type: contractsai.StreamEventTypeDone}}, events)
	s.Equal([]contractsai.Message{
		{Role: contractsai.RoleUser, Content: "hello"},
		{Role: contractsai.RoleAssistant, Content: "short stream"},
	}, conv.Messages())
}

func (s *ConversationTestSuite) TestStreamShortCircuitMiddlewareRequiresResponse() {
	ctx := context.Background()
	providerCalled := false
	provider := &conversationToolProviderStub{
		streamFn: func(context.Context, contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
			providerCalled = true
			return nil, nil
		},
	}

	middleware := promptMiddlewareFunc(func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
		return nil, nil
	})

	conv := NewConversation(ctx, &agentStub{}, provider, "m", []contractsai.Middleware{middleware})

	stream, err := conv.Stream("hello")
	s.Equal(errors.AIResponseIsNil, err)
	s.Nil(stream)
	s.False(providerCalled)
	s.Empty(conv.Messages())
}

// stubResponse is a minimal Response with configurable text and tool calls.
type stubResponse struct {
	text      string
	toolCalls []contractsai.ToolCall
	usage     contractsai.Usage
}

func (r *stubResponse) Text() string                      { return r.text }
func (r *stubResponse) Usage() contractsai.Usage          { return r.usage }
func (r *stubResponse) ToolCalls() []contractsai.ToolCall { return r.toolCalls }
func (r *stubResponse) Then(callback func(contractsai.Response)) contractsai.Response {
	if callback == nil {
		return r
	}

	callback(r)

	return r
}

type promptMiddlewareFunc func(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error)

func (f promptMiddlewareFunc) Handle(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
	return f(ctx, prompt, next)
}

type conversationNilTestMiddleware struct{}

func (m *conversationNilTestMiddleware) Handle(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
	return next(ctx, prompt)
}

// stubTool records calls and returns a fixed result or error.
type stubTool struct {
	name      string
	result    string
	execErr   error
	callCount int
	lastCtx   context.Context
}

func (t *stubTool) Name() string               { return t.name }
func (t *stubTool) Description() string        { return "" }
func (t *stubTool) Parameters() map[string]any { return nil }
func (t *stubTool) Execute(ctx context.Context, _ map[string]any) (string, error) {
	t.callCount++
	t.lastCtx = ctx
	return t.result, t.execErr
}
