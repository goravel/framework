package console

import (
	"github.com/goravel/framework/contracts/console"
)

type Application struct {
}

func (app *Application) Init() console.Artisan {
	return NewCli()
}
