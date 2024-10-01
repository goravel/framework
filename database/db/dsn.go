package db

import (
	"fmt"

	"github.com/goravel/framework/contracts/database"
)

func Dsn(config database.FullConfig) string {
	if config.Host == "" && config.Driver != database.DriverSqlite {
		return ""
	}

	switch config.Driver {
	case database.DriverMysql:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&multiStatements=true",
			config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset, true, config.Loc)
	case database.DriverPostgres:
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
			config.Username, config.Password, config.Host, config.Port, config.Database, config.Sslmode, config.Timezone)
	case database.DriverSqlite:
		return fmt.Sprintf("%s?multi_stmts=true", config.Database)
	case database.DriverSqlserver:
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&charset=%s&MultipleActiveResultSets=true",
			config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset)
	default:
		return ""
	}
}
