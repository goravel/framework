package migrations

import (
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	support2 "github.com/goravel/framework/console/support"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrateCommand(t *testing.T) {
	configServiceProvider := config.ServiceProvider{}
	configServiceProvider.Register()

	consoleApp := console.Application{}
	consoleApp.Init().Register([]support2.Command{
		MigrateCommand{},
	})

	assert.Panics(t, func() {
		consoleApp.Call("migrate")
	})
}
