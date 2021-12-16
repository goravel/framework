package console

import (
	"github.com/goravel/framework/support/facades"
)

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (console *ServiceProvider) Boot() {

}

//Register Register any application services.
func (console *ServiceProvider) Register() {
	app := Application{}
	facades.Artisan = app.Init()
}
