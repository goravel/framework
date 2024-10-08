package queue

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	queueConsole "github.com/goravel/framework/queue/console"
)

const Binding = "goravel.queue"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewApplication(app.MakeConfig(), app.MakeLog()), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]console.Command{
		&queueConsole.JobMakeCommand{},
	})
}
