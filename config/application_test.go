package config

import (
	"os"
	"testing"

	"github.com/goravel/framework/testing/file"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	err := file.CreateEnv()
	assert.Nil(t, err)
	assert.NotPanics(t, func() {
		app := Application{}
		app.Init()
	})
}

func TestEnv(t *testing.T) {
	app := Application{}
	app.Init()

	assert.Equal(t, "goravel", app.Env("APP_NAME").(string))
	assert.Equal(t, "127.0.0.1", app.Env("DB_HOST", "127.0.0.1").(string))
}

func TestAdd(t *testing.T) {
	app := Application{}
	app.Init()
	app.Add("app", map[string]interface{}{
		"env": "local",
	})

	assert.Equal(t, "local", app.GetString("app.env"))
}

func TestGet(t *testing.T) {
	app := Application{}
	app.Init()

	assert.Equal(t, "goravel", app.Get("APP_NAME").(string))
}

func TestGetString(t *testing.T) {
	app := Application{}
	app.Init()

	app.Add("database", map[string]interface{}{
		"default": app.Env("DB_CONNECTION", "mysql"),
		"connections": map[string]interface{}{
			"mysql": map[string]interface{}{
				"host": app.Env("DB_HOST", "127.0.0.1"),
			},
		},
	})

	assert.Equal(t, "goravel", app.GetString("APP_NAME"))
	assert.Equal(t, "127.0.0.1", app.GetString("database.connections.mysql.host"))
	assert.Equal(t, "mysql", app.GetString("database.default"))
}

func TestGetInt(t *testing.T) {
	app := Application{}
	app.Init()

	assert.Equal(t, app.GetInt("DB_PORT"), 3306)
}

func TestGetBool(t *testing.T) {
	app := Application{}
	app.Init()

	assert.Equal(t, true, app.GetBool("APP_DEBUG"))

	err := os.Remove(".env")
	assert.Nil(t, err)
}
