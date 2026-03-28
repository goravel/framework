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
