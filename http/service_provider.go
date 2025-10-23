package http

import (
	"time"

	contractsbinding "github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	contractsclient "github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/http/client"
	"github.com/goravel/framework/http/console"
	"github.com/goravel/framework/support/binding"
)

type ServiceProvider struct{}

var (
	CacheFacade       cache.Cache
	ConfigFacade      config.Config
	LogFacade         log.Log
	RateLimiterFacade http.RateLimiter
	JsonFacade        foundation.Json
)

func (r *ServiceProvider) Relationship() contractsbinding.Relationship {
	bindings := []string{
		contractsbinding.Http,
		contractsbinding.RateLimiter,
		contractsbinding.View,
	}

	return contractsbinding.Relationship{
		Bindings:     bindings,
		Dependencies: binding.Dependencies(bindings...),
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contractsbinding.RateLimiter, func(app foundation.Application) (any, error) {
		return NewRateLimiter(), nil
	})
	app.Bind(contractsbinding.Http, func(app foundation.Application) (any, error) {
		ConfigFacade = app.MakeConfig()
		if ConfigFacade == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleHttp)
		}

		j := app.GetJson()
		if j == nil {
			return nil, errors.JSONParserNotSet.SetModule(errors.ModuleHttp)
		}

		return client.NewRequest(&contractsclient.Config{
			Timeout:             time.Duration(ConfigFacade.GetInt("http.client.timeout", 30)) * time.Second,
			BaseUrl:             ConfigFacade.GetString("http.client.base_url"),
			MaxIdleConns:        ConfigFacade.GetInt("http.client.max_idle_conns"),
			MaxIdleConnsPerHost: ConfigFacade.GetInt("http.client.max_idle_conns_per_host"),
			MaxConnsPerHost:     ConfigFacade.GetInt("http.client.max_conns_per_host"),
			IdleConnTimeout:     ConfigFacade.GetDuration("http.client.idle_conn_timeout"),
		}, j), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	CacheFacade = app.MakeCache()
	if CacheFacade == nil {
		panic(errors.CacheFacadeNotSet.SetModule(errors.ModuleHttp))
	}

	LogFacade = app.MakeLog()
	if LogFacade == nil {
		panic(errors.LogFacadeNotSet.SetModule(errors.ModuleHttp))
	}

	RateLimiterFacade = app.MakeRateLimiter()
	if RateLimiterFacade == nil {
		panic(errors.RateLimiterFacadeNotSet.SetModule(errors.ModuleHttp))
	}

	JsonFacade = app.GetJson()
	if JsonFacade == nil {
		panic(errors.JSONParserNotSet.SetModule(errors.ModuleHttp))
	}

	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]contractsconsole.Command{
		&console.RequestMakeCommand{},
		&console.ControllerMakeCommand{},
		&console.MiddlewareMakeCommand{},
	})
}
