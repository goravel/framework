package validation

import (
	"github.com/goravel/framework/config"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/validation/console"
)

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(config.BindingValidation, func(app foundation.Application) (any, error) {
		return NewValidation(), nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]consolecontract.Command{
		&console.RuleMakeCommand{},
		&console.FilterMakeCommand{},
	})
}
