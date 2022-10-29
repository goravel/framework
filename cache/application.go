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
		return NewRedis(context.Background())
	}
	if driver == "custom" {
		if custom, ok := facades.Config.Get("cache.stores." + defaultStore + ".via").(cache.Store); ok {
			return custom
		}
		color.Redln("%s doesn't implement contracts/cache/store", defaultStore)

		return nil
	}

	color.Redln("Not supported cache store: %s", defaultStore)

	return nil
}
