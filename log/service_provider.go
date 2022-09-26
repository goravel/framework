package log

import "github.com/goravel/framework/facades"

type ServiceProvider struct {
}

func (log *ServiceProvider) Register() {
	app := Application{}
	facades.Log = app.Init()
}

func (log *ServiceProvider) Boot() {

}
