package limiter

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"

	"github.com/goravel/framework/facades"
)

// CheckRate check rate limit
func CheckRate(c *gin.Context, key string, formatted string) (limiter.Context, error) {

	var context limiter.Context
	rate, err := limiter.NewRateFromFormatted(formatted)
	if err != nil {
		return context, err
	}

	// init store, default is memory
	store := memory.NewStore()
	storeDevice := facades.Config.GetString("limiter.store", "memory")

	if storeDevice == "redis" {
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", facades.Config.GetString("database.redis.default.host"), facades.Config.GetString("database.redis.default.port")),
			Password: facades.Config.GetString("database.redis.default.password"),
			DB:       facades.Config.GetInt("database.redis.default.database"),
		})
		store, err = sredis.NewStoreWithOptions(client, limiter.StoreOptions{
			Prefix: facades.Config.GetString("app.name", "Goravel") + ":limiter",
		})
		if err != nil {
			panic(err.Error())
		}
	}

	limiterObj := limiter.New(store, rate)

	// use limit-once- as prefix to make sure the limiter only run once when the route has multiple limiter
	if c.GetBool("limit-once-" + key) {
		return limiterObj.Peek(c, key)
	} else {
		c.Set("limit-once-"+key, true)
		return limiterObj.Get(c, key)
	}
}

// RouteToKeyString change route path to key string
func RouteToKeyString(routeName string) string {
	routeName = strings.ReplaceAll(routeName, "/", "-")
	routeName = strings.ReplaceAll(routeName, ":", "_")
	return routeName
}
