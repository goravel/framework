package log

import (
	"github.com/goravel/framework/contracts/log"
)

type Application struct {
}

func (app *Application) Init() log.Log {
	return NewLogrus()
}
