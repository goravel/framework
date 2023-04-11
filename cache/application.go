package cache

import (
	"context"
	"fmt"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/facades"
)

type Application struct {
	cache.Driver
	stores map[string]cache.Driver
}

func NewApplication(store string) *Application {
	driver := driver(store)
	if driver == nil {
		return nil
	}

	return &Application{
		Driver: driver,
		stores: map[string]cache.Driver{
			store: driver,
		},
	}
}

func (app *Application) Store(name string) cache.Driver {
	if driver, exist := app.stores[name]; exist {
		return driver
	}

	return driver(name)
}

func driver(store string) cache.Driver {
	driver := facades.Config.GetString(fmt.Sprintf("cache.stores.%s.driver", store))
	switch driver {
	case "redis":
		return initRedis(store)
	case "memory":
		return initMemory()
	case "custom":
		return initCustom(store)
	default:
		color.Redf("[Cache] Not supported cache store: %s\n", store)
		return nil
	}
}

func initRedis(store string) cache.Driver {
	redis, err := NewRedis(context.Background(), facades.Config.GetString(fmt.Sprintf("cache.stores.%s.connection", store), "default"))
	if err != nil {
		color.Redf("[Cache] Init redis driver error: %v\n", err)
		return nil
	}
	if redis == nil {
		return nil
	}

	return redis
}

func initMemory() cache.Driver {
	memory, err := NewMemory()
	if err != nil {
		color.Redf("[Cache] Init memory driver error: %v\n", err)
		return nil
	}

	return memory
}

func initCustom(store string) cache.Driver {
	if custom, ok := facades.Config.Get(fmt.Sprintf("cache.stores.%s.via", store)).(cache.Driver); ok {
		return custom
	}
	color.Redf("[Cache] %s doesn't implement contracts/cache/store\n", store)

	return nil
}
