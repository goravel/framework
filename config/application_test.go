package config

import (
	testing2 "github.com/goravel/framework/support/testing"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	err := testing2.CreateEnv()
	assert.Nil(t, err)
	assert.NotPanics(t, func() {
		app := Application{}
		app.Init()
	})
}

func TestEnv(t *testing.T) {
	app := Application{}
	app.Init()
	assert.Equal(t, app.Env("APP_NAME"), "goravel")
}

func TestAdd(t *testing.T) {
	app := Application{}
	app.Init()
	app.Add("app", map[string]interface{}{
		"env": "local",
	})

	assert.Equal(t, app.GetString("app.env"), "local")
}

func TestGet(t *testing.T) {
	app := Application{}
	app.Init()

	assert.Equal(t, app.Get("APP_NAME").(string), "goravel")
}

func TestGetString(t *testing.T) {
	app := Application{}
	app.Init()

	assert.Equal(t, app.GetString("APP_NAME"), "goravel")
}

func TestGetInt(t *testing.T) {
	app := Application{}
	app.Init()

	assert.Equal(t, app.GetInt("DB_PORT"), 3306)
}

func TestGetBool(t *testing.T) {
	app := Application{}
	app.Init()

	assert.Equal(t, app.GetBool("APP_DEBUG"), true)

	err := os.Remove(".env")
	assert.Nil(t, err)
}
