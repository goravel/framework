package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type SqlDriver struct {
	config config.Config
}

func NewSqlDriver(config config.Config) *SqlDriver {
	return &SqlDriver{
		config: config,
	}
}

func (r *SqlDriver) Create(name string) error {
	// We will attempt to guess the table name if this the migration has
	// "create" in the name. This will allow us to provide a convenient way
	// of creating migrations that create new tables for the application.
	table, create := TableGuesser{}.Guess(name)

	// First we will get the stub file for the migration, which serves as a type
	// of template for the migration. Once we have those we will populate the
	// various place-holders, save the file, and run the post create event.
	upStub, downStub := r.getStub(table, create)

	// Create the up.sql file.
	if err := file.Create(r.getPath(name, "up"), r.populateStub(upStub, table)); err != nil {
		return err
	}

	// Create the down.sql file.
	if err := file.Create(r.getPath(name, "down"), r.populateStub(downStub, table)); err != nil {
		return err
	}

	return nil
}

// getStub Get the migration stub file.
func (r *SqlDriver) getStub(table string, create bool) (string, string) {
	if table == "" {
		return "", ""
	}

	driver := r.config.GetString("database.connections." + r.config.GetString("database.default") + ".driver")
	switch database.Driver(driver) {
	case database.DriverPostgres:
		if create {
			return PostgresStubs{}.CreateUp(), PostgresStubs{}.CreateDown()
		}

		return PostgresStubs{}.UpdateUp(), PostgresStubs{}.UpdateDown()
	case database.DriverSqlite:
		if create {
			return SqliteStubs{}.CreateUp(), SqliteStubs{}.CreateDown()
		}

		return SqliteStubs{}.UpdateUp(), SqliteStubs{}.UpdateDown()
	case database.DriverSqlserver:
		if create {
			return SqlserverStubs{}.CreateUp(), SqlserverStubs{}.CreateDown()
		}

		return SqlserverStubs{}.UpdateUp(), SqlserverStubs{}.UpdateDown()
	default:
		if create {
			return MysqlStubs{}.CreateUp(), MysqlStubs{}.CreateDown()
		}

		return MysqlStubs{}.UpdateUp(), MysqlStubs{}.UpdateDown()
	}
}

// populateStub Populate the place-holders in the migration stub.
func (r *SqlDriver) populateStub(stub string, table string) string {
	stub = strings.ReplaceAll(stub, "DummyDatabaseCharset", r.config.GetString("database.connections."+r.config.GetString("database.default")+".charset"))

	if table != "" {
		stub = strings.ReplaceAll(stub, "DummyTable", table)
	}

	return stub
}

// getPath Get the full path to the migration.
func (r *SqlDriver) getPath(name string, category string) string {
	pwd, _ := os.Getwd()

	return filepath.Join(pwd, "database", "migrations", fmt.Sprintf("%s_%s.%s.sql", carbon.Now().ToShortDateTimeString(), name, category))
}
