package cache

import (
	"context"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/facades"
)

type Application struct {
}

func (app *Application) Init() cache.Store {
	defaultStore := facades.Config.GetString("cache.default")
	driver := facades.Config.GetString("cache.stores." + defaultStore + ".driver")
	if driver == "redis" {
		redis, err := NewRedis(context.Background())
		if err != nil {
			color.Redf("[Cache] Init redis driver error: %v\n", err)
			return nil
		}

		return redis
	}

	if driver == "memory" {
		memory, err := NewMemory()
		if err != nil {
			color.Redf("[Cache] Init memory driver error: %v\n", err)
			return nil
		}

		return memory
	}

	if driver == "custom" {
		if custom, ok := facades.Config.Get("cache.stores." + defaultStore + ".via").(cache.Store); ok {
			return custom
		}
		color.Redf("[Cache] %s doesn't implement contracts/cache/store\n", defaultStore)

		return nil
	}

	color.Redf("[Cache] Not supported cache store: %s\n", defaultStore)

	return nil
}
