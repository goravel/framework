package foundation

import (
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/support/facades"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	configApp := config.ServiceProvider{}
	configApp.Register()

	facadesConfig := facades.Config
	facadesConfig.Add("app", map[string]interface{}{
		"providers": []contracts.ServiceProvider{
			&console.ServiceProvider{},
		},
	})

	assert.NotPanics(t, func() {
		app := Application{}
		app.Boot()
	})
}
