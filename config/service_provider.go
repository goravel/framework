package config

import (
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (config *ServiceProvider) Register() {
	app := Application{}
	facades.Config = app.Init()
}

func (config *ServiceProvider) Boot() {

}
