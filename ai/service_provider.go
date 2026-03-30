package ai

import (
	"context"

	"github.com/goravel/framework/ai/console"
	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
)

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.AI,
		},
		Dependencies: binding.Bindings[binding.AI].Dependencies,
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.AI, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		var aiConfig contractsai.Config
		if err := config.UnmarshalKey("ai", &aiConfig); err != nil {
			return nil, err
		}

		return NewApplication(context.Background(), aiConfig), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]contractsconsole.Command{
		&console.AgentMakeCommand{},
	})
}
