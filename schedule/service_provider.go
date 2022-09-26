package schedule

import (
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register() {
	facades.Schedule = &Application{}
}

func (receiver *ServiceProvider) Boot() {

}
