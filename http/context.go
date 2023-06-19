package http

import (
	"fmt"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/http"
)

type Driver string

const (
	DriverGin   Driver = "gin"
	DriverFiber Driver = "fiber"
)

type Context struct {
	http.Context
	config config.Config
}

func Background() http.Context {
	c := NewContext(ConfigFacade)
	return c.Context
}

func NewContext(config config.Config) *Context {
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

	return &Context{
		Context: driver,
		config:  config,
	}
}

func NewDriver(config config.Config, driver string) (http.Context, error) {
	switch Driver(driver) {
	case DriverGin:
		driver, ok := config.Get("http.drivers.gin.http").(http.Context)
		if ok {
			return driver, nil
		}

		driverCallback, ok := config.Get("http.drivers.gin.http").(func() (http.Context, error))
		if ok {
			return driverCallback()
		}

		return nil, fmt.Errorf("init gin http driver fail: http must be implement http.Context or func() (http.Context, error)")
	case DriverFiber:
		driver, ok := config.Get("http.drivers.fiber.http").(http.Context)
		if ok {
			return driver, nil
		}

		driverCallback, ok := config.Get("http.drivers.fiber.http").(func() (http.Context, error))
		if ok {
			return driverCallback()
		}

		return nil, fmt.Errorf("init fiber http driver fail: http must be implement http.Context or func() (http.Context, error)")
	}

	return nil, fmt.Errorf("invalid driver: %s, only support gin, fiber", driver)
}
