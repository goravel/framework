package http

import (
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (database *ServiceProvider) Register() {
	app := Application{}
	facades.Request, facades.Response = app.Init()
}

func (database *ServiceProvider) Boot() {

}
