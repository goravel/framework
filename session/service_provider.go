package session

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.session"
const BindingStore = "goravel.session.store"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(BindingStore, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		return NewManager(config), nil
	})

	//app.Singleton(BindingStore, func(app foundation.Application) (any, error) {
	//	driver, err := app.Make(Binding)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return driver.(sessioncontract.Manager).Driver(), nil
	//})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
}
