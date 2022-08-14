package migrations

import (
	"testing"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	console2 "github.com/goravel/framework/contracts/console"
	"github.com/stretchr/testify/assert"
)

func TestMigrateCommand(t *testing.T) {
	configServiceProvider := config.ServiceProvider{}
	configServiceProvider.Register()

	consoleApp := console.Application{}
	consoleApp.Init().Register([]console2.Command{
		&MigrateCommand{},
	})

	assert.Panics(t, func() {
		consoleApp.Call("migrate")
	})
}
