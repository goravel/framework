package support

import (
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/support/facades"
	testing2 "github.com/goravel/framework/support/testing"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetServer(t *testing.T) {
	initConfig()
	server, err := getServer("sync", "")
	assert.Nil(t, server)
	assert.NotNil(t, err)

	server, err = getServer("redis", "")
	assert.Nil(t, err)
	assert.NotNil(t, server)

	server, err = getServer("custom", "")
	assert.Nil(t, server)
	assert.NotNil(t, err)
}

func TestGetDriver(t *testing.T) {
	initConfig()
	assert.Equal(t, "sync", getDriver("sync"))
	assert.Equal(t, "redis", getDriver("redis"))
}

func TestGetRedisServer(t *testing.T) {
	initConfig()
	assert.NotNil(t, getRedisServer("default"))
}

func TestGetRedisConfig(t *testing.T) {
	initConfig()
	redisConfig, database, queue := getRedisConfig()
	assert.Equal(t, "127.0.0.1:6379", redisConfig)
	assert.Equal(t, 0, database)
	assert.Equal(t, "default", queue)
}

func initConfig() {
	testing2.CreateEnv()
	configServiceProvider := config.ServiceProvider{}
	configServiceProvider.Register()

	facadesConfig := facades.Config
	facadesConfig.Add("queue", map[string]interface{}{
		"default": facadesConfig.Env("QUEUE_CONNECTION", "redis"),
		"connections": map[string]interface{}{
			"sync": map[string]interface{}{
				"driver": "sync",
			},
			"redis": map[string]interface{}{
				"driver":      "redis",
				"connection":  "default",
				"queue": "default",
				"retry_after": 90,
			},
		},
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
