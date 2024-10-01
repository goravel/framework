package console

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlserver"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/console/driver"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/support"
)

func getMigrate(config config.Config) (*migrate.Migrate, error) {
	connection := config.GetString("database.default")
	driver := config.GetString("database.connections." + connection + ".driver")
	dir := "file://./database/migrations"
	if support.RelativePath != "" {
		dir = fmt.Sprintf("file://%s/database/migrations", support.RelativePath)
	}

	configBuilder := databasedb.NewConfigBuilder(config, connection)
	writeConfigs := configBuilder.Writes()
	if len(writeConfigs) == 0 {
		return nil, errors.New("not found database configuration")
	}

	switch database.Driver(driver) {
	case database.DriverMysql:
		mysqlDsn := databasedb.Dsn(writeConfigs[0])
		if mysqlDsn == "" {
			return nil, nil
		}

		db, err := sql.Open("mysql", mysqlDsn)
		if err != nil {
			return nil, err
		}

		instance, err := mysql.WithInstance(db, &mysql.Config{
			MigrationsTable: config.GetString("database.migrations"),
		})
		if err != nil {
			return nil, err
		}

		return migrate.NewWithDatabaseInstance(dir, "mysql", instance)
	case database.DriverPostgres:
		postgresDsn := databasedb.Dsn(writeConfigs[0])
		if postgresDsn == "" {
			return nil, nil
		}

		db, err := sql.Open("postgres", postgresDsn)
		if err != nil {
			return nil, err
		}

		instance, err := postgres.WithInstance(db, &postgres.Config{
			MigrationsTable: config.GetString("database.migrations"),
		})
		if err != nil {
			return nil, err
		}

		return migrate.NewWithDatabaseInstance(dir, "postgres", instance)
	case database.DriverSqlite:
		sqliteDsn := databasedb.Dsn(writeConfigs[0])
		if sqliteDsn == "" {
			return nil, nil
		}

		db, err := sql.Open("sqlite", sqliteDsn)
		if err != nil {
			return nil, err
		}

		instance, err := sqlite.WithInstance(db, &sqlite.Config{
			MigrationsTable: config.GetString("database.migrations"),
		})
		if err != nil {
			return nil, err
		}

		return migrate.NewWithDatabaseInstance(dir, "sqlite3", instance)
	case database.DriverSqlserver:
		sqlserverDsn := databasedb.Dsn(writeConfigs[0])
		if sqlserverDsn == "" {
			return nil, nil
		}

		db, err := sql.Open("sqlserver", sqlserverDsn)
		if err != nil {
			return nil, err
		}

		instance, err := sqlserver.WithInstance(db, &sqlserver.Config{
			MigrationsTable: config.GetString("database.migrations"),
		})

		if err != nil {
			return nil, err
		}

		return migrate.NewWithDatabaseInstance(dir, "sqlserver", instance)
	default:
		return nil, errors.New("database driver only support mysql, postgres, sqlite and sqlserver")
	}
}
