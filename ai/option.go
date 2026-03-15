package ai

import (
	"time"

	contractsai "github.com/goravel/framework/contracts/ai"
)

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

func WithTimeout(timeout time.Duration) contractsai.Option {
	return func(options map[string]any) {
		options[contractsai.OptionTimeout] = timeout
	}
}
