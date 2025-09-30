package route

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type Driver string

type Route struct {
	route.Route
	config config.Config
}

func NewRoute(config config.Config) (*Route, error) {
	var driver route.Route

	defaultDriver := config.GetString("http.default")
	if defaultDriver == "" {
		color.Warningln(errors.RouteDefaultDriverNotSet.SetModule(errors.ModuleRoute).Error())
	} else {
		var err error
		driver, err = NewDriver(config, defaultDriver)
		if err != nil {
			return nil, err
		}
	}

	return &Route{
		Route:  driver,
		config: config,
	}, nil
}

func NewDriver(config config.Config, driver string) (route.Route, error) {
	engine, ok := config.Get("http.drivers." + driver + ".route").(route.Route)
	if ok {
		return engine, nil
	}

	engineCallback, ok := config.Get("http.drivers." + driver + ".route").(func() (route.Route, error))
	if ok {
		return engineCallback()
	}

	return nil, errors.RouteInvalidDriver.Args(driver).SetModule(errors.ModuleRoute)
}

// GlobalMiddleware registers global middleware for the route.
// It's an empty function to avoid panic when installing a http driver and the http.default configuration is empty,
// Given it is called in the AppServiceProvider boot method.
func (r *Route) GlobalMiddleware(middlewares ...http.Middleware) {}
