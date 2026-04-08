package validation

import (
	"context"

	"github.com/goravel/framework/contracts/binding"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/foundation"
	contractstranslation "github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/validation/console"
)

var (
	ormFacade  orm.Orm
	langFacade = func(ctx context.Context) contractstranslation.Translator { return nil }
)

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
	langFacade = app.MakeLang

	app.Publishes("github.com/goravel/framework/validation", map[string]string{
		"lang": app.LangPath(),
	}, "goravel-validation-lang")

	app.Commands([]consolecontract.Command{
		&console.RuleMakeCommand{},
		&console.FilterMakeCommand{},
	})
}
