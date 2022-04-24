package console

import (
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts"
	console2 "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/facades"
	testing2 "github.com/goravel/framework/support/testing"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestKeyGenerateCommand(t *testing.T) {
	err := testing2.CreateEnv()
	assert.Nil(t, err)

	configApp := config.ServiceProvider{}
	configApp.Register()

	facadesConfig := facades.Config
	facadesConfig.Add("app", map[string]interface{}{
		"providers": []contracts.ServiceProvider{},
	})

	consoleApp := console.Application{}
	consoleApp.Init().Register([]console2.Command{
		&KeyGenerateCommand{},
	})

	assert.NotPanics(t, func() {
		consoleApp.Call("key:generate")
	})

	err = os.Remove(".env")
	assert.Nil(t, err)
}
