package ai

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
)

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.AI,
		},
		// Dependencies: binding.Bindings[binding.AI].Dependencies,
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.AI, func(app foundation.Application) (any, error) {
		return NewApplication(app.MakeConfig()), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {}
