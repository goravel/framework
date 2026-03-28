package ai

import contractsai "github.com/goravel/framework/contracts/ai"

func WithProvider(provider string) contractsai.Option {
	return func(options map[string]any) {
		options[contractsai.OptionProvider] = provider
	}
}

func WithModel(model string) contractsai.Option {
	return func(options map[string]any) {
		options[contractsai.OptionModel] = model
	}
}
