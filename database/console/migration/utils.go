package migration

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/migration"
)

func GetDriver(config config.Config) (contractsmigration.Driver, error) {
	connection := config.GetString("database.default")
	driver := config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))

	switch driver {
	case contractsmigration.DriverDefault:
		return migration.NewDefaultDriver(), nil
	case contractsmigration.DriverSql:
		charset := config.GetString(fmt.Sprintf("database.connections.%s.charset", connection))

		return migration.NewSqlDriver(driver, charset), nil
	default:
		return nil, fmt.Errorf("unsupported migration driver: %s", driver)
	}
}
