package database

import (
	"testing"

	"github.com/goravel/framework/config"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	configApp := config.ServiceProvider{}
	configApp.Register()

	assert.NotPanics(t, func() {
		app := Application{}
		app.Init()
		app.InitGorm()
	})
}
