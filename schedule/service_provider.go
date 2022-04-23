package schedule

import (
	"github.com/goravel/framework/support/facades"
)

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (receiver *ServiceProvider) Boot() {

}

//Register any application services.
func (receiver *ServiceProvider) Register() {
	facades.Schedule = &Application{}
}
