package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsai "github.com/goravel/framework/contracts/ai"
	mocksai "github.com/goravel/framework/mocks/ai"
)

func TestConversation_Prompt(t *testing.T) {
	ctx := context.Background()

	var (
		mockProvider *mocksai.Provider
		conv         *conversation
	)

	beforeEach := func(initial []contractsai.Message, model string) {
		mockAgent := mocksai.NewAgent(t)
		mockAgent.EXPECT().Messages().Return(initial).Once()
		mockProvider = mocksai.NewProvider(t)
		conv = NewConversation(ctx, mockAgent, mockProvider, model)
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
				mockResponse := mocksai.NewResponse(t)
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
				mockResponse := mocksai.NewResponse(t)
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
		t.Run(tt.name, func(t *testing.T) {
			beforeEach(tt.initial, tt.model)
			expectResp := tt.setup()

			resp, err := conv.Prompt(tt.input)
			assert.Equal(t, tt.expectError, err)
			assert.Equal(t, expectResp, resp)
			assert.Equal(t, tt.expectMessages, conv.Messages())
		})
	}
}

func TestConversation_Reset(t *testing.T) {
	ctx := context.Background()

	var (
		mockProvider *mocksai.Provider
		conv         *conversation
	)

	beforeEach := func(initial []contractsai.Message) {
		mockAgent := mocksai.NewAgent(t)
		mockAgent.EXPECT().Messages().Return(initial).Times(2)
		mockProvider = mocksai.NewProvider(t)
		conv = NewConversation(ctx, mockAgent, mockProvider, "model-z")
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
				response := mocksai.NewResponse(t)
				mockProvider.EXPECT().Prompt(ctx, contractsai.AgentPrompt{Agent: conv, Input: "append", Model: "model-z"}).Return(response, nil).Once()
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
		t.Run(tt.name, func(t *testing.T) {
			beforeEach(tt.initial)
			tt.setup()

			if tt.promptBefore {
				_, err := conv.Prompt(tt.input)
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectBefore, conv.Messages())
			conv.Reset()
			resetMessages := conv.Messages()
			assert.Equal(t, tt.expectAfter, resetMessages)
			assert.NotSame(t, &tt.initial[0], &resetMessages[0])
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

func TestConversation_Stream(t *testing.T) {
	ctx := context.Background()
	model := "stream-model"

	t.Run("returns provider error without mutating messages", func(t *testing.T) {
		initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}
		mockAgent := mocksai.NewAgent(t)
		mockAgent.EXPECT().Messages().Return(initial).Once()

		var conv *conversation
		provider := &conversationProviderStub{
			streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				assert.Equal(t, ctx, gotCtx)
				assert.Equal(t, contractsai.AgentPrompt{Agent: conv, Input: "hello", Model: model}, prompt)
				return nil, assert.AnError
			},
		}
		conv = NewConversation(ctx, mockAgent, provider, model)

		stream, err := conv.Stream("hello")

		assert.Equal(t, assert.AnError, err)
		assert.Nil(t, stream)
		assert.Equal(t, initial, conv.Messages())
	})

	t.Run("appends messages after successful stream completion", func(t *testing.T) {
		initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}
		mockAgent := mocksai.NewAgent(t)
		mockAgent.EXPECT().Messages().Return(initial).Once()

		var conv *conversation
		provider := &conversationProviderStub{
			streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				assert.Equal(t, ctx, gotCtx)
				assert.Equal(t, contractsai.AgentPrompt{Agent: conv, Input: "hi", Model: model}, prompt)
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
		conv = NewConversation(ctx, mockAgent, provider, model)

		stream, err := conv.Stream("hi")
		require.NoError(t, err)
		require.NotNil(t, stream)
		assert.Equal(t, initial, conv.Messages())

		var events []contractsai.StreamEvent
		err = stream.Each(func(event contractsai.StreamEvent) error {
			events = append(events, event)
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, normalizeStreamEvents([]contractsai.StreamEvent{
			{Type: contractsai.StreamEventTypeTextDelta, Delta: "partial"},
			{Type: contractsai.StreamEventTypeDone},
		}), normalizeStreamEvents(events))
		assert.Equal(t, []contractsai.Message{
			{Role: contractsai.RoleAssistant, Content: "seed"},
			{Role: contractsai.RoleUser, Content: "hi"},
			{Role: contractsai.RoleAssistant, Content: "assistant reply"},
		}, conv.Messages())
	})

	t.Run("does not append messages when stream completes with error", func(t *testing.T) {
		initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}
		mockAgent := mocksai.NewAgent(t)
		mockAgent.EXPECT().Messages().Return(initial).Once()

		var conv *conversation
		provider := &conversationProviderStub{
			streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				assert.Equal(t, ctx, gotCtx)
				assert.Equal(t, contractsai.AgentPrompt{Agent: conv, Input: "hi", Model: model}, prompt)
				return NewStreamableResponse(gotCtx, func(_ context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
					if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "partial"}); err != nil {
						return nil, err
					}

					return nil, assert.AnError
				}), nil
			},
		}
		conv = NewConversation(ctx, mockAgent, provider, model)

		stream, err := conv.Stream("hi")
		require.NoError(t, err)
		require.NotNil(t, stream)

		err = stream.Each(nil)
		assert.Equal(t, assert.AnError, err)
		assert.Equal(t, initial, conv.Messages())
	})
}

func TestConversation_MessagesClone(t *testing.T) {
	ctx := context.Background()
	initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}

	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Messages().Return(initial).Once()
	conv := NewConversation(ctx, mockAgent, &conversationProviderStub{}, "model")

	initial[0].Content = "mutated"
	initial = append(initial, contractsai.Message{Role: contractsai.RoleUser, Content: "new"})

	got := conv.Messages()
	assert.Equal(t, []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}, got)

	got[0].Content = "changed"
	assert.Equal(t, []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}, conv.Messages())
}
