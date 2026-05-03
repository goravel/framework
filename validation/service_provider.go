package validation

import (
	"github.com/goravel/framework/contracts/binding"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/validation/console"
)

var ormFacade orm.Orm

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Validation,
		},
		Dependencies: binding.Bindings[binding.Validation].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Validation, func(app foundation.Application) (any, error) {
		return NewValidation(), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	ormFacade = app.MakeOrm()

	app.Commands([]consolecontract.Command{
		&console.RuleMakeCommand{},
		&console.FilterMakeCommand{},
	})
}
