package database

import (
	"github.com/goravel/framework/contracts/database/orm"
)

type Application struct {
}

func (app *Application) Init() orm.Orm {
	return NewOrm()
}
