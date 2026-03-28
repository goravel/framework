package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsai "github.com/goravel/framework/contracts/ai"
	mocksai "github.com/goravel/framework/mocks/ai"
)

func TestApplication_Agent(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name            string
		promptInput     string
		options         []contractsai.Option
		setupConfig     func(t *testing.T) (contractsai.Config, *mocksai.Provider)
		expectedModel   string
		responseText    string
		expectResponse  bool
		promptErr       error
		expectPromptErr bool
	}{
		{
			name:        "default provider",
			promptInput: "ping",
			setupConfig: func(t *testing.T) (contractsai.Config, *mocksai.Provider) {
				provider := mocksai.NewProvider(t)
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: provider},
					},
				}, provider
			},
			responseText:   "ok",
			expectResponse: true,
		},
		{
			name:        "provider override",
			promptInput: "override",
			options:     []contractsai.Option{WithProvider("alternative")},
			setupConfig: func(t *testing.T) (contractsai.Config, *mocksai.Provider) {
				defaultProvider := mocksai.NewProvider(t)
				alternativeProvider := mocksai.NewProvider(t)
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default":     {Via: defaultProvider},
						"alternative": {Via: alternativeProvider},
					},
				}, alternativeProvider
			},
			responseText:   "override",
			expectResponse: true,
		},
		{
			name:        "model option",
			promptInput: "any",
			options:     []contractsai.Option{WithModel("custom-model")},
			setupConfig: func(t *testing.T) (contractsai.Config, *mocksai.Provider) {
				provider := mocksai.NewProvider(t)
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: provider},
					},
				}, provider
			},
			expectedModel:  "custom-model",
			responseText:   "modelled",
			expectResponse: true,
		},
		{
			name:        "provider error",
			promptInput: "fail",
			setupConfig: func(t *testing.T) (contractsai.Config, *mocksai.Provider) {
				provider := mocksai.NewProvider(t)
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: provider},
					},
				}, provider
			},
			expectPromptErr: true,
			promptErr:       assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, provider := tt.setupConfig(t)
			agent := mocksai.NewAgent(t)
			agent.EXPECT().Messages().Return(nil).Once()

			app := NewApplication(ctx, config)
			conv, err := app.Agent(agent, tt.options...)
			assert.NoError(t, err)

			convImpl, ok := conv.(*conversation)
			assert.True(t, ok)

			expectedPrompt := contractsai.AgentPrompt{
				Agent: convImpl,
				Input: tt.promptInput,
				Model: tt.expectedModel,
			}

			var response *mocksai.Response
			if tt.expectResponse {
				response = mocksai.NewResponse(t)
				response.EXPECT().Text().Return(tt.responseText).Once()
			}

			provider.EXPECT().
				Prompt(ctx, expectedPrompt).
				Return(response, tt.promptErr).
				Once()

			resp, err := conv.Prompt(tt.promptInput)
			if tt.expectPromptErr {
				assert.Equal(t, tt.promptErr, err)
				assert.Nil(t, resp)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, response, resp)
		})
	}
}

func TestApplication_Agent_ResolverError(t *testing.T) {
	ctx := context.Background()
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {
				Via: func() (contractsai.Provider, error) {
					return nil, assert.AnError
				},
			},
		},
	}

	app := NewApplication(ctx, config)
	_, err := app.Agent(mocksai.NewAgent(t))
	assert.Equal(t, assert.AnError, err)
}

type testCtxKey string

func TestApplication_WithContext(t *testing.T) {
	origCtx := context.WithValue(context.Background(), testCtxKey("orig"), true)
	provider := mocksai.NewProvider(t)
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: provider},
		},
	}

	app := NewApplication(origCtx, config)
	newCtx := context.WithValue(context.Background(), testCtxKey("orig"), false)
	aiWithCtx := app.WithContext(newCtx)
	aiImpl, ok := aiWithCtx.(*Application)
	assert.True(t, ok)

	assert.Same(t, newCtx, aiImpl.ctx)
	assert.Same(t, app.resolver, aiImpl.resolver)
	assert.Equal(t, app.config, aiImpl.config)
}
