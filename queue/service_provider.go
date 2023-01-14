package queue

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/facades"
	queueConsole "github.com/goravel/framework/queue/console"
)

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register() {
	facades.Queue = NewApplication()
}

func (receiver *ServiceProvider) Boot() {
	receiver.registerCommands()
}

func (receiver *ServiceProvider) registerCommands() {
	facades.Artisan.Register([]console.Command{
		&queueConsole.JobMakeCommand{},
	})
}
