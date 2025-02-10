package queue

import (
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	queueconsole "github.com/goravel/framework/queue/console"
)

var (
	LogFacade log.Log
	OrmFacade orm.Orm
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingQueue, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleQueue)
		}

		return NewApplication(app.MakeConfig()), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	LogFacade = app.MakeLog()
	OrmFacade = app.MakeOrm()

	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]console.Command{
		&queueconsole.JobMakeCommand{},
	})
}
