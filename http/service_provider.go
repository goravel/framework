package http

import (
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/cache"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http/console"
)

type ServiceProvider struct{}

var (
	CacheFacade       cache.Cache
	RateLimiterFacade http.RateLimiter
	JsonFacade        foundation.Json
)

func (http *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingRateLimiter, func(app foundation.Application) (any, error) {
		return NewRateLimiter(), nil
	})
	app.Singleton(contracts.BindingView, func(app foundation.Application) (any, error) {
		return NewView(), nil
	})
}

func (http *ServiceProvider) Boot(app foundation.Application) {
	CacheFacade = app.MakeCache()
	RateLimiterFacade = app.MakeRateLimiter()
	JsonFacade = app.GetJson()

	http.registerCommands(app)
}

func (http *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]consolecontract.Command{
		&console.RequestMakeCommand{},
		&console.ControllerMakeCommand{},
		&console.MiddlewareMakeCommand{},
	})
}
