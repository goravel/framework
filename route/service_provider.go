package route

import (
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/support/facades"
)

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (route *ServiceProvider) Boot() {

}

//Register Register any application services.
func (route *ServiceProvider) Register() {
	app := foundation.Application{}
	if !app.RunningInConsole() {
		app := Application{}
		facades.Route = app.Init()
	}
}
