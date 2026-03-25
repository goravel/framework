package ai

import (
	"context"
	"testing"

	contractsai "github.com/goravel/framework/contracts/ai"
	mockai "github.com/goravel/framework/mocks/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuntimeAgent(t *testing.T) {
	initial := []contractsai.Message{{Role: contractsai.RoleUser, Content: "welcome"}}
	agent := mockai.NewAgent(t)
	agent.EXPECT().Messages().Return(initial).Once()

	runtime := newRuntimeAgent(agent)
	runtimeMessages := runtime.Messages()
	require.Len(t, runtimeMessages, 1)
	assert.Equal(t, initial, runtimeMessages)
	assert.NotSame(t, &initial[0], &runtimeMessages[0])

	runtimeMessages[0].Content = "changed"
	assert.Equal(t, []contractsai.Message{{Role: contractsai.RoleUser, Content: "welcome"}}, initial)
}

func TestConversation_Prompt(t *testing.T) {
	ctx := context.Background()

	var (
		mockProvider *mockai.Provider
		conv         *conversation
	)

	beforeEach := func(initial []contractsai.Message, model string) {
		mockAgent := mockai.NewAgent(t)
		mockAgent.EXPECT().Messages().Return(initial).Once()
		mockProvider = mockai.NewProvider(t)
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
				mockResponse := mockai.NewResponse(t)
				mockResponse.EXPECT().Text().Return("got it").Once()
				mockProvider.EXPECT().Prompt(ctx, contractsai.AgentPrompt{Agent: conv.agent, Input: "hello", Model: "model-x"}).Return(mockResponse, nil).Once()
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
				mockProvider.EXPECT().Prompt(ctx, contractsai.AgentPrompt{Agent: conv.agent, Input: "fail", Model: "model-y"}).Return(nil, assert.AnError).Once()
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
				mockResponse := mockai.NewResponse(t)
				mockResponse.EXPECT().Text().Return("").Once()
				mockProvider.EXPECT().Prompt(ctx, contractsai.AgentPrompt{Agent: conv.agent, Input: "", Model: "model-empty"}).Return(mockResponse, nil).Once()
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
		mockProvider *mockai.Provider
		conv         *conversation
	)

	beforeEach := func(initial []contractsai.Message) {
		mockAgent := mockai.NewAgent(t)
		mockAgent.EXPECT().Messages().Return(initial).Times(2)
		mockProvider = mockai.NewProvider(t)
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
				response := mockai.NewResponse(t)
				mockProvider.EXPECT().Prompt(ctx, contractsai.AgentPrompt{Agent: conv.agent, Input: "append", Model: "model-z"}).Return(response, nil).Once()
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
