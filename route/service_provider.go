package route

import (
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (route *ServiceProvider) Register() {
	app := Application{}
	facades.Route = app.Init()
}

func (route *ServiceProvider) Boot() {

}
