package docker

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/database/gorm"
)

type Docker struct {
	app foundation.Application
}

func NewDocker(app foundation.Application) *Docker {
	return &Docker{
		app: app,
	}
}

func (receiver *Docker) Database(connection ...string) (testing.Database, error) {
	if len(connection) == 0 {
		return NewDatabase(receiver.app, "", gorm.NewInitializeImpl())
	} else {
		return NewDatabase(receiver.app, connection[0], gorm.NewInitializeImpl())
	}
}
