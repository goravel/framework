package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type DefaultDriver struct {
}

func NewDefaultDriver() *DefaultDriver {
	return &DefaultDriver{}
}

func (r *DefaultDriver) Create(name string) error {
	// We will attempt to guess the table name if this the migration has
	// "create" in the name. This will allow us to provide a convenient way
	// of creating migrations that create new tables for the application.
	table, create := TableGuesser{}.Guess(name)

	// First we will get the stub file for the migration, which serves as a type
	// of template for the migration. Once we have those we will populate the
	// various place-holders, save the file, and run the post create event.
	stub := r.getStub(table, create)

	// Prepend timestamp to the file name.
	fileName := r.getFileName(name)

	// Create the up.sql file.
	if err := file.Create(r.getPath(fileName), r.populateStub(stub, fileName)); err != nil {
		return err
	}

	return nil
}

// getStub Get the migration stub file.
func (r *DefaultDriver) getStub(table string, create bool) string {
	if table == "" {
		return Stubs{}.Empty()
	}

	if create {
		return Stubs{}.Create()
	}

	return Stubs{}.Update()
}

// populateStub Populate the place-holders in the migration stub.
func (r *DefaultDriver) populateStub(stub, fileName string) string {
	stub = strings.ReplaceAll(stub, "DummyMigration", str.Of(fileName).Prepend("m_").Studly().String())
	stub = strings.ReplaceAll(stub, "DummyName", fileName)

	return stub
}

// getPath Get the full path to the migration.
func (r *DefaultDriver) getPath(name string) string {
	pwd, _ := os.Getwd()

	return filepath.Join(pwd, "database", "migrations", name+".go")
}

func (r *DefaultDriver) getFileName(name string) string {
	return fmt.Sprintf("%s_%s", carbon.Now().ToShortDateTimeString(), name)
}
