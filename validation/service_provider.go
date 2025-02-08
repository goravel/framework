package validation

import (
	"github.com/goravel/framework/contracts"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/validation/console"
)

type ServiceProvider struct {
}

func (validation *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingValidation, func(app foundation.Application) (any, error) {
		return NewValidation(), nil
	})
}

func (validation *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]consolecontract.Command{
		&console.RuleMakeCommand{},
		&console.FilterMakeCommand{},
	})
}
