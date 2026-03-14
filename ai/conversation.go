package ai

import (
	"context"
	"sync"
	"time"

	contractsai "github.com/goravel/framework/contracts/ai"
)

type conversation struct {
	ctx          context.Context
	provider     contractsai.Provider
	agent        contractsai.Agent
	options      []contractsai.Option
	timeout      time.Duration
	baseMessages []contractsai.Message
	messages     []contractsai.Message
	mu           sync.RWMutex
}

func (r *conversation) Prompt(ctx context.Context, input string) (contractsai.Response, error) {
	return nil, nil
}

func (r *conversation) Messages() []contractsai.Message {
	return nil
}

func (r *conversation) Reset() {
}
