package database

import (
	"github.com/goravel/framework/contracts/database"
)

const (
	Mysql      = "mysql"
	Postgresql = "postgresql"
	Sqlite     = "sqlite"
	Sqlserver  = "sqlserver"
)

type Application struct {
}

func (app *Application) InitDB() database.DB {
	return nil
}

func (app *Application) InitGorm() database.Gorm {
	return &Gorm{}
}
