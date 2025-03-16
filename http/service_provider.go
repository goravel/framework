package http

import (
	"time"

	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/cache"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	contractsclient "github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/http/client"
	"github.com/goravel/framework/http/console"
)

type ServiceProvider struct{}

var (
	CacheFacade       cache.Cache
	LogFacade         log.Log
	RateLimiterFacade http.RateLimiter
	JsonFacade        foundation.Json
)

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingRateLimiter, func(app foundation.Application) (any, error) {
		return NewRateLimiter(), nil
	})
	app.Singleton(contracts.BindingView, func(app foundation.Application) (any, error) {
		return NewView(), nil
	})
	app.Bind(contracts.BindingHttp, func(app foundation.Application) (any, error) {
		c := app.MakeConfig()
		if c == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleHttp)
		}

		j := app.GetJson()
		if j == nil {
			return nil, errors.JSONParserNotSet.SetModule(errors.ModuleHttp)
		}

		config := &contractsclient.Config{
			Timeout:             c.GetDuration("http.client.timeout", 30*time.Second),
			BaseUrl:             c.GetString("http.client.base_url"),
			MaxIdleConns:        c.GetInt("http.client.max_idle_conns"),
			MaxIdleConnsPerHost: c.GetInt("http.client.max_idle_conns_per_host"),
			MaxConnsPerHost:     c.GetInt("http.client.max_conns_per_host"),
			IdleConnTimeout:     c.GetDuration("http.client.idle_conn_timeout"),
		}
		return client.NewRequest(config, j), nil
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

func (http *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]contractsconsole.Command{
		&console.RequestMakeCommand{},
		&console.ControllerMakeCommand{},
		&console.MiddlewareMakeCommand{},
	})
}
