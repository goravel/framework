package filesystem

import (
	configcontract "github.com/goravel/framework/contracts/config"
	filesystemcontract "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.filesystem"

var configModule configcontract.Config
var storageModule filesystemcontract.Storage

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	configModule = app.MakeConfig()
	storageModule = app.MakeStorage()

	app.Singleton(Binding, func() (any, error) {
		return NewStorage(app.MakeConfig()), nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {

}
