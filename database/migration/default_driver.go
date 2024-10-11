package migration

import (
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/support/file"
)

type DefaultDriver struct {
	repository migration.Repository
}

func NewDefaultDriver(schema migration.Schema, table string) *DefaultDriver {
	return &DefaultDriver{
		repository: NewRepository(schema, table),
	}
}

func (r *DefaultDriver) Create(name string) error {
	creator := NewDefaultCreator()
	table, create := TableGuesser{}.Guess(name)

	stub := creator.GetStub(table, create)

	// Prepend timestamp to the file name.
	fileName := creator.GetFileName(name)

	// Create the up.sql file.
	if err := file.Create(creator.GetPath(fileName), creator.PopulateStub(stub, fileName)); err != nil {
		return err
	}

	return nil
}

func (r *DefaultDriver) Run(paths []string) error {
	return nil
}

// prepareDatabase Prepare the migration database for running.
func (r *DefaultDriver) prepareDatabase() error {
	return nil
}
