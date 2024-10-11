package migration

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/migration"
)

func GetDriver(config config.Config, schema contractsmigration.Schema) (contractsmigration.Driver, error) {
	driver := config.GetString("database.migrations.driver")
	table := config.GetString("database.migrations.table")

	switch driver {
	case contractsmigration.DriverDefault:
		return migration.NewDefaultDriver(schema, table), nil
	case contractsmigration.DriverSql:
		connection := config.GetString("database.default")

		return migration.NewSqlDriver(config, connection), nil
	default:
		return nil, fmt.Errorf("unsupported migration driver: %s", driver)
	}
}
