package route

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/errors"
)

type Driver string

type Route struct {
	route.Route
	config config.Config
}

func NewRoute(config config.Config) (*Route, error) {
	defaultDriver := config.GetString("http.default")
	if defaultDriver == "" {
		return nil, errors.RouteDefaultDriverNotSet.SetModule(errors.ModuleRoute)
	}

	driver, err := NewDriver(config, defaultDriver)
	if err != nil {
		return nil, err
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
