package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	app := Application{}
	app.Init()
	assert.Nil(t, nil)
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

	assert.Equal(t, app.GetInt("APP_INT"), 1)
}

func TestGetBool(t *testing.T) {
	app := Application{}
	app.Init()

	assert.Equal(t, app.GetBool("APP_BOOL"), true)
}
