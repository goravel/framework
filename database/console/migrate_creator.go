package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
	supporttime "github.com/goravel/framework/support/time"
)

type MigrateCreator struct {
}

//Create a new migration
func (receiver MigrateCreator) Create(name string, table string, create bool) {
	// First we will get the stub file for the migration, which serves as a type
	// of template for the migration. Once we have those we will populate the
	// various place-holders, save the file, and run the post create event.
	upStub, downStub := receiver.getStub(table, create)

	//Create the up.sql file.
	file.Create(receiver.getPath(name, "up"), receiver.populateStub(upStub, table))

	//Create the down.sql file.
	file.Create(receiver.getPath(name, "down"), receiver.populateStub(downStub, table))
}

//getStub Get the migration stub file.
func (receiver MigrateCreator) getStub(table string, create bool) (string, string) {
	if table == "" {
		return "", ""
	}

	driver := facades.Config.GetString("database.connections." + facades.Config.GetString("database.default") + ".driver")
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

//populateStub Populate the place-holders in the migration stub.
func (receiver MigrateCreator) populateStub(stub string, table string) string {
	stub = strings.ReplaceAll(stub, "DummyDatabaseCharset", facades.Config.GetString("database.connections."+facades.Config.GetString("database.default")+".charset"))

	if table != "" {
		stub = strings.ReplaceAll(stub, "DummyTable", table)
	}

	return stub
}

//getPath Get the full path to the migration.
func (receiver MigrateCreator) getPath(name string, category string) string {
	pwd, _ := os.Getwd()

	return fmt.Sprintf("%s/database/migrations/%s_%s.%s.sql", pwd, supporttime.Now().Format("20060102150405"), name, category)
}
