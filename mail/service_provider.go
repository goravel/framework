package mail

import (
	"fmt"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/mail/console"
)

const Binding = "goravel.mail"

type ServiceProvider struct {
}

func (route *ServiceProvider) Register(app foundation.Application) {
	app.Bind(Binding, func(app foundation.Application) (any, error) {
		return NewApplication(app.MakeConfig(), app.MakeQueue()), nil
	})
}

func (route *ServiceProvider) Boot(app foundation.Application) {
	err := app.MakeQueue().Register([]queue.Job{
		NewSendMailJob(app.MakeConfig()),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to register mail job: %v", err))
	}

	app.MakeArtisan().Register([]consolecontract.Command{
		console.NewMailMakeCommand(),
	})
}
