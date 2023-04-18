package gorm

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

func dialectors(connection string, configs []contractsdatabase.Config) ([]gorm.Dialector, error) {
	var dialectors []gorm.Dialector
	for _, config := range configs {
		dialector, err := dialector(connection, config)
		if err != nil {
			return nil, err
		}
		dialectors = append(dialectors, dialector)
	}

	return dialectors, nil
}

func dialector(connection string, config contractsdatabase.Config) (gorm.Dialector, error) {
	driver := facades.Config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))

	switch orm.Driver(driver) {
	case orm.DriverMysql:
		return mysqlDialector(connection, config), nil
	case orm.DriverPostgresql:
		return postgresqlDialector(connection, config), nil
	case orm.DriverSqlite:
		return sqliteDialector(config), nil
	case orm.DriverSqlserver:
		return sqlserverDialector(connection, config), nil
	default:
		return nil, fmt.Errorf("err database driver: %s, only support mysql, postgresql, sqlite and sqlserver", driver)
	}
}

func mysqlDialector(connection string, config contractsdatabase.Config) gorm.Dialector {
	dsn := MysqlDsn(connection, config)
	if dsn == "" {
		return nil
	}

	return mysql.New(mysql.Config{
		DSN: dsn,
	})
}

func postgresqlDialector(connection string, config contractsdatabase.Config) gorm.Dialector {
	dsn := PostgresqlDsn(connection, config)
	if dsn == "" {
		return nil
	}

	return postgres.New(postgres.Config{
		DSN: dsn,
	})
}

func sqliteDialector(config contractsdatabase.Config) gorm.Dialector {
	dsn := SqliteDsn(config)
	if dsn == "" {
		return nil
	}

	return sqlite.Open(dsn)
}

func sqlserverDialector(connection string, config contractsdatabase.Config) gorm.Dialector {
	dsn := SqlserverDsn(connection, config)
	if dsn == "" {
		return nil
	}

	return sqlserver.New(sqlserver.Config{
		DSN: dsn,
	})
}
