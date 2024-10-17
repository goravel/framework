package migration

import (
	"github.com/goravel/framework/contracts/config"
	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/migration"
	"github.com/goravel/framework/errors"
)

func GetDriver(config config.Config, schema contractsmigration.Schema) (contractsmigration.Driver, error) {
	driver := config.GetString("database.migrations.driver")

	switch driver {
	case contractsmigration.DriverDefault:
		table := config.GetString("database.migrations.table")

		return migration.NewDefaultDriver(schema, table), nil
	case contractsmigration.DriverSql:
		return migration.NewSqlDriver(config), nil
	default:
		return nil, errors.MigrationUnsupportedDriver.Args(driver).SetModule(errors.ModuleMigration)
	}
}
