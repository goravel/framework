package gorm

import (
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/support"
	"github.com/goravel/framework/facades"
)

func config(connection string) (gorm.Dialector, error) {
	driver := facades.Config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))

	switch orm.Driver(driver) {
	case orm.DriverMysql:
		return mysqlConfig(connection), nil
	case orm.DriverPostgresql:
		return postgresqlConfig(connection), nil
	case orm.DriverSqlite:
		return sqliteConfig(connection), nil
	case orm.DriverSqlserver:
		return sqlserverConfig(connection), nil
	default:
		return nil, errors.New(fmt.Sprintf("err database driver: %s, only support mysql, postgresql, sqlite and sqlserver", driver))
	}
}

func mysqlConfig(connection string) gorm.Dialector {
	dsn := support.GetMysqlDsn(connection)
	if dsn == "" {
		return nil
	}

	return mysql.New(mysql.Config{
		DSN: dsn,
	})
}

func postgresqlConfig(connection string) gorm.Dialector {
	dsn := support.GetPostgresqlDsn(connection)
	if dsn == "" {
		return nil
	}

	return postgres.New(postgres.Config{
		DSN: dsn,
	})
}

func sqliteConfig(connection string) gorm.Dialector {
	dsn := support.GetSqliteDsn(connection)
	if dsn == "" {
		return nil
	}

	return sqlite.Open(dsn)
}

func sqlserverConfig(connection string) gorm.Dialector {
	dsn := support.GetSqlserverDsn(connection)
	if dsn == "" {
		return nil
	}

	return sqlserver.New(sqlserver.Config{
		DSN: dsn,
	})
}
