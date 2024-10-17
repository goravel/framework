package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

type DefaultCreator struct {
}

func NewDefaultCreator() *DefaultCreator {
	return &DefaultCreator{}
}

// GetStub Get the migration stub file.
func (r *DefaultCreator) GetStub(table string, create bool) string {
	if table == "" {
		return Stubs{}.Empty()
	}

	if create {
		return Stubs{}.Create()
	}

	return Stubs{}.Update()
}

// PopulateStub Populate the place-holders in the migration stub.
func (r *DefaultCreator) PopulateStub(stub, signature, table string) string {
	stub = strings.ReplaceAll(stub, "DummyMigration", str.Of(signature).Prepend("m_").Studly().String())
	stub = strings.ReplaceAll(stub, "DummySignature", signature)
	stub = strings.ReplaceAll(stub, "DummyTable", table)

	return stub
}

// GetPath Get the full path to the migration.
func (r *DefaultCreator) GetPath(name string) string {
	pwd, _ := os.Getwd()

	return filepath.Join(pwd, "database", "migrations", name+".go")
}

// GetFileName Get the full path to the migration.
func (r *DefaultCreator) GetFileName(name string) string {
	return fmt.Sprintf("%s_%s", carbon.Now().ToShortDateTimeString(), name)
}
