package route

import (
	"fmt"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/route"
)

type Driver string

const (
	DriverGin   Driver = "gin"
	DriverFiber Driver = "fiber"
)

type Route struct {
	route.Engine
	config config.Config
}

func NewRoute(config config.Config) *Route {
	defaultDriver := config.GetString("http.driver")
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
	switch Driver(driver) {
	case DriverGin:
		driver, ok := config.Get("http.drivers.gin.via").(route.Engine)
		if ok {
			return driver, nil
		}

		driverCallback, ok := config.Get("http.drivers.gin.via").(func() (route.Engine, error))
		if ok {
			return driverCallback()
		}

		return nil, fmt.Errorf("init gin route driver fail: via must be implement route.Route or func() (route.Route, error)")
	case DriverFiber:
		driver, ok := config.Get("http.drivers.fiber.via").(route.Engine)
		if ok {
			return driver, nil
		}

		driverCallback, ok := config.Get("http.drivers.fiber.via").(func() (route.Engine, error))
		if ok {
			return driverCallback()
		}

		return nil, fmt.Errorf("init fiber route driver fail: via must be implement route.Route or func() (route.Route, error)")
	}

	return nil, fmt.Errorf("invalid driver: %s, only support gin, fiber", driver)
}
