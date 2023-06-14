package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
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
func (receiver MigrateCreator) Create(name string, table string, create bool) error {
	// First we will get the stub file for the migration, which serves as a type
	// of template for the migration. Once we have those we will populate the
	// various place-holders, save the file, and run the post create event.
	upStub, downStub := receiver.getStub(table, create)

	//Create the up.sql file.
	if err := file.Create(receiver.getPath(name, "up"), receiver.populateStub(upStub, table)); err != nil {
		return err
	}

	//Create the down.sql file.
	if err := file.Create(receiver.getPath(name, "down"), receiver.populateStub(downStub, table)); err != nil {
		return err
	}

	return nil
}

// getStub Get the migration stub file.
func (receiver MigrateCreator) getStub(table string, create bool) (string, string) {
	if table == "" {
		return "", ""
	}

	driver := receiver.config.GetString("database.connections." + receiver.config.GetString("database.default") + ".driver")
	switch orm.Driver(driver) {
	case orm.DriverPostgresql:
		if create {
			return PostgresqlStubs{}.CreateUp(), PostgresqlStubs{}.CreateDown()
		}

		return PostgresqlStubs{}.UpdateUp(), PostgresqlStubs{}.UpdateDown()
	case orm.DriverSqlite:
		if create {
			return SqliteStubs{}.CreateUp(), SqliteStubs{}.CreateDown()
		}

		return SqliteStubs{}.UpdateUp(), SqliteStubs{}.UpdateDown()
	case orm.DriverSqlserver:
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
func (receiver MigrateCreator) populateStub(stub string, table string) string {
	stub = strings.ReplaceAll(stub, "DummyDatabaseCharset", receiver.config.GetString("database.connections."+receiver.config.GetString("database.default")+".charset"))

	if table != "" {
		stub = strings.ReplaceAll(stub, "DummyTable", table)
	}

	return stub
}

// getPath Get the full path to the migration.
func (receiver MigrateCreator) getPath(name string, category string) string {
	pwd, _ := os.Getwd()

	return fmt.Sprintf("%s/database/migrations/%s_%s.%s.sql", pwd, carbon.Now().ToShortDateTimeString(), name, category)
}
