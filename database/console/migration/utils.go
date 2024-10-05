package migration

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/migration"
)

func GetDriver(config config.Config) (contractsmigration.Driver, error) {
	driver := config.GetString("database.migrations.driver")

	switch driver {
	case contractsmigration.DriverDefault:
		return migration.NewDefaultDriver(), nil
	case contractsmigration.DriverSql:
		connection := config.GetString("database.default")
		dbDriver := config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))
		charset := config.GetString(fmt.Sprintf("database.connections.%s.charset", connection))

		return migration.NewSqlDriver(dbDriver, charset), nil
	default:
		return nil, fmt.Errorf("unsupported migration driver: %s", driver)
	}
}
