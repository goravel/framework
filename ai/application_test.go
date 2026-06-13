package ai

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
	mocksai "github.com/goravel/framework/mocks/ai"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
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
			agent.EXPECT().Middleware().Return(nil).Once()

			app := NewApplication(ctx, config)
			conv, err := app.Agent(agent, tt.options...)
			assert.NoError(t, err)

			convImpl, ok := conv.(*conversation)
			assert.True(t, ok)

			expectedPrompt := contractsai.AgentPrompt{
				Agent:         convImpl,
				Input:         tt.promptInput,
				Model:         tt.expectedModel,
				ProviderState: convImpl.providerState,
			}

			agent.EXPECT().Tools().Return(nil).Once()

			var response *mocksai.AgentResponse
			if tt.expectResponse {
				response = mocksai.NewAgentResponse(t)
				response.EXPECT().ToolCalls().Return(nil).Once()
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

func TestApplication_Agent_Failover(t *testing.T) {
	failoverErr := NewFailoverError("primary", contractsai.FailoverReason("rate_limited"), assert.AnError)
	primaryProvider := &applicationPromptProviderStub{name: "primary", err: failoverErr}
	backupResponse := &stubResponse{text: "backup response"}
	backupProvider := &applicationPromptProviderStub{name: "backup", response: backupResponse}
	config := contractsai.Config{
		Default: "primary",
		Providers: map[string]contractsai.ProviderConfig{
			"primary": {Via: primaryProvider},
			"backup":  {Via: backupProvider},
		},
	}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Messages().Return(nil).Once()
	mockAgent.EXPECT().Middleware().Return(nil).Once()
	mockAgent.EXPECT().Tools().Return(nil).Once()

	conversation, err := NewApplication(context.Background(), config).Agent(mockAgent, WithProvider("primary", "backup"))
	require.NoError(t, err)

	response, err := conversation.Prompt("hello")
	require.NoError(t, err)
	assert.Equal(t, backupResponse, response)
	assert.Len(t, primaryProvider.prompts, 1)
	assert.Len(t, backupProvider.prompts, 1)
	assert.Nil(t, primaryProvider.previousState)
	assert.Nil(t, backupProvider.previousState)
}

func TestApplication_Agent_FailoverStopsOnNonFailoverError(t *testing.T) {
	primaryProvider := &applicationPromptProviderStub{name: "primary", err: assert.AnError}
	backupProvider := &applicationPromptProviderStub{name: "backup", response: &stubResponse{text: "backup response"}}
	config := contractsai.Config{
		Default: "primary",
		Providers: map[string]contractsai.ProviderConfig{
			"primary": {Via: primaryProvider},
			"backup":  {Via: backupProvider},
		},
	}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Messages().Return(nil).Once()
	mockAgent.EXPECT().Middleware().Return(nil).Once()
	mockAgent.EXPECT().Tools().Return(nil).Once()

	conversation, err := NewApplication(context.Background(), config).Agent(mockAgent, WithProvider("primary", "backup"))
	require.NoError(t, err)

	response, err := conversation.Prompt("hello")
	assert.Equal(t, assert.AnError, err)
	assert.Nil(t, response)
	assert.Len(t, primaryProvider.prompts, 1)
	assert.Empty(t, backupProvider.prompts)
}

func TestApplication_Agent_FailoverReturnsLastError(t *testing.T) {
	primaryErr := NewFailoverError("primary", contractsai.FailoverReason("rate_limited"), assert.AnError)
	backupErr := NewFailoverError("backup", contractsai.FailoverReason("provider_overloaded"), assert.AnError)
	primaryProvider := &applicationPromptProviderStub{name: "primary", err: primaryErr}
	backupProvider := &applicationPromptProviderStub{name: "backup", err: backupErr}
	config := contractsai.Config{
		Default: "primary",
		Providers: map[string]contractsai.ProviderConfig{
			"primary": {Via: primaryProvider},
			"backup":  {Via: backupProvider},
		},
	}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Messages().Return(nil).Once()
	mockAgent.EXPECT().Middleware().Return(nil).Once()
	mockAgent.EXPECT().Tools().Return(nil).Once()

	conversation, err := NewApplication(context.Background(), config).Agent(mockAgent, WithProvider("primary", "backup"))
	require.NoError(t, err)

	response, err := conversation.Prompt("hello")
	assert.Equal(t, backupErr, err)
	assert.Nil(t, response)
}

func TestApplication_Agent_StreamFailover(t *testing.T) {
	failoverErr := NewFailoverError("primary", contractsai.FailoverReason("rate_limited"), assert.AnError)
	primaryProvider := &applicationPromptProviderStub{name: "primary", streamErr: failoverErr}
	backupProvider := &applicationPromptProviderStub{name: "backup", streamResponse: &stubResponse{text: "backup response"}, streamEvents: []contractsai.StreamEvent{
		{Type: contractsai.StreamEventTypeTextDelta, Delta: "backup"},
		{Type: contractsai.StreamEventTypeDone},
	}}
	config := contractsai.Config{
		Default: "primary",
		Providers: map[string]contractsai.ProviderConfig{
			"primary": {Via: primaryProvider},
			"backup":  {Via: backupProvider},
		},
	}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Messages().Return(nil).Once()
	mockAgent.EXPECT().Middleware().Return(nil).Once()
	mockAgent.EXPECT().Tools().Return(nil).Once()

	conversation, err := NewApplication(context.Background(), config).Agent(mockAgent, WithProvider("primary", "backup"))
	require.NoError(t, err)
	stream, err := conversation.Stream("hello")
	require.NoError(t, err)

	var events []contractsai.StreamEvent
	err = stream.Each(func(event contractsai.StreamEvent) error {
		events = append(events, event)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []contractsai.StreamEvent{
		{Type: contractsai.StreamEventTypeTextDelta, Delta: "backup"},
		{Type: contractsai.StreamEventTypeDone},
	}, events)
	assert.Len(t, primaryProvider.streamPrompts, 1)
	assert.Len(t, backupProvider.streamPrompts, 1)
}

func TestApplication_Agent_StreamDoesNotFailoverAfterOutput(t *testing.T) {
	failoverErr := NewFailoverError("primary", contractsai.FailoverReason("rate_limited"), assert.AnError)
	primaryProvider := &applicationPromptProviderStub{name: "primary", streamErr: failoverErr, streamEvents: []contractsai.StreamEvent{
		{Type: contractsai.StreamEventTypeTextDelta, Delta: "partial"},
	}}
	backupProvider := &applicationPromptProviderStub{name: "backup", streamResponse: &stubResponse{text: "backup response"}, streamEvents: []contractsai.StreamEvent{
		{Type: contractsai.StreamEventTypeTextDelta, Delta: "backup"},
	}}
	config := contractsai.Config{
		Default: "primary",
		Providers: map[string]contractsai.ProviderConfig{
			"primary": {Via: primaryProvider},
			"backup":  {Via: backupProvider},
		},
	}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Messages().Return(nil).Once()
	mockAgent.EXPECT().Middleware().Return(nil).Once()
	mockAgent.EXPECT().Tools().Return(nil).Once()

	conversation, err := NewApplication(context.Background(), config).Agent(mockAgent, WithProvider("primary", "backup"))
	require.NoError(t, err)
	stream, err := conversation.Stream("hello")
	require.NoError(t, err)

	var events []contractsai.StreamEvent
	err = stream.Each(func(event contractsai.StreamEvent) error {
		events = append(events, event)
		return nil
	})
	assert.Equal(t, failoverErr, err)
	assert.Equal(t, []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeTextDelta, Delta: "partial"}}, events)
	assert.Len(t, primaryProvider.streamPrompts, 1)
	assert.Empty(t, backupProvider.streamPrompts)
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

func TestApplication_Agent_WithMiddleware(t *testing.T) {
	ctx := context.Background()
	mockProvider := mocksai.NewProvider(t)
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: mockProvider},
		},
	}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Messages().Return(nil).Once()
	mockAgent.EXPECT().Middleware().Return(nil).Once()
	mockAgent.EXPECT().Tools().Return(nil).Once()

	middleware := &applicationTestMiddleware{}

	app := NewApplication(ctx, config)
	conv, err := app.Agent(mockAgent, WithMiddleware(middleware))
	assert.NoError(t, err)
	convImpl, ok := conv.(*conversation)
	assert.True(t, ok)

	mockProvider.EXPECT().
		Prompt(ctx, contractsai.AgentPrompt{Agent: convImpl, Input: "hello", Tools: nil, ProviderState: convImpl.providerState}).
		Return(&stubResponse{text: "before middleware"}, nil).
		Once()

	resp, err := conv.Prompt("hello")
	assert.NoError(t, err)
	assert.Equal(t, "before middleware after middleware", resp.Text())
}

func TestApplication_Agent_WithDefaultMiddleware(t *testing.T) {
	ctx := context.Background()
	mockProvider := mocksai.NewProvider(t)
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: mockProvider},
		},
	}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Messages().Return(nil).Once()
	mockAgent.EXPECT().Middleware().Return([]contractsai.Middleware{&applicationTestMiddleware{}}).Once()
	mockAgent.EXPECT().Tools().Return(nil).Once()

	app := NewApplication(ctx, config)
	conv, err := app.Agent(mockAgent)
	assert.NoError(t, err)
	convImpl, ok := conv.(*conversation)
	assert.True(t, ok)

	mockProvider.EXPECT().
		Prompt(ctx, contractsai.AgentPrompt{Agent: convImpl, Input: "hello", Tools: nil, ProviderState: convImpl.providerState}).
		Return(&stubResponse{text: "before middleware"}, nil).
		Once()

	resp, err := conv.Prompt("hello")
	assert.NoError(t, err)
	assert.Equal(t, "before middleware after middleware", resp.Text())
}

func TestApplication_Agent_MergesDefaultMiddlewareWithOptions(t *testing.T) {
	ctx := context.Background()
	mockProvider := mocksai.NewProvider(t)
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: mockProvider},
		},
	}
	mockAgent := mocksai.NewAgent(t)
	mockAgent.EXPECT().Messages().Return(nil).Once()
	mockAgent.EXPECT().Middleware().Return([]contractsai.Middleware{&applicationTestMiddleware{}}).Once()
	mockAgent.EXPECT().Tools().Return(nil).Once()

	app := NewApplication(ctx, config)
	conv, err := app.Agent(mockAgent, WithMiddleware(&applicationTestMiddleware{}))
	assert.NoError(t, err)
	convImpl, ok := conv.(*conversation)
	assert.True(t, ok)

	mockProvider.EXPECT().
		Prompt(ctx, contractsai.AgentPrompt{Agent: convImpl, Input: "hello", Tools: nil, ProviderState: convImpl.providerState}).
		Return(&stubResponse{text: "before middleware"}, nil).
		Once()

	resp, err := conv.Prompt("hello")
	assert.NoError(t, err)
	assert.Equal(t, "before middleware after middleware after middleware", resp.Text())
}

func TestApplication_putFile(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		options     []contractsai.Option
		setup       func(t *testing.T, ctx context.Context, file contractsai.StorableFile) contractsai.Config
		expectError error
		expectID    string
	}{
		{
			name:    "success",
			ctx:     context.WithValue(context.Background(), testCtxKey("upload"), "success"),
			options: []contractsai.Option{WithProvider("openai")},
			setup: func(t *testing.T, ctx context.Context, file contractsai.StorableFile) contractsai.Config {
				fileProvider := mocksai.NewFileProvider(t)
				response := mocksai.NewFileResponse(t)
				response.EXPECT().ID().Return("file-123").Once()
				fileProvider.EXPECT().PutFile(ctx, file).Return(response, nil).Once()

				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: mocksai.NewProvider(t)},
						"openai":  {Via: uploadTestProvider{fileProvider: fileProvider}},
					},
				}
			},
			expectID: "file-123",
		},
		{
			name: "provider does not support files",
			ctx:  context.Background(),
			setup: func(t *testing.T, _ context.Context, _ contractsai.StorableFile) contractsai.Config {
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: mocksai.NewProvider(t)},
					},
				}
			},
			expectError: errors.AIProviderDoesNotSupportFiles.Args("default"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := mocksai.NewStorableFile(t)
			config := tt.setup(t, tt.ctx, file)

			app := NewApplication(context.Background(), config)
			stored, err := app.putFile(tt.ctx, file, tt.options...)
			assert.Equal(t, tt.expectError, err)
			if tt.expectError != nil {
				assert.Nil(t, stored)
				return
			}

			require.NotNil(t, stored)
			assert.Equal(t, tt.expectID, stored.ID())
		})
	}
}

func TestApplication_Image(t *testing.T) {
	ctx := context.Background()
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: mocksai.NewProvider(t)},
		},
	}

	app := NewApplication(ctx, config)
	request := app.Image("draw a cat").Provider("default").Model("gpt-image-1")

	req, ok := request.(*imageRequest)
	assert.True(t, ok)
	assert.Equal(t, ctx, req.ctx)
	assert.Equal(t, app, req.app)
	assert.Equal(t, "draw a cat", req.prompt)
	assert.Equal(t, "default", req.provider)
	assert.Equal(t, "gpt-image-1", req.model)

	assert.Same(t, req, request.Square())
	assert.Same(t, req, request.Portrait())
	assert.Same(t, req, request.Landscape())
	assert.Same(t, req, request.Quality(contractsai.ImageQualityHigh))
	assert.Same(t, req, request.Timeout(2*time.Second))

	attachment := ImageFromByte([]byte("image"), WithMimeType("image/png"))
	assert.Same(t, req, request.Attachments(attachment))
	assert.Equal(t, contractsai.ImageSizeLandscape, req.size)
	assert.Equal(t, contractsai.ImageQualityHigh, req.quality)
	assert.Equal(t, 2*time.Second, req.timeout)
	assert.Equal(t, []contractsai.Attachment{attachment}, req.attachments)
}

func TestImageRequest_Generate(t *testing.T) {
	ctx := context.Background()
	provider := &applicationImageProviderStub{}
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: provider},
		},
	}

	app := NewApplication(context.Background(), config)
	attachment := ImageFromByte([]byte("image"), WithMimeType("image/png"))
	response := &applicationImageResponseStub{}
	provider.response = response

	result, err := app.Image("draw a cat").
		Landscape().
		Quality(contractsai.ImageQualityHigh).
		Attachments(attachment).
		Timeout(3 * time.Second).
		Generate()

	require.NoError(t, err)
	assert.Equal(t, response, result)
	assert.Equal(t, ctx, provider.ctx)
	assert.Equal(t, contractsai.ImagePrompt{
		Prompt:      "draw a cat",
		Size:        contractsai.ImageSizeLandscape,
		Quality:     contractsai.ImageQualityHigh,
		Attachments: []contractsai.Attachment{attachment},
		Timeout:     3 * time.Second,
	}, provider.prompt)
}

func TestImageRequest_Store(t *testing.T) {
	ctx := context.Background()
	provider := &applicationImageProviderStub{}
	storage := mocksfilesystem.NewStorage(t)
	previousStorageFacade := storageFacade
	storageFacade = storage
	t.Cleanup(func() {
		storageFacade = previousStorageFacade
	})

	app := NewApplication(context.Background(), contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: provider},
		},
	})
	response := &applicationImageResponseStub{}
	provider.response = response

	driver := mocksfilesystem.NewDriver(t)
	storage.EXPECT().Disk("s3").Return(driver).Once()
	driver.EXPECT().Put("generated.png", "image").Return(nil).Once()

	path, err := app.Image("draw a cat").Store("s3")

	require.NoError(t, err)
	assert.Equal(t, "generated.png", path)
	assert.Equal(t, ctx, provider.ctx)
	assert.Equal(t, "draw a cat", provider.prompt.Prompt)
	assert.Equal(t, 1, response.storeCalls)
	assert.Equal(t, 0, response.storeAsCalls)
	assert.Equal(t, []string{"s3"}, response.storePath)
}

func TestImageRequest_StoreUsesResponseStore(t *testing.T) {
	provider := &applicationImageProviderStub{}
	app := NewApplication(context.Background(), contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: provider},
		},
	})
	response := &applicationImageResponseStub{storePathResult: "images/generated.png"}
	provider.response = response

	path, err := app.Image("draw a cat").Store()

	require.NoError(t, err)
	assert.Equal(t, "images/generated.png", path)
	assert.Equal(t, 1, response.storeCalls)
	assert.Equal(t, 0, response.storeAsCalls)
	assert.Empty(t, response.storePath)
}

func TestApplication_Audio(t *testing.T) {
	ctx := context.Background()
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: mocksai.NewProvider(t)},
		},
	}

	app := NewApplication(ctx, config)
	request := app.Audio("welcome to goravel").Provider("default").Model("gpt-4o-mini-tts").Male().Instructions("Speak slowly").Timeout(2 * time.Second)

	req, ok := request.(*audioRequest)
	assert.True(t, ok)
	assert.Equal(t, ctx, req.ctx)
	assert.Equal(t, app, req.app)
	assert.Equal(t, "welcome to goravel", req.prompt)
	assert.Equal(t, "default", req.provider)
	assert.Equal(t, "gpt-4o-mini-tts", req.model)
	assert.Equal(t, DefaultMaleVoice, req.voice)
	assert.Equal(t, "Speak slowly", req.instructions)
	assert.Equal(t, 2*time.Second, req.timeout)

	assert.Same(t, req, request.Female())
	assert.Equal(t, DefaultFemaleVoice, req.voice)
}

func TestApplication_Transcription(t *testing.T) {
	ctx := context.Background()
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: mocksai.NewProvider(t)},
		},
	}
	file := mocksai.NewStorableFile(t)

	app := NewApplication(ctx, config)
	request := app.Transcription(file).Provider("default").Model("gpt-4o-mini-transcribe").Language("en").Diarize().Timeout(2 * time.Second)

	req, ok := request.(*transcriptionRequest)
	assert.True(t, ok)
	assert.Equal(t, ctx, req.ctx)
	assert.Equal(t, app, req.app)
	assert.Equal(t, file, req.file)
	assert.Equal(t, "default", req.provider)
	assert.Equal(t, "gpt-4o-mini-transcribe", req.model)
	assert.Equal(t, "en", req.language)
	assert.True(t, req.diarize)
	assert.Equal(t, 2*time.Second, req.timeout)
}

func TestTranscriptionRequest_Generate(t *testing.T) {
	ctx := context.Background()
	provider := &applicationTranscriptionProviderStub{}
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: provider},
		},
	}
	file := mocksai.NewStorableFile(t)

	app := NewApplication(context.Background(), config)
	response := &applicationTranscriptionResponseStub{}
	provider.response = response

	result, err := app.Transcription(file).
		Provider("default").
		Model("gpt-4o-mini-transcribe").
		Language("en").
		Diarize().
		Timeout(3 * time.Second).
		Generate()

	require.NoError(t, err)
	assert.Equal(t, response, result)
	assert.Equal(t, ctx, provider.ctx)
	assert.Equal(t, contractsai.TranscriptionPrompt{
		File:     file,
		Model:    "gpt-4o-mini-transcribe",
		Language: "en",
		Diarize:  true,
		Timeout:  3 * time.Second,
	}, provider.prompt)
}

func TestAudioRequest_Generate(t *testing.T) {
	ctx := context.Background()
	provider := &applicationAudioProviderStub{}
	config := contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: provider},
		},
	}

	app := NewApplication(context.Background(), config)
	response := &applicationAudioResponseStub{}
	provider.response = response

	result, err := app.Audio("welcome to goravel").
		Provider("default").
		Model("gpt-4o-mini-tts").
		Male().
		Instructions("Speak slowly").
		Timeout(3 * time.Second).
		Generate()

	require.NoError(t, err)
	assert.Equal(t, response, result)
	assert.Equal(t, ctx, provider.ctx)
	assert.Equal(t, contractsai.AudioPrompt{
		Prompt:       "welcome to goravel",
		Model:        "gpt-4o-mini-tts",
		Voice:        DefaultMaleVoice,
		Instructions: "Speak slowly",
		Timeout:      3 * time.Second,
	}, provider.prompt)
}

func TestAudioRequest_StoreUsesResponseStore(t *testing.T) {
	provider := &applicationAudioProviderStub{}
	app := NewApplication(context.Background(), contractsai.Config{
		Default: "default",
		Providers: map[string]contractsai.ProviderConfig{
			"default": {Via: provider},
		},
	})
	response := &applicationAudioResponseStub{storePathResult: "audio/generated.mp3"}
	provider.response = response

	path, err := app.Audio("welcome to goravel").Store()

	require.NoError(t, err)
	assert.Equal(t, "audio/generated.mp3", path)
	assert.Equal(t, 1, response.storeCalls)
	assert.Equal(t, 0, response.storeAsCalls)
	assert.Empty(t, response.storePath)
}

func TestApplication_audio(t *testing.T) {
	tests := []struct {
		name         string
		options      []contractsai.Option
		prompt       contractsai.AudioPrompt
		setup        func() contractsai.Config
		expectError  error
		expectPrompt contractsai.AudioPrompt
	}{
		{
			name:    "success with default model",
			options: []contractsai.Option{WithProvider("openai")},
			prompt: contractsai.AudioPrompt{
				Prompt: "welcome to goravel",
				Voice:  DefaultFemaleVoice,
			},
			setup: func() contractsai.Config {
				provider := &applicationAudioProviderStub{}
				provider.response = &applicationAudioResponseStub{}
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: mocksai.NewProvider(t)},
						"openai":  {Via: provider},
					},
				}
			},
			expectPrompt: contractsai.AudioPrompt{
				Prompt: "welcome to goravel",
				Voice:  DefaultFemaleVoice,
			},
		},
		{
			name: "provider does not support audio",
			prompt: contractsai.AudioPrompt{
				Prompt: "welcome to goravel",
			},
			setup: func() contractsai.Config {
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: mocksai.NewProvider(t)},
					},
				}
			},
			expectError: errors.AIProviderDoesNotSupportAudio.Args("default"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApplication(context.Background(), tt.setup())
			response, err := app.audio(context.Background(), tt.prompt, tt.options...)
			assert.Equal(t, tt.expectError, err)
			if tt.expectError != nil {
				assert.Nil(t, response)
				return
			}

			require.NotNil(t, response)
			provider, ok := app.config.Providers["openai"].Via.(*applicationAudioProviderStub)
			require.True(t, ok)
			assert.Equal(t, tt.expectPrompt, provider.prompt)
		})
	}
}

func TestApplication_audio_Failover(t *testing.T) {
	failoverErr := NewFailoverError("primary", contractsai.FailoverReason("rate_limited"), assert.AnError)
	primaryProvider := &applicationAudioProviderStub{err: failoverErr}
	backupResponse := &applicationAudioResponseStub{}
	backupProvider := &applicationAudioProviderStub{response: backupResponse}
	app := NewApplication(context.Background(), contractsai.Config{
		Default: "primary",
		Providers: map[string]contractsai.ProviderConfig{
			"primary": {Via: primaryProvider},
			"backup":  {Via: backupProvider},
		},
	})

	response, err := app.audio(context.Background(), contractsai.AudioPrompt{Prompt: "welcome"}, WithProvider("primary", "backup"))

	require.NoError(t, err)
	assert.Equal(t, backupResponse, response)
	assert.Equal(t, "welcome", primaryProvider.prompt.Prompt)
	assert.Equal(t, "welcome", backupProvider.prompt.Prompt)
}

func TestApplication_transcription(t *testing.T) {
	file := mocksai.NewStorableFile(t)
	tests := []struct {
		name         string
		options      []contractsai.Option
		prompt       contractsai.TranscriptionPrompt
		setup        func() contractsai.Config
		expectError  error
		expectPrompt contractsai.TranscriptionPrompt
	}{
		{
			name:    "success with default model",
			options: []contractsai.Option{WithProvider("openai")},
			prompt: contractsai.TranscriptionPrompt{
				File: file,
			},
			setup: func() contractsai.Config {
				provider := &applicationTranscriptionProviderStub{}
				provider.response = &applicationTranscriptionResponseStub{}
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: mocksai.NewProvider(t)},
						"openai":  {Via: provider},
					},
				}
			},
			expectPrompt: contractsai.TranscriptionPrompt{File: file},
		},
		{
			name: "provider does not support transcription",
			prompt: contractsai.TranscriptionPrompt{
				File: file,
			},
			setup: func() contractsai.Config {
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: mocksai.NewProvider(t)},
					},
				}
			},
			expectError: errors.AIProviderDoesNotSupportTranscription.Args("default"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApplication(context.Background(), tt.setup())
			response, err := app.transcription(context.Background(), tt.prompt, tt.options...)
			assert.Equal(t, tt.expectError, err)
			if tt.expectError != nil {
				assert.Nil(t, response)
				return
			}

			require.NotNil(t, response)
			provider, ok := app.config.Providers["openai"].Via.(*applicationTranscriptionProviderStub)
			require.True(t, ok)
			assert.Equal(t, tt.expectPrompt, provider.prompt)
		})
	}
}

func TestApplication_image(t *testing.T) {
	tests := []struct {
		name        string
		options     []contractsai.Option
		setup       func() contractsai.Config
		expectError error
	}{
		{
			name:    "success",
			options: []contractsai.Option{WithProvider("openai"), WithModel("gpt-image-override")},
			setup: func() contractsai.Config {
				provider := &applicationImageProviderStub{}
				provider.response = &applicationImageResponseStub{}
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: mocksai.NewProvider(t)},
						"openai":  {Via: provider},
					},
				}
			},
		},
		{
			name: "provider does not support images",
			setup: func() contractsai.Config {
				return contractsai.Config{
					Default: "default",
					Providers: map[string]contractsai.ProviderConfig{
						"default": {Via: mocksai.NewProvider(t)},
					},
				}
			},
			expectError: errors.AIProviderDoesNotSupportImages.Args("default"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApplication(context.Background(), tt.setup())
			response, err := app.image(context.Background(), contractsai.ImagePrompt{Prompt: "draw a cat"}, tt.options...)
			assert.Equal(t, tt.expectError, err)
			if tt.expectError != nil {
				assert.Nil(t, response)
				return
			}

			require.NotNil(t, response)
			provider, ok := app.config.Providers["openai"].Via.(*applicationImageProviderStub)
			if ok {
				assert.Equal(t, "gpt-image-override", provider.prompt.Model)
			}
		})
	}
}

func TestApplication_getFile(t *testing.T) {
	ctx := context.WithValue(context.Background(), testCtxKey("get"), "success")
	fileProvider := mocksai.NewFileProvider(t)
	response := mocksai.NewFileResponse(t)
	response.EXPECT().ID().Return("file-123").Once()
	fileProvider.EXPECT().GetFile(ctx, "file-123").Return(response, nil).Once()

	app := NewApplication(context.Background(), contractsai.Config{
		Default: "openai",
		Providers: map[string]contractsai.ProviderConfig{
			"openai": {Via: uploadTestProvider{fileProvider: fileProvider}},
		},
	})

	file, err := app.getFile(ctx, "file-123")
	require.NoError(t, err)
	assert.Equal(t, "file-123", file.ID())
}

func TestApplication_getFileReturnsErrorWhenIDEmpty(t *testing.T) {
	app := NewApplication(context.Background(), contractsai.Config{})
	file, err := app.getFile(context.Background(), "")
	assert.Nil(t, file)
	assert.Equal(t, errors.AIStoredFileIDEmpty, err)
}

func TestApplication_deleteFile(t *testing.T) {
	ctx := context.WithValue(context.Background(), testCtxKey("delete"), "success")
	fileProvider := mocksai.NewFileProvider(t)
	fileProvider.EXPECT().DeleteFile(ctx, "file-123").Return(nil).Once()

	app := NewApplication(context.Background(), contractsai.Config{
		Default: "openai",
		Providers: map[string]contractsai.ProviderConfig{
			"openai": {Via: uploadTestProvider{fileProvider: fileProvider}},
		},
	})

	assert.NoError(t, app.deleteFile(ctx, "file-123"))
}

func TestApplication_deleteFileReturnsErrorWhenIDEmpty(t *testing.T) {
	app := NewApplication(context.Background(), contractsai.Config{})
	assert.Equal(t, errors.AIStoredFileIDEmpty, app.deleteFile(context.Background(), ""))
}

type applicationTestMiddleware struct{}

type uploadTestProvider struct {
	fileProvider contractsai.FileProvider
}

type applicationPromptProviderStub struct {
	name           string
	response       contractsai.AgentResponse
	err            error
	prompts        []contractsai.AgentPrompt
	previousState  any
	streamResponse contractsai.AgentResponse
	streamErr      error
	streamEvents   []contractsai.StreamEvent
	streamPrompts  []contractsai.AgentPrompt
}

func (p uploadTestProvider) Prompt(context.Context, contractsai.AgentPrompt) (contractsai.AgentResponse, error) {
	return nil, nil
}

func (p uploadTestProvider) Stream(context.Context, contractsai.AgentPrompt) (contractsai.StreamableAgentResponse, error) {
	return nil, nil
}

func (p uploadTestProvider) PutFile(ctx context.Context, file contractsai.StorableFile) (contractsai.FileResponse, error) {
	return p.fileProvider.PutFile(ctx, file)
}

func (p uploadTestProvider) GetFile(ctx context.Context, id string) (contractsai.FileResponse, error) {
	return p.fileProvider.GetFile(ctx, id)
}

func (p uploadTestProvider) DeleteFile(ctx context.Context, id string) error {
	return p.fileProvider.DeleteFile(ctx, id)
}

func (p *applicationPromptProviderStub) Prompt(_ context.Context, prompt contractsai.AgentPrompt) (contractsai.AgentResponse, error) {
	p.prompts = append(p.prompts, prompt)
	if prompt.ProviderState != nil {
		p.previousState = prompt.ProviderState.Get("response_id")
		prompt.ProviderState.Set("response_id", p.name)
	}

	return p.response, p.err
}

func (p *applicationPromptProviderStub) Stream(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableAgentResponse, error) {
	p.streamPrompts = append(p.streamPrompts, prompt)

	return NewStreamableResponse(ctx, func(_ context.Context, emit func(contractsai.StreamEvent) error) (contractsai.AgentResponse, error) {
		for _, event := range p.streamEvents {
			if err := emit(event); err != nil {
				return nil, err
			}
		}

		return p.streamResponse, p.streamErr
	}), nil
}

type applicationImageProviderStub struct {
	ctx      context.Context
	prompt   contractsai.ImagePrompt
	response contractsai.ImageResponse
	err      error
}

func (p *applicationImageProviderStub) Prompt(context.Context, contractsai.AgentPrompt) (contractsai.AgentResponse, error) {
	return nil, nil
}

func (p *applicationImageProviderStub) Stream(context.Context, contractsai.AgentPrompt) (contractsai.StreamableAgentResponse, error) {
	return nil, nil
}

func (p *applicationImageProviderStub) Image(ctx context.Context, prompt contractsai.ImagePrompt) (contractsai.ImageResponse, error) {
	p.ctx = ctx
	p.prompt = prompt
	return p.response, p.err
}

type applicationImageResponseStub struct {
	storePath       []string
	storePathResult string
	storeAsName     string
	storeAsPath     []string
	storeCalls      int
	storeAsCalls    int
}

func (r *applicationImageResponseStub) Content() ([]byte, error) {
	return []byte("image"), nil
}

func (r *applicationImageResponseStub) MimeType() string { return "image/png" }

func (r *applicationImageResponseStub) Store(disk ...string) (string, error) {
	r.storeCalls++
	r.storePath = append([]string(nil), disk...)
	if r.storePathResult != "" {
		return r.storePathResult, nil
	}

	content, err := r.Content()
	if err != nil {
		return "", err
	}

	resolvedDisk, err := resolveImageStoreDisk(disk)
	if err != nil {
		return "", err
	}

	return imageStorer{}.Store(content, "generated.png", resolvedDisk)
}

func (r *applicationImageResponseStub) StoreAs(path string, disk ...string) (string, error) {
	r.storeAsCalls++
	r.storeAsName = path
	r.storeAsPath = append([]string(nil), disk...)

	content, err := r.Content()
	if err != nil {
		return "", err
	}

	resolvedDisk, err := resolveImageStoreDisk(disk)
	if err != nil {
		return "", err
	}

	return imageStorer{}.StoreAs(content, path, resolvedDisk)
}

func (r *applicationImageResponseStub) Usage() contractsai.Usage { return nil }

func (r *applicationImageResponseStub) Then(callback func(contractsai.ImageResponse)) contractsai.ImageResponse {
	if callback != nil {
		callback(r)
	}

	return r
}

type applicationAudioProviderStub struct {
	ctx      context.Context
	prompt   contractsai.AudioPrompt
	response contractsai.AudioResponse
	err      error
}

type applicationTranscriptionProviderStub struct {
	ctx      context.Context
	prompt   contractsai.TranscriptionPrompt
	response contractsai.TranscriptionResponse
	err      error
}

func (p *applicationAudioProviderStub) Prompt(context.Context, contractsai.AgentPrompt) (contractsai.AgentResponse, error) {
	return nil, nil
}

func (p *applicationAudioProviderStub) Stream(context.Context, contractsai.AgentPrompt) (contractsai.StreamableAgentResponse, error) {
	return nil, nil
}

func (p *applicationAudioProviderStub) Audio(ctx context.Context, prompt contractsai.AudioPrompt) (contractsai.AudioResponse, error) {
	p.ctx = ctx
	p.prompt = prompt
	return p.response, p.err
}

func (p *applicationTranscriptionProviderStub) Prompt(context.Context, contractsai.AgentPrompt) (contractsai.AgentResponse, error) {
	return nil, nil
}

func (p *applicationTranscriptionProviderStub) Stream(context.Context, contractsai.AgentPrompt) (contractsai.StreamableAgentResponse, error) {
	return nil, nil
}

func (p *applicationTranscriptionProviderStub) Transcription(ctx context.Context, prompt contractsai.TranscriptionPrompt) (contractsai.TranscriptionResponse, error) {
	p.ctx = ctx
	p.prompt = prompt
	return p.response, p.err
}

type applicationAudioResponseStub struct {
	storePath       []string
	storePathResult string
	storeAsName     string
	storeAsPath     []string
	storeCalls      int
	storeAsCalls    int
}

func (r *applicationAudioResponseStub) Content() ([]byte, error) {
	return []byte("audio"), nil
}

func (r *applicationAudioResponseStub) MimeType() string { return "audio/mpeg" }

func (r *applicationAudioResponseStub) Store(disk ...string) (string, error) {
	r.storeCalls++
	r.storePath = append([]string(nil), disk...)
	if r.storePathResult != "" {
		return r.storePathResult, nil
	}

	content, err := r.Content()
	if err != nil {
		return "", err
	}

	resolvedDisk, err := resolveAudioStoreDisk(disk)
	if err != nil {
		return "", err
	}

	return audioStorer{}.Store(content, "generated.mp3", resolvedDisk)
}

func (r *applicationAudioResponseStub) StoreAs(path string, disk ...string) (string, error) {
	r.storeAsCalls++
	r.storeAsName = path
	r.storeAsPath = append([]string(nil), disk...)

	content, err := r.Content()
	if err != nil {
		return "", err
	}

	resolvedDisk, err := resolveAudioStoreDisk(disk)
	if err != nil {
		return "", err
	}

	return audioStorer{}.StoreAs(content, path, resolvedDisk)
}

func (r *applicationAudioResponseStub) Usage() contractsai.Usage { return nil }

func (r *applicationAudioResponseStub) Then(callback func(contractsai.AudioResponse)) contractsai.AudioResponse {
	if callback != nil {
		callback(r)
	}

	return r
}

type applicationTranscriptionResponseStub struct {
	text     string
	segments []contractsai.TranscriptionSegment
	usage    contractsai.Usage
}

func (r *applicationTranscriptionResponseStub) Text() string { return r.text }

func (r *applicationTranscriptionResponseStub) Segments() []contractsai.TranscriptionSegment {
	return append([]contractsai.TranscriptionSegment(nil), r.segments...)
}

func (r *applicationTranscriptionResponseStub) Usage() contractsai.Usage { return r.usage }

func (r *applicationTranscriptionResponseStub) Then(callback func(contractsai.TranscriptionResponse)) contractsai.TranscriptionResponse {
	if callback != nil {
		callback(r)
	}

	return r
}

func (m *applicationTestMiddleware) Handle(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.AgentResponse, error) {
	response, err := next(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return response.Then(func(response contractsai.AgentResponse) {
		if stub, ok := response.(*middlewareResponse); ok {
			stub.response = &stubResponse{text: response.Text() + " after middleware"}
		}
	}), nil
}
