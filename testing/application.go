package testing

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/testing/docker"
)

type Application struct {
	app foundation.Application
}

func NewApplication(app foundation.Application) *Application {
	return &Application{
		app: app,
	}
}

func (receiver *Application) Docker() testing.Docker {
	return docker.NewDocker(receiver.app)
}
