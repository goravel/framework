package http

import (
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http/console"
)

const BindingRateLimiter = "goravel.rate_limiter"
const BindingView = "goravel.view"

type ServiceProvider struct{}

var (
	CacheFacade       cache.Cache
	ConfigFacade      config.Config
	RateLimiterFacade http.RateLimiter
)

func (http *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(BindingRateLimiter, func(app foundation.Application) (any, error) {
		return NewRateLimiter(), nil
	})
	app.Singleton(BindingView, func(app foundation.Application) (any, error) {
		return NewView(), nil
	})
}

func (http *ServiceProvider) Boot(app foundation.Application) {
	CacheFacade = app.MakeCache()
	ConfigFacade = app.MakeConfig()
	RateLimiterFacade = app.MakeRateLimiter()

	http.registerCommands(app)
}

func (http *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]consolecontract.Command{
		&console.RequestMakeCommand{},
		&console.ControllerMakeCommand{},
		&console.MiddlewareMakeCommand{},
	})
}
