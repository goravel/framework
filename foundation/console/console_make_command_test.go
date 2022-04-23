package console

import (
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts"
	console2 "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/facades"
	testing2 "github.com/goravel/framework/support/testing"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConsoleMakeCommand(t *testing.T) {
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
		&ConsoleMakeCommand{},
	})

	assert.NotPanics(t, func() {
		consoleApp.CallDontExit("make:command GoravelCommand")
	})

	assert.True(t, support.Helpers{}.ExistFile("app/console/commands/goravel_command.go"))
	assert.True(t, support.Helpers{}.RemoveFile("app"))
	err = os.Remove(".env")
	assert.Nil(t, err)
}
