package database

import (
	"errors"
	"fmt"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/support/facades"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Gorm struct {
	connection      string
	defaultInstance *gorm.DB
	instances       map[string]*gorm.DB
}

func (r *Gorm) Connection(name string) contractsdatabase.Gorm {
	defaultConnection := facades.Config.GetString("database.default")
	if name == "" {
		name = defaultConnection
	}

	r.connection = name

	if _, exist := r.instances[name]; exist {
		return r
	}

	gormConfig, err := getGormConfig(name)
	if err != nil {
		facades.Log.Errorf("init gorm config error: %v", err)
	}

	var logLevel gormLogger.LogLevel
	if facades.Config.GetBool("app.debug") {
		logLevel = gormLogger.Info
	} else {
		logLevel = gormLogger.Error
	}

	db, err := gorm.Open(gormConfig, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger.Default.LogMode(logLevel),
	})
	if err != nil {
		facades.Log.Errorf("gorm open ddatabase error: %v", err)

		return r
	}

	r.instances[name] = db

	if name == defaultConnection {
		r.defaultInstance = db
	}

	return r
}

func (r *Gorm) Query() *gorm.DB {
	if r.connection == "" {
		if r.defaultInstance == nil {
			r.Connection("")
		}

		return r.defaultInstance
	}

	instance, exist := r.instances[r.connection]
	if !exist {
		return nil
	}

	r.connection = ""

	return instance
}

func getGormConfig(connection string) (gorm.Dialector, error) {
	defaultDatabase := facades.Config.GetString("database.default")
	driver := facades.Config.GetString("database.connections." + defaultDatabase + ".driver")
	switch driver {
	case Mysql:
		return getMysqlGormConfig(connection), nil
	case Postgresql:
		return getPostgresqlGormConfig(connection), nil
	case Sqlite:
		return getSqliteGormConfig(connection), nil
	case Sqlserver:
		return getSqlserverGormConfig(connection), nil
	default:
		return nil, errors.New("database driver only support mysql, postgresql, sqlite and sqlserver")
	}
}

func getMysqlGormConfig(connection string) gorm.Dialector {
	host := facades.Config.GetString("database.connections." + connection + ".host")
	port := facades.Config.GetString("database.connections." + connection + ".port")
	database := facades.Config.GetString("database.connections." + connection + ".database")
	username := facades.Config.GetString("database.connections." + connection + ".username")
	password := facades.Config.GetString("database.connections." + connection + ".password")
	charset := facades.Config.GetString("database.connections." + connection + ".charset")
	loc := facades.Config.GetString("database.connections." + connection + ".charset")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		username, password, host, port, database, charset, true, loc)

	return mysql.New(mysql.Config{
		DSN: dsn,
	})
}

func getPostgresqlGormConfig(connection string) gorm.Dialector {
	host := facades.Config.GetString("database.connections." + connection + ".host")
	port := facades.Config.GetString("database.connections." + connection + ".port")
	database := facades.Config.GetString("database.connections." + connection + ".database")
	username := facades.Config.GetString("database.connections." + connection + ".username")
	password := facades.Config.GetString("database.connections." + connection + ".password")
	sslmode := facades.Config.GetString("database.connections." + connection + ".sslmode")
	timezone := facades.Config.GetString("database.connections." + connection + ".timezone")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, username, password, database, port, sslmode, timezone)

	return postgres.New(postgres.Config{
		DSN: dsn,
	})
}

func getSqliteGormConfig(connection string) gorm.Dialector {
	database := facades.Config.GetString("database.connections." + connection + ".database")

	return sqlite.Open(database)
}

func getSqlserverGormConfig(connection string) gorm.Dialector {
	host := facades.Config.GetString("database.connections." + connection + ".host")
	port := facades.Config.GetString("database.connections." + connection + ".port")
	database := facades.Config.GetString("database.connections." + connection + ".database")
	username := facades.Config.GetString("database.connections." + connection + ".username")
	password := facades.Config.GetString("database.connections." + connection + ".password")

	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		username, password, host, port, database)

	return sqlserver.New(sqlserver.Config{
		DSN: dsn,
	})
}
