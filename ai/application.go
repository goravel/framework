package ai

import (
	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/contracts/config"
)

var _ contractsai.AI = (*Application)(nil)

// Application is the AI manager implementation.
type Application struct {
}

func NewApplication(config config.Config) *Application {
	return &Application{}
}

func (r *Application) Agent(agent contractsai.Agent, options ...contractsai.Option) (contractsai.Conversation, error) {
	return &conversation{}, nil
}
