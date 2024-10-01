package gorm

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/db"
)

func getDialectors(configs []database.FullConfig) ([]gorm.Dialector, error) {
	var dialectors []gorm.Dialector

	for _, config := range configs {
		var dialector gorm.Dialector
		dsn := db.Dsn(config)
		if dsn == "" {
			return nil, fmt.Errorf("failed to generate DSN for connection: %s", config.Connection)
		}

		switch config.Driver {
		case database.DriverMysql:
			dialector = mysql.New(mysql.Config{
				DSN: dsn,
			})
		case database.DriverPostgres:
			dialector = postgres.New(postgres.Config{
				DSN: dsn,
			})
		case database.DriverSqlite:
			dialector = sqlite.Open(dsn)
		case database.DriverSqlserver:
			dialector = sqlserver.New(sqlserver.Config{
				DSN: dsn,
			})
		default:
			return nil, fmt.Errorf("err database driver: %s, only support mysql, postgres, sqlite and sqlserver", config.Driver)
		}

		dialectors = append(dialectors, dialector)
	}

	return dialectors, nil
}
