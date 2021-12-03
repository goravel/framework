package database

import (
	"fmt"
	"github.com/goravel/framework/database/helpers"
	"github.com/goravel/framework/support/facades"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Application struct {
}

func (app *Application) Init() *gorm.DB {
	var db *gorm.DB
	config := helpers.GetDatabaseConfig()
	if config["host"] == "" || config["username"] == "" {
		return db
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		config["username"], config["password"], config["host"], config["port"], config["database"], config["charset"], true, "Local")

	gormConfig := mysql.New(mysql.Config{
		DSN: dsn,
	})

	var level gormLogger.LogLevel
	if facades.Config.GetBool("app.debug") {
		level = gormLogger.Info
	} else {
		level = gormLogger.Error
	}

	db, _ = gorm.Open(gormConfig, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger.Default.LogMode(level),
	})

	return db
}
