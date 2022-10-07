package console

import (
	"github.com/goravel/framework/console/console"
	console2 "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Boot() {
	receiver.registerCommands()
}

func (receiver *ServiceProvider) Register() {
	app := Application{}
	facades.Artisan = app.Init()
}

func (receiver *ServiceProvider) registerCommands() {
	facades.Artisan.Register([]console2.Command{
		&console.ListCommand{},
		&console.KeyGenerateCommand{},
		&console.MakeCommand{},
	})
}
