package log

import "github.com/goravel/framework/support/facades"

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (log *ServiceProvider) Boot() {
}

//Register Register any application services.
func (log *ServiceProvider) Register() {
	app := Application{}
	facades.Log = app.Init()
}
