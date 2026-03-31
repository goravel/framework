package ai

import contractsai "github.com/goravel/framework/contracts/ai"

func WithProvider(provider string) contractsai.Option {
	return func(options *contractsai.Options) {
		options.Provider = provider
	}
}

func WithModel(model string) contractsai.Option {
	return func(options *contractsai.Options) {
		options.Model = model
	}
}

func WithStreamCode(code int) contractsai.StreamOption {
	return func(options *contractsai.StreamOptions) {
		options.Code = code
	}
}

func WithStreamRender(render contractsai.RenderFunc) contractsai.StreamOption {
	return func(options *contractsai.StreamOptions) {
		options.Render = render
	}
}
