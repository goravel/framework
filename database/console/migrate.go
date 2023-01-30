package console

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/database/sqlserver"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/support"
	"github.com/goravel/framework/facades"
)

func getMigrate() (*migrate.Migrate, error) {
	connection := facades.Config.GetString("database.default")
	driver := facades.Config.GetString("database.connections." + connection + ".driver")
	dir := "file://./database/migrations"
	switch orm.Driver(driver) {
	case orm.DriverMysql:
		dsn := support.GetMysqlDsn(connection)
		if dsn == "" {
			return nil, nil
		}

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}

		//if err := db.Ping(); err != nil {
		//	return nil, errors.New("Could not ping to database: " + err.Error())
		//}

		instance, err := mysql.WithInstance(db, &mysql.Config{
			MigrationsTable: facades.Config.GetString("database.migrations"),
		})
		if err != nil {
			return nil, err
		}

		return migrate.NewWithDatabaseInstance(dir, "mysql", instance)
	case orm.DriverPostgresql:
		dsn := support.GetPostgresqlDsn(connection)
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
	case orm.DriverSqlite:
		dsn := support.GetSqliteDsn(connection)
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
	case orm.DriverSqlserver:
		dsn := support.GetSqlserverDsn(connection)
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
