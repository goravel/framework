package testing

import (
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/color"
)

const Binding = "goravel.testing"

var artisanFacade contractsconsole.Artisan

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewApplication(app), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	artisanFacade = app.MakeArtisan()
	if artisanFacade == nil {
		color.Red().Println("Warning: Artisan facade is not initialized")
	}
}
