package database

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
)

type Seeder interface {
	Run(ctx console.Context) error
	SetContainer(container foundation.Container)
	SetCommand(command console.Context)
}
