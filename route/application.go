package route

import (
	"github.com/goravel/framework/contracts/route"
)

type Application struct {
}

func (app *Application) Init() route.Engine {
	return NewGin()
}
