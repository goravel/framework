package ai

import (
	"context"

	contractsai "github.com/goravel/framework/contracts/ai"
)

type conversation struct {
}

func (r *conversation) Prompt(ctx context.Context, input string) (contractsai.Response, error) {
	return nil, nil
}

func (r *conversation) Messages() []contractsai.Message {
	return nil
}

func (r *conversation) Reset() {
}
