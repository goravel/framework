package console

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/database/sqlserver"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/facades"
)

func getMigrate() (*migrate.Migrate, error) {
	connection := facades.Config.GetString("database.default")
	driver := facades.Config.GetString("database.connections." + connection + ".driver")
	dir := "file://./database/migrations"
	_, writeConfigs, err := gorm.Configs(connection)
	if err != nil {
		return nil, err
	}

	switch database.Driver(driver) {
	case database.DriverMysql:
		dsn := gorm.MysqlDsn(connection, writeConfigs[0])
		if dsn == "" {
			return nil, nil
		}

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}

		instance, err := mysql.WithInstance(db, &mysql.Config{
			MigrationsTable: facades.Config.GetString("database.migrations"),
		})
		if err != nil {
			return nil, err
		}

		return migrate.NewWithDatabaseInstance(dir, "mysql", instance)
	case database.DriverPostgresql:
		dsn := gorm.PostgresqlDsn(connection, writeConfigs[0])
		if dsn == "" {
			return nil, nil
		}

		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, err
		}

		instance, err := postgres.WithInstance(db, &postgres.Config{
			MigrationsTable: facades.Config.GetString("database.migrations"),
		})
		if err != nil {
			return nil, err
		}

		return migrate.NewWithDatabaseInstance(dir, "postgres", instance)
	case database.DriverSqlite:
		dsn := gorm.SqliteDsn(writeConfigs[0])
		if dsn == "" {
			return nil, nil
		}

		db, err := sql.Open("sqlite3", dsn)
		if err != nil {
			return nil, err
		}

		instance, err := sqlite3.WithInstance(db, &sqlite3.Config{
			MigrationsTable: facades.Config.GetString("database.migrations"),
		})
		if err != nil {
			return nil, err
		}

		return migrate.NewWithDatabaseInstance(dir, "sqlite3", instance)
	case database.DriverSqlserver:
		dsn := gorm.SqlserverDsn(connection, writeConfigs[0])
		if dsn == "" {
			return nil, nil
		}

		db, err := sql.Open("sqlserver", dsn)
		if err != nil {
			return nil, err
		}

		instance, err := sqlserver.WithInstance(db, &sqlserver.Config{
			MigrationsTable: facades.Config.GetString("database.migrations"),
		})

		if err != nil {
			return nil, err
		}

		return migrate.NewWithDatabaseInstance(dir, "sqlserver", instance)
	default:
		return nil, errors.New("database driver only support mysql, postgresql, sqlite and sqlserver")
	}
}
