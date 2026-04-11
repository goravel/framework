package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsai "github.com/goravel/framework/contracts/ai"
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
		conv         *conversation
	)

	beforeEach := func(initial []contractsai.Message, model string) {
		mockAgent := mocksai.NewAgent(s.T())
		mockAgent.EXPECT().Messages().Return(initial).Once()
		mockProvider = mocksai.NewProvider(s.T())
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
				mockResponse := mocksai.NewResponse(s.T())
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
				mockResponse := mocksai.NewResponse(s.T())
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
		conv         *conversation
	)

	beforeEach := func(initial []contractsai.Message) {
		mockAgent := mocksai.NewAgent(s.T())
		mockAgent.EXPECT().Messages().Return(initial).Times(2)
		mockProvider = mocksai.NewProvider(s.T())
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
				response := mocksai.NewResponse(s.T())
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

func (s *ConversationTestSuite) TestStream() {
	ctx := context.Background()
	model := "stream-model"

	s.Run("returns provider error without mutating messages", func() {
		initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}
		mockAgent := mocksai.NewAgent(s.T())
		mockAgent.EXPECT().Messages().Return(initial).Once()

		var conv *conversation
		provider := &conversationProviderStub{
			streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				s.Equal(ctx, gotCtx)
				s.Equal(contractsai.AgentPrompt{Agent: conv, Input: "hello", Model: model}, prompt)
				return nil, assert.AnError
			},
		}
		conv = NewConversation(ctx, mockAgent, provider, model)

		stream, err := conv.Stream("hello")

		s.Equal(assert.AnError, err)
		s.Nil(stream)
		s.Equal(initial, conv.Messages())
	})

	s.Run("appends messages after successful stream completion", func() {
		initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}
		mockAgent := mocksai.NewAgent(s.T())
		mockAgent.EXPECT().Messages().Return(initial).Once()

		var conv *conversation
		provider := &conversationProviderStub{
			streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				s.Equal(ctx, gotCtx)
				s.Equal(contractsai.AgentPrompt{Agent: conv, Input: "hi", Model: model}, prompt)
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
		mockAgent := mocksai.NewAgent(s.T())
		mockAgent.EXPECT().Messages().Return(initial).Once()

		var conv *conversation
		provider := &conversationProviderStub{
			streamFn: func(gotCtx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
				s.Equal(ctx, gotCtx)
				s.Equal(contractsai.AgentPrompt{Agent: conv, Input: "hi", Model: model}, prompt)
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

	mockAgent := mocksai.NewAgent(s.T())
	mockAgent.EXPECT().Messages().Return(initial).Once()
	conv := NewConversation(ctx, mockAgent, &conversationProviderStub{}, "model")

	initial[0].Content = "mutated"
	initial = append(initial, contractsai.Message{Role: contractsai.RoleUser, Content: "new"})

	got := conv.Messages()
	s.Equal([]contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}, got)

	got[0].Content = "changed"
	s.Equal([]contractsai.Message{{Role: contractsai.RoleAssistant, Content: "seed"}}, conv.Messages())
}
