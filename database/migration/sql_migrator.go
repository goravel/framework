package migration

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratedatabase "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlserver"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/console/driver"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

// TODO Remove in v1.16
type SqlMigrator struct {
	configBuilder *databasedb.ConfigBuilder
	creator       *SqlCreator
	migrator      *migrate.Migrate
	table         string
}

func NewSqlMigrator(config config.Config) (*SqlMigrator, error) {
	connection := config.GetString("database.default")
	charset := config.GetString(fmt.Sprintf("database.connections.%s.charset", connection))
	dbDriver := database.Driver(config.GetString(fmt.Sprintf("database.connections.%s.driver", connection)))
	table := config.GetString("database.migrations.table")
	configBuilder := databasedb.NewConfigBuilder(config, connection)
	migrator, err := getMigrator(configBuilder, table)
	if err != nil {
		return nil, err
	}

	return &SqlMigrator{
		configBuilder: configBuilder,
		creator:       NewSqlCreator(dbDriver, charset),
		migrator:      migrator,
		table:         table,
	}, nil
}

func (r *SqlMigrator) Create(name string) error {
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

func (r *SqlMigrator) Fresh() error {
	if err := r.migrator.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	// Recreate the migrations table.
	migrator, err := getMigrator(r.configBuilder, r.table)
	if err != nil {
		return err
	}

	r.migrator = migrator

	return r.Run()
}

func (r *SqlMigrator) Reset() error {
	if err := r.migrator.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.MigrationResetFailed.Args(err)
	}

	color.Successln("Migration reset success")

	return nil
}

func (r *SqlMigrator) Rollback(step, batch int) error {
	if err := r.migrator.Steps(-step); err != nil {
		var errShortLimit migrate.ErrShortLimit
		if errors.Is(err, migrate.ErrNoChange) || errors.Is(err, migrate.ErrNilVersion) || errors.As(err, &errShortLimit) {
			return nil
		}

		return errors.MigrationRollbackFailed.Args(err)
	}

	return nil
}

func (r *SqlMigrator) Run() error {
	if err := r.migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (r *SqlMigrator) Status() error {
	version, dirty, err := r.migrator.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			color.Warningln("No migrations found")

			return nil
		} else {
			return errors.MigrationGetStatusFailed.Args(err)
		}
	}
	if dirty {
		color.Warningln("Migration status: dirty")
	}

	color.Successln(fmt.Sprintf("Migration version: %d", version))

	return nil
}

func getMigrator(configBuilder *databasedb.ConfigBuilder, table string) (*migrate.Migrate, error) {
	path := "file://./database/migrations"
	if support.RelativePath != "" {
		path = fmt.Sprintf("file://%s/database/migrations", support.RelativePath)
	}

	writeConfigs := configBuilder.Writes()
	if len(writeConfigs) == 0 {
		return nil, errors.OrmDatabaseConfigNotFound
	}

	writeConfig := writeConfigs[0]
	dsn := databasedb.Dsn(writeConfigs[0])
	if dsn == "" {
		return nil, errors.OrmFailedToGenerateDNS.Args(writeConfig.Connection)
	}

	var (
		databaseName string
		db           *sql.DB
		dbDriver     migratedatabase.Driver
		err          error
	)

	switch writeConfig.Driver {
	case database.DriverMysql:
		databaseName = "mysql"
		db, err = sql.Open(databaseName, dsn)
		if err != nil {
			return nil, err
		}

		dbDriver, err = mysql.WithInstance(db, &mysql.Config{
			MigrationsTable: table,
		})
	case database.DriverPostgres:
		databaseName = "postgres"
		db, err = sql.Open(databaseName, dsn)
		if err != nil {
			return nil, err
		}

		dbDriver, err = postgres.WithInstance(db, &postgres.Config{
			MigrationsTable: table,
		})
	case database.DriverSqlite:
		databaseName = "sqlite3"
		db, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, err
		}

		dbDriver, err = driver.WithInstance(db, &driver.Config{
			MigrationsTable: table,
		})
	case database.DriverSqlserver:
		databaseName = "sqlserver"
		db, err = sql.Open(databaseName, dsn)
		if err != nil {
			return nil, err
		}

		dbDriver, err = sqlserver.WithInstance(db, &sqlserver.Config{
			MigrationsTable: table,
		})
	default:
		err = errors.OrmDriverNotSupported.Args(writeConfig.Connection)
	}

	if err != nil {
		return nil, err
	}

	return migrate.NewWithDatabaseInstance(path, databaseName, dbDriver)
}
