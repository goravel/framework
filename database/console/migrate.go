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
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/console/driver"
	"github.com/goravel/framework/database/db"
	"github.com/goravel/framework/support"
)

func getMigrate(config config.Config) (*migrate.Migrate, error) {
	connection := config.GetString("database.default")
	driver := config.GetString("database.connections." + connection + ".driver")
	dir := "file://./database/migrations"
	if support.RelativePath != "" {
		dir = fmt.Sprintf("file://%s/database/migrations", support.RelativePath)
	}

	gormConfig := db.NewConfigImpl(config, connection)
	writeConfigs := gormConfig.Writes()
	if len(writeConfigs) == 0 {
		return nil, errors.New("not found database configuration")
	}

	switch orm.Driver(driver) {
	case orm.DriverMysql:
		dsn := db.NewDsnImpl(config, connection)
		mysqlDsn := dsn.Mysql(writeConfigs[0])
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
	case orm.DriverPostgresql:
		dsn := db.NewDsnImpl(config, connection)
		postgresqlDsn := dsn.Postgresql(writeConfigs[0])
		if postgresqlDsn == "" {
			return nil, nil
		}

		db, err := sql.Open("postgres", postgresqlDsn)
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
	case orm.DriverSqlite:
		dsn := db.NewDsnImpl(config, "")
		sqliteDsn := dsn.Sqlite(writeConfigs[0])
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
	case orm.DriverSqlserver:
		dsn := db.NewDsnImpl(config, connection)
		sqlserverDsn := dsn.Sqlserver(writeConfigs[0])
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
		return nil, errors.New("database driver only support mysql, postgresql, sqlite and sqlserver")
	}
}
