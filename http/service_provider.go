package http

import (
	"github.com/goravel/framework/support/facades"
)

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (database *ServiceProvider) Boot() {
}

//Register Register any application services.
func (database *ServiceProvider) Register() {
	app := Application{}
	facades.Request, facades.Response = app.Init()
}
