package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/support/carbon"
)

type SqlCreator struct {
	driver  database.Driver
	charset string
}

func NewSqlCreator(driver database.Driver, charset string) *SqlCreator {
	return &SqlCreator{
		driver:  driver,
		charset: charset,
	}
}

// GetPath Get the full path to the migration.
func (r *SqlCreator) GetPath(name string, category string) string {
	pwd, _ := os.Getwd()

	return filepath.Join(pwd, "database", "migrations", fmt.Sprintf("%s_%s.%s.sql", carbon.Now().ToShortDateTimeString(), name, category))
}

// GetStub Get the migration stub file.
func (r *SqlCreator) GetStub(table string, create bool) (string, string) {
	if table == "" {
		return "", ""
	}

	switch r.driver {
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

// PopulateStub Populate the place-holders in the migration stub.
func (r *SqlCreator) PopulateStub(stub string, table string) string {
	stub = strings.ReplaceAll(stub, "DummyDatabaseCharset", r.charset)

	if table != "" {
		stub = strings.ReplaceAll(stub, "DummyTable", table)
	}

	return stub
}
