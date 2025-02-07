package docker

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/testing/docker"
)

type Docker struct {
	app foundation.Application
}

func NewDocker(app foundation.Application) *Docker {
	return &Docker{
		app: app,
	}
}

func (receiver *Docker) Database(connection ...string) (docker.Database, error) {
	if len(connection) == 0 {
		return NewDatabase(receiver.app, "")
	} else {
		return NewDatabase(receiver.app, connection[0])
	}
}
