package http

import (
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/cache"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/http/client"
	"github.com/goravel/framework/http/console"
)

type ServiceProvider struct{}

var (
	CacheFacade       cache.Cache
	RateLimiterFacade http.RateLimiter
)

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingRateLimiter, func(app foundation.Application) (any, error) {
		return NewRateLimiter(), nil
	})
	app.Singleton(contracts.BindingView, func(app foundation.Application) (any, error) {
		return NewView(), nil
	})
	app.Singleton(contracts.BindingHttp, func(app foundation.Application) (any, error) {
		c := app.MakeConfig()
		if c == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleHttp)
		}

		j := app.GetJson()
		if j == nil {
			return nil, errors.JSONParserNotSet.SetModule(errors.ModuleHttp)
		}

		return client.NewRequest(c, j), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	CacheFacade = app.MakeCache()
	RateLimiterFacade = app.MakeRateLimiter()

	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]consolecontract.Command{
		&console.RequestMakeCommand{},
		&console.ControllerMakeCommand{},
		&console.MiddlewareMakeCommand{},
	})
}
