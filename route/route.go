package route

import (
	"fmt"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/route"
)

type Driver string

type Route struct {
	route.Engine
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
		Engine: driver,
		config: config,
	}
}

func NewDriver(config config.Config, driver string) (route.Engine, error) {
	engine, ok := config.Get("http.drivers." + driver + ".route").(route.Engine)
	if ok {
		return engine, nil
	}

	engineCallback, ok := config.Get("http.drivers." + driver + ".route").(func() (route.Engine, error))
	if ok {
		return engineCallback()
	}

	return nil, fmt.Errorf("init route driver fail: route must be implement route.Engine or func() (route.Engine, error)")
}
