package ai

import (
	"time"

	contractsai "github.com/goravel/framework/contracts/ai"
)

func WithTimeout(timeout time.Duration) contractsai.Option {
	return func(options map[string]any) {
		options[contractsai.OptionTimeout] = timeout
	}
}
