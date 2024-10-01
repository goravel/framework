package db

import (
	"fmt"

	"github.com/goravel/framework/contracts/database"
)

func Dsn(config database.FullConfig) string {
	switch config.Driver {
	case database.DriverMysql:
		host := config.Host
		if host == "" {
			return ""
		}

		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&multiStatements=true",
			config.Username, config.Password, host, config.Port, config.Database, config.Charset, true, config.Loc)
	case database.DriverPostgres:
		host := config.Host
		if host == "" {
			return ""
		}

		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
			config.Username, config.Password, host, config.Port, config.Database, config.Sslmode, config.Timezone)
	case database.DriverSqlite:
		return fmt.Sprintf("%s?multi_stmts=true", config.Database)
	case database.DriverSqlserver:
		host := config.Host
		if host == "" {
			return ""
		}

		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&charset=%s&MultipleActiveResultSets=true",
			config.Username, config.Password, host, config.Port, config.Database, config.Charset)
	default:
		return ""
	}
}
