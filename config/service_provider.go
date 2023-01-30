package config

import (
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (config *ServiceProvider) Register() {
	facades.Config = NewApplication(".env")
}

func (config *ServiceProvider) Boot() {

}
