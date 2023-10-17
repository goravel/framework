package testing

import (
	"github.com/goravel/framework/contracts/foundation"
)

type Application struct {
	app foundation.Application
}

func NewApplication(app foundation.Application) *Application {
	return &Application{
		app: app,
	}
}
