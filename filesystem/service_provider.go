package filesystem

import (
	frameworkconfig "github.com/goravel/framework/config"
	configcontract "github.com/goravel/framework/contracts/config"
	filesystemcontract "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

var (
	ConfigFacade  configcontract.Config
	StorageFacade filesystemcontract.Storage
)

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(frameworkconfig.BindingFilesystem, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleFilesystem)
		}

		return NewStorage(config)
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {
	ConfigFacade = app.MakeConfig()
	StorageFacade = app.MakeStorage()
}
