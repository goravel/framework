package route

import (
	"fmt"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/route"
)

type Driver string

type Route struct {
	route.Route
	config config.Config
}

func NewRoute(config config.Config) *Route {
	defaultDriver := config.GetString("http.default")
	if defaultDriver == "" {
		color.Redln("[http] please set default driver")

		return nil
	}

	driver, err := NewDriver(config, defaultDriver)
	if err != nil {
		color.Redf("[http] %s\n", err)

		return nil
	}

	return &Route{
		Route:  driver,
		config: config,
	}
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

	return nil, fmt.Errorf("init route driver fail: route must be implement route.Route or func() (route.Route, error)")
}
