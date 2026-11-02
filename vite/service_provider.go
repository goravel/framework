package vite

import (
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/vite/console"
)

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Vite,
		},
		Dependencies: binding.Bindings[binding.Vite].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Vite, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleVite)
		}

		return NewVite(config), nil
	})
}

// Boot registers the vite template functions and the vite:inputs command. The
// view set is compiled only after every provider has booted, so the functions
// are in place before the first template is parsed.
func (r *ServiceProvider) Boot(app foundation.Application) {
	view := app.MakeView()
	if view == nil {
		return
	}

	vite := app.MakeVite()
	if vite != nil {
		view.AddFuncs(Funcs(vite))
	}

	app.Commands([]contractsconsole.Command{
		console.NewInputsCommand(view, vite),
	})
}
