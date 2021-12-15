package console

import "github.com/goravel/framework/support/facades"

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (console *ServiceProvider) Boot() {
	app := Application{}
	facades.Artisan = app.Init()
}

//Register Register any application services.
func (console *ServiceProvider) Register() {
}
