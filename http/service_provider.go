package http

import (
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/http/console"
)

const Binding = "goravel.http"

var (
	ConfigFacade      config.Config
	CacheFacade       cache.Cache
	LogFacade         log.Log
	RateLimiterFacade http.RateLimiter
	ValidationFacade  validation.Validation
)

type ServiceProvider struct {
}

func (http *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewRateLimiter(), nil
	})
}

func (http *ServiceProvider) Boot(app foundation.Application) {
	ConfigFacade = app.MakeConfig()
	CacheFacade = app.MakeCache()
	LogFacade = app.MakeLog()
	RateLimiterFacade = app.MakeRateLimiter()
	ValidationFacade = app.MakeValidation()

	http.registerCommands(app)
}

func (http *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]consolecontract.Command{
		&console.RequestMakeCommand{},
		&console.ControllerMakeCommand{},
		&console.MiddlewareMakeCommand{},
	})
}
