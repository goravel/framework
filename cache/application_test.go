package cache

import (
	"os"
	"testing"
	"time"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/facades"
	goraveltesting "github.com/goravel/framework/testing"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	initConfig()

	assert.NotPanics(t, func() {
		app := Application{}
		app.Init()
	})
}

func TestClearCommand(t *testing.T) {
	initConfig()

	consoleServiceProvider := console.ServiceProvider{}
	consoleServiceProvider.Register()

	cacheServiceProvider := ServiceProvider{}
	cacheServiceProvider.Register()
	cacheServiceProvider.Boot()

	err := facades.Cache.Put("test-clear-command", "goravel", 5*time.Second)
	assert.Nil(t, err)
	assert.True(t, facades.Cache.Has("test-clear-command"))

	assert.NotPanics(t, func() {
		facades.Artisan.Call("cache:clear")
	})

	assert.False(t, facades.Cache.Has("test-clear-command"))
}

func initConfig() {
	goraveltesting.CreateEnv()
	configServiceProvider := config.ServiceProvider{}
	configServiceProvider.Register()

	facadesConfig := facades.Config
	facadesConfig.Add("cache", map[string]interface{}{
		"default": facadesConfig.Env("CACHE_DRIVER", "redis"),
		"stores": map[string]interface{}{
			"redis": map[string]interface{}{
				"driver":     "redis",
				"connection": "default",
			},
		},
		"prefix": "goravel_cache",
	})

	facadesConfig.Add("database", map[string]interface{}{
		"redis": map[string]interface{}{
			"default": map[string]interface{}{
				"host":     facadesConfig.Env("REDIS_HOST", "127.0.0.1"),
				"password": facadesConfig.Env("REDIS_PASSWORD", ""),
				"port":     facadesConfig.Env("REDIS_PORT", 6379),
				"database": facadesConfig.Env("REDIS_DB", 0),
			},
		},
	})

	os.Remove(".env")
}
