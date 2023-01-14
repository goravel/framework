package support

import (
	"fmt"

	"github.com/goravel/framework/facades"
)

func GetMysqlDsn(connection string) string {
	host := facades.Config.GetString("database.connections." + connection + ".host")
	if host == "" {
		return ""
	}

	port := facades.Config.GetString("database.connections." + connection + ".port")
	database := facades.Config.GetString("database.connections." + connection + ".database")
	username := facades.Config.GetString("database.connections." + connection + ".username")
	password := facades.Config.GetString("database.connections." + connection + ".password")
	charset := facades.Config.GetString("database.connections." + connection + ".charset")
	loc := facades.Config.GetString("database.connections." + connection + ".loc")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		username, password, host, port, database, charset, true, loc)
}

func GetPostgresqlDsn(connection string) string {
	host := facades.Config.GetString("database.connections." + connection + ".host")
	if host == "" {
		return ""
	}

	port := facades.Config.GetString("database.connections." + connection + ".port")
	database := facades.Config.GetString("database.connections." + connection + ".database")
	username := facades.Config.GetString("database.connections." + connection + ".username")
	password := facades.Config.GetString("database.connections." + connection + ".password")
	sslmode := facades.Config.GetString("database.connections." + connection + ".sslmode")
	timezone := facades.Config.GetString("database.connections." + connection + ".timezone")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, username, password, database, port, sslmode, timezone)
}

func GetSqliteDsn(connection string) string {
	return facades.Config.GetString("database.connections." + connection + ".database")
}

func GetSqlserverDsn(connection string) string {
	host := facades.Config.GetString("database.connections." + connection + ".host")
	if host == "" {
		return ""
	}

	port := facades.Config.GetString("database.connections." + connection + ".port")
	database := facades.Config.GetString("database.connections." + connection + ".database")
	username := facades.Config.GetString("database.connections." + connection + ".username")
	password := facades.Config.GetString("database.connections." + connection + ".password")

	return fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		username, password, host, port, database)
}
