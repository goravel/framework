package vite

import (
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
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
// A missing dependency is reported, not swallowed: otherwise the first symptom is
// `function "vite" not defined` at parse time with nothing pointing at the cause.
func (r *ServiceProvider) Boot(app foundation.Application) {
	view := app.MakeView()
	if view == nil {
		color.Warningln(errors.ViewFacadeNotSet.SetModule(errors.ModuleVite))
		return
	}

	vite := app.MakeVite()
	if vite == nil {
		color.Warningln(errors.ViteFacadeNotSet)
	} else {
		view.AddFuncs(Funcs(vite))
	}

	app.Commands([]contractsconsole.Command{
		console.NewInputsCommand(view, vite),
	})
}
