package ai

import (
	"context"

	contractsai "github.com/goravel/framework/contracts/ai"
)

func NextFunc(run func(context.Context, contractsai.AgentPrompt) (contractsai.Response, error)) contractsai.Next {
	return contractsai.Next(run)
}
