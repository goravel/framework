package db

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	databasecontract "github.com/goravel/framework/contracts/database"
)

type Dsn interface {
	Mysql(config databasecontract.Config) string
	Postgresql(config databasecontract.Config) string
	Sqlite(config databasecontract.Config) string
	Sqlserver(config databasecontract.Config) string
}

type DsnImpl struct {
	config     config.Config
	connection string
}

func NewDsnImpl(config config.Config, connection string) *DsnImpl {
	return &DsnImpl{
		config:     config,
		connection: connection,
	}
}

func (d *DsnImpl) Mysql(config databasecontract.Config) string {
	host := config.Host
	if host == "" {
		return ""
	}

	charset := d.config.GetString("database.connections." + d.connection + ".charset")
	loc := d.config.GetString("database.connections." + d.connection + ".loc")

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&multiStatements=true",
		config.Username, config.Password, host, config.Port, config.Database, charset, true, loc)
}

func (d *DsnImpl) Postgresql(config databasecontract.Config) string {
	host := config.Host
	if host == "" {
		return ""
	}

	sslmode := d.config.GetString("database.connections." + d.connection + ".sslmode")
	timezone := d.config.GetString("database.connections." + d.connection + ".timezone")

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
		config.Username, config.Password, host, config.Port, config.Database, sslmode, timezone)
}

func (d *DsnImpl) Sqlite(config databasecontract.Config) string {
	return fmt.Sprintf("%s?multi_stmts=true", config.Database)
}

func (d *DsnImpl) Sqlserver(config databasecontract.Config) string {
	host := config.Host
	if host == "" {
		return ""
	}

	charset := d.config.GetString("database.connections." + d.connection + ".charset")

	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&charset=%s&MultipleActiveResultSets=true",
		config.Username, config.Password, host, config.Port, config.Database, charset)
}
