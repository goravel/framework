package foundation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/facades"
)

func TestInit(t *testing.T) {
	configApp := config.ServiceProvider{}
	configApp.Register()

	facadesConfig := facades.Config
	facadesConfig.Add("app", map[string]any{
		"providers": []contracts.ServiceProvider{
			&console.ServiceProvider{},
		},
	})

	assert.NotPanics(t, func() {
		app := Application{}
		app.Boot()
	})
}
