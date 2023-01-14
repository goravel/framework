package route

import (
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (route *ServiceProvider) Register() {
	facades.Route = NewGin()
}

func (route *ServiceProvider) Boot() {

}
