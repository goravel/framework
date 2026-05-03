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

			var response *mocksai.Response
			if tt.expectResponse {
				response = mocksai.NewResponse(t)
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
				response := mocksai.NewStoredFileResponse(t)
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
	request := app.Image("draw a cat", WithProvider("default"), WithModel("gpt-image-1"))

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
		Timeout:     int64(3 * time.Second),
	}, provider.prompt)
}

func TestApplication_image(t *testing.T) {
	tests := []struct {
		name        string
		options     []contractsai.Option
		setup       func() contractsai.Config
		expectError error
	}{
		{
			name: "success",
			options: []contractsai.Option{WithProvider("openai")},
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
		})
	}
}

type applicationTestMiddleware struct{}

type uploadTestProvider struct {
	fileProvider contractsai.FileProvider
}

func (p uploadTestProvider) Prompt(context.Context, contractsai.AgentPrompt) (contractsai.Response, error) {
	return nil, nil
}

func (p uploadTestProvider) Stream(context.Context, contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
	return nil, nil
}

func (p uploadTestProvider) PutFile(ctx context.Context, file contractsai.StorableFile) (contractsai.StoredFileResponse, error) {
	return p.fileProvider.PutFile(ctx, file)
}

type applicationImageProviderStub struct {
	ctx      context.Context
	prompt   contractsai.ImagePrompt
	response contractsai.ImageResponse
	err      error
}

func (p *applicationImageProviderStub) Prompt(context.Context, contractsai.AgentPrompt) (contractsai.Response, error) {
	return nil, nil
}

func (p *applicationImageProviderStub) Stream(context.Context, contractsai.AgentPrompt) (contractsai.StreamableResponse, error) {
	return nil, nil
}

func (p *applicationImageProviderStub) Image(ctx context.Context, prompt contractsai.ImagePrompt) (contractsai.ImageResponse, error) {
	p.ctx = ctx
	p.prompt = prompt
	return p.response, p.err
}

type applicationImageResponseStub struct{}

func (r *applicationImageResponseStub) Content(context.Context) ([]byte, error) { return []byte("image"), nil }

func (r *applicationImageResponseStub) MimeType() string { return "image/png" }

func (r *applicationImageResponseStub) Usage() contractsai.Usage { return nil }

func (r *applicationImageResponseStub) Then(callback func(contractsai.ImageResponse)) contractsai.ImageResponse {
	if callback != nil {
		callback(r)
	}

	return r
}

func (m *applicationTestMiddleware) Handle(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
	response, err := next(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return response.Then(func(response contractsai.Response) {
		if stub, ok := response.(*middlewareResponse); ok {
			stub.response = &stubResponse{text: response.Text() + " after middleware"}
		}
	}), nil
}
