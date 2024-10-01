package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/migration"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type MigrateCreator struct {
	config config.Config
}

func NewMigrateCreator(config config.Config) *MigrateCreator {
	return &MigrateCreator{
		config: config,
	}
}

// Create a new migration
func (receiver *MigrateCreator) Create(name string, table string, create bool) error {
	// First we will get the stub file for the migration, which serves as a type
	// of template for the migration. Once we have those we will populate the
	// various place-holders, save the file, and run the post create event.
	upStub, downStub := receiver.getStub(table, create)

	// Create the up.sql file.
	if err := file.Create(receiver.getPath(name, "up"), receiver.populateStub(upStub, table)); err != nil {
		return err
	}

	// Create the down.sql file.
	if err := file.Create(receiver.getPath(name, "down"), receiver.populateStub(downStub, table)); err != nil {
		return err
	}

	return nil
}

// getStub Get the migration stub file.
func (receiver *MigrateCreator) getStub(table string, create bool) (string, string) {
	if table == "" {
		return "", ""
	}

	driver := receiver.config.GetString("database.connections." + receiver.config.GetString("database.default") + ".driver")
	switch database.Driver(driver) {
	case database.DriverPostgres:
		if create {
			return migration.PostgresStubs{}.CreateUp(), migration.PostgresStubs{}.CreateDown()
		}

		return migration.PostgresStubs{}.UpdateUp(), migration.PostgresStubs{}.UpdateDown()
	case database.DriverSqlite:
		if create {
			return migration.SqliteStubs{}.CreateUp(), migration.SqliteStubs{}.CreateDown()
		}

		return migration.SqliteStubs{}.UpdateUp(), migration.SqliteStubs{}.UpdateDown()
	case database.DriverSqlserver:
		if create {
			return migration.SqlserverStubs{}.CreateUp(), migration.SqlserverStubs{}.CreateDown()
		}

		return migration.SqlserverStubs{}.UpdateUp(), migration.SqlserverStubs{}.UpdateDown()
	default:
		if create {
			return migration.MysqlStubs{}.CreateUp(), migration.MysqlStubs{}.CreateDown()
		}

		return migration.MysqlStubs{}.UpdateUp(), migration.MysqlStubs{}.UpdateDown()
	}
}

// populateStub Populate the place-holders in the migration stub.
func (receiver *MigrateCreator) populateStub(stub string, table string) string {
	stub = strings.ReplaceAll(stub, "DummyDatabaseCharset", receiver.config.GetString("database.connections."+receiver.config.GetString("database.default")+".charset"))

	if table != "" {
		stub = strings.ReplaceAll(stub, "DummyTable", table)
	}

	return stub
}

// getPath Get the full path to the migration.
func (receiver *MigrateCreator) getPath(name string, category string) string {
	pwd, _ := os.Getwd()

	return fmt.Sprintf("%s/database/migrations/%s_%s.%s.sql", pwd, carbon.Now().ToShortDateTimeString(), name, category)
}
