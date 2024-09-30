package crypt

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.crypt"

type ServiceProvider struct {
}

func (crypt *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		c := app.MakeConfig()
		if c == nil {
			return nil, ErrConfigNotSet
		}

		j := app.GetJson()
		if j == nil {
			return nil, ErrJsonParserNotSet
		}

		return NewAES(c, j)
	})
}

func (crypt *ServiceProvider) Boot(app foundation.Application) {

}
