package route

import (
	"github.com/goravel/framework/support/facades"
)

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (route *ServiceProvider) Boot() {
	app := Application{}
	facades.Route = app.Init()
}

//Register Register any application services.
func (route *ServiceProvider) Register() {

}
