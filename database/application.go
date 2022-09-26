package database

import (
	"github.com/gookit/color"
	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

type Application struct {
}

func (app *Application) Init() orm.Orm {
	return &Orm{}
}

func (app *Application) InitGorm() *gorm.DB {
	db, err := NewGormInstance(facades.Config.GetString("database.default"))
	if err != nil {
		color.Redln("init facades.Gorm error:", err.Error())

		return nil
	}

	return db
}
