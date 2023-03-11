package gorm

import (
	"fmt"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/facades"
)

func MysqlDsn(connection string, config contractsdatabase.Config) string {
	host := config.Host
	if host == "" {
		return ""
	}

	charset := facades.Config.GetString("database.connections." + connection + ".charset")
	loc := facades.Config.GetString("database.connections." + connection + ".loc")

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&multiStatements=true",
		config.Username, config.Password, host, config.Port, config.Database, charset, true, loc)
}

func PostgresqlDsn(connection string, config contractsdatabase.Config) string {
	host := config.Host
	if host == "" {
		return ""
	}

	sslmode := facades.Config.GetString("database.connections." + connection + ".sslmode")
	timezone := facades.Config.GetString("database.connections." + connection + ".timezone")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		host, config.Username, config.Password, config.Database, config.Port, sslmode, timezone)
}

func SqliteDsn(config contractsdatabase.Config) string {
	return fmt.Sprintf("%s?multi_stmts=true", config.Database)
}

func SqlserverDsn(connection string, config contractsdatabase.Config) string {
	host := config.Host
	if host == "" {
		return ""
	}

	charset := facades.Config.GetString("database.connections." + connection + ".charset")

	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&charset=%s&MultipleActiveResultSets=true",
		config.Username, config.Password, host, config.Port, config.Database, charset)
}
