package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/goravel/framework/cache/support"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/support/facades"
	"runtime/debug"
)

type Application struct {
}

func (app *Application) Init() cache.Store {
	var store cache.Store

	defaultStore := facades.Config.GetString("cache.default")
	driver := facades.Config.GetString("cache.stores." + defaultStore + ".driver")
	if driver == "redis" {
		return app.createRedisDriver()
	}
	if driver == "custom" {
		return facades.Config.Get("cache.stores." + defaultStore + ".via").(support.Store).Handle()
	}

	facades.Log.Warnf("Not supported cache store:" + defaultStore)

	return store
}

func (app *Application) createRedisDriver() *Redis {
	connection := facades.Config.GetString("cache.stores." + facades.Config.GetString("cache.default") + ".connection")
	if connection == "" {
		connection = "default"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     facades.Config.GetString("database.redis."+connection+".host") + ":" + facades.Config.GetString("database.redis."+connection+".port"),
		Password: facades.Config.GetString("database.redis." + connection + ".password"),
		DB:       facades.Config.GetInt("database.redis." + connection + ".database"),
	})

	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		facades.Log.Warnf("Failed to link redis:%s, %s\n%+v", pong, err, string(debug.Stack()))
	}

	return &Redis{
		Redis:  client,
		Prefix: facades.Config.GetString("cache.prefix" + ":"),
	}
}
