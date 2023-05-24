package cache

import (
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
)

type Application struct {
	cache.Driver
	config config.Config
	driver Driver
	stores map[string]cache.Driver
}

func NewApplication(config config.Config, store string) *Application {
	driver := NewDriverImpl(config)
	instance := driver.New(store)
	if instance == nil {
		return nil
	}

	return &Application{
		Driver: instance,
		config: config,
		driver: driver,
		stores: map[string]cache.Driver{
			store: instance,
		},
	}
}

func (app *Application) Store(name string) cache.Driver {
	if driver, exist := app.stores[name]; exist {
		return driver
	}

	instance := app.driver.New(name)
	app.stores[name] = instance

	return instance
}
