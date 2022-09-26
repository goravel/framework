package grpc

import (
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (route *ServiceProvider) Register() {
	app := Application{}
	facades.Grpc = app.Init()
}

func (route *ServiceProvider) Boot() {

}
