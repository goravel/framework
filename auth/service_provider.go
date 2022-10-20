package auth

import (
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (database *ServiceProvider) Register() {
	facades.Auth = NewApplication()
}

func (database *ServiceProvider) Boot() {
}
