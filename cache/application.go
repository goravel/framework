package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
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
		return app.createRedisDriver()
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

func (app *Application) createRedisDriver() *Redis {
	connection := facades.Config.GetString("cache.stores." + facades.Config.GetString("cache.default") + ".connection")
	if connection == "" {
		connection = "default"
	}

	host := facades.Config.GetString("database.redis." + connection + ".host")
	if host == "" {
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + facades.Config.GetString("database.redis."+connection+".port"),
		Password: facades.Config.GetString("database.redis." + connection + ".password"),
		DB:       facades.Config.GetInt("database.redis." + connection + ".database"),
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		color.Redln(fmt.Sprintf("[Cache] Init connection error, %s", err.Error()))

		return nil
	}

	return &Redis{
		Redis:  client,
		Prefix: facades.Config.GetString("cache.prefix" + ":"),
	}
}
