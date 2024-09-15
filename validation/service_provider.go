package validation

import (
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/validation/console"
)

const Binding = "goravel.validation"

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewValidation(), nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {
	if artisanFacade := app.MakeArtisan(); artisanFacade != nil {
		artisanFacade.Register([]consolecontract.Command{
			&console.RuleMakeCommand{},
			&console.FilterMakeCommand{},
		})
	}
}
