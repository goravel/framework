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

func (r *Docker) Cache(connection string) (docker.Cache, error) {
	return nil, nil
}

func (r *Docker) Database(connection ...string) (docker.Database, error) {
	if len(connection) == 0 {
		return NewDatabase(r.app.MakeArtisan(), r.app.MakeConfig(), r.app.MakeOrm(), "")
	} else {
		return NewDatabase(r.app.MakeArtisan(), r.app.MakeConfig(), r.app.MakeOrm(), connection[0])
	}
}
