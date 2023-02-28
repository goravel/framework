package crypt

import (
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (crypt *ServiceProvider) Register() {
	facades.Crypt = NewApplication()
}

func (crypt *ServiceProvider) Boot() {

}
