package log

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (log *ServiceProvider) Register(app foundation.Application) {
	facades.Log = NewLogrusApplication()
}

func (log *ServiceProvider) Boot(app foundation.Application) {

}
