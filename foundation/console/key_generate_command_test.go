package console

import (
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	support2 "github.com/goravel/framework/console/support"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/facades"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestKeyGenerateCommand(t *testing.T) {
	err := support.CreateEnv()
	assert.Nil(t, err)

	configApp := config.ServiceProvider{}
	configApp.Register()

	facadesConfig := facades.Config
	facadesConfig.Add("app", map[string]interface{}{
		"providers": []support.ServiceProvider{},
	})

	consoleApp := console.Application{}
	consoleApp.Init().Register([]support2.Command{
		KeyGenerateCommand{},
	})

	assert.NotPanics(t, func() {
		consoleApp.Call("key:generate")
	})

	err = os.Remove(".env")
	assert.Nil(t, err)
}
