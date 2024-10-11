package migration

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratedatabase "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlserver"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/console/driver"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

// TODO Remove in v1.16
type SqlDriver struct {
	config     config.Config
	connection string
	creator    *SqlCreator
	driver     database.Driver
}

func NewSqlDriver(config config.Config, connection string) *SqlDriver {
	driver := database.Driver(config.GetString(fmt.Sprintf("database.connections.%s.driver", connection)))
	charset := config.GetString(fmt.Sprintf("database.connections.%s.charset", connection))

	return &SqlDriver{
		config:     config,
		connection: connection,
		creator:    NewSqlCreator(driver, charset),
		driver:     driver,
	}
}

func (r *SqlDriver) Create(name string) error {
	table, create := TableGuesser{}.Guess(name)

	upStub, downStub := r.creator.GetStub(table, create)

	// Create the up.sql file.
	if err := file.Create(r.creator.GetPath(name, "up"), r.creator.PopulateStub(upStub, table)); err != nil {
		return err
	}

	// Create the down.sql file.
	if err := file.Create(r.creator.GetPath(name, "down"), r.creator.PopulateStub(downStub, table)); err != nil {
		return err
	}

	return nil
}

func (r *SqlDriver) Run(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	migrator, err := r.getMigrator(paths[0])
	if err != nil {
		return err
	}

	if err = migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		color.Red().Println("Migration failed:", err.Error())
	}

	return nil
}

func (r *SqlDriver) getMigrator(path string) (*migrate.Migrate, error) {
	if support.RelativePath != "" {
		path = fmt.Sprintf("%s/%s", support.RelativePath, str.Of(path).LTrim("./"))
	}
	path = fmt.Sprintf("file://%s", path)

	configBuilder := databasedb.NewConfigBuilder(r.config, r.connection)
	writeConfigs := configBuilder.Writes()
	if len(writeConfigs) == 0 {
		return nil, errors.OrmDatabaseConfigNotFound
	}

	table := r.config.GetString("database.migrations.table")
	dsn := databasedb.Dsn(writeConfigs[0])
	if dsn == "" {
		return nil, errors.OrmFailedToGenerateDNS.Args(r.connection)
	}

	var (
		databaseName string
		db           *sql.DB
		driver       migratedatabase.Driver
		err          error
	)

	switch r.driver {
	case database.DriverMysql:
		databaseName = "mysql"
		db, err = sql.Open(databaseName, dsn)
		if err != nil {
			return nil, err
		}

		driver, err = mysql.WithInstance(db, &mysql.Config{
			MigrationsTable: table,
		})
	case database.DriverPostgres:
		databaseName = "postgres"
		db, err = sql.Open(databaseName, dsn)
		if err != nil {
			return nil, err
		}

		driver, err = postgres.WithInstance(db, &postgres.Config{
			MigrationsTable: table,
		})
	case database.DriverSqlite:
		databaseName = "sqlite3"
		db, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, err
		}

		driver, err = sqlite.WithInstance(db, &sqlite.Config{
			MigrationsTable: table,
		})
	case database.DriverSqlserver:
		databaseName = "sqlserver"
		db, err = sql.Open(databaseName, dsn)
		if err != nil {
			return nil, err
		}

		driver, err = sqlserver.WithInstance(db, &sqlserver.Config{
			MigrationsTable: table,
		})
	default:
		err = errors.OrmDriverNotSupported.Args(r.driver)
	}

	if err != nil {
		return nil, err
	}

	return migrate.NewWithDatabaseInstance(path, databaseName, driver)
}
