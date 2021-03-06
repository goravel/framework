package console

import (
	"github.com/goravel/framework/console/console"
	console2 "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/facades"
)

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (receiver *ServiceProvider) Boot() {
	receiver.registerCommands()
}

//Register any application services.
func (receiver *ServiceProvider) Register() {
	app := Application{}
	facades.Artisan = app.Init()
}

func (receiver *ServiceProvider) registerCommands() {
	facades.Artisan.Register([]console2.Command{
		&console.ConsoleMakeCommand{},
	})
}
