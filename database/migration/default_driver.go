package migration

import (
	"slices"

	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

type DefaultDriver struct {
	creator    *DefaultCreator
	repository migration.Repository
	schema     migration.Schema
}

func NewDefaultDriver(schema migration.Schema, table string) *DefaultDriver {
	return &DefaultDriver{
		creator:    NewDefaultCreator(),
		repository: NewRepository(schema, table),
		schema:     schema,
	}
}

func (r *DefaultDriver) Create(name string) error {
	table, create := TableGuesser{}.Guess(name)

	stub := r.creator.GetStub(table, create)

	// Prepend timestamp to the file name.
	fileName := r.creator.GetFileName(name)

	// Create the up.sql file.
	if err := file.Create(r.creator.GetPath(fileName), r.creator.PopulateStub(stub, fileName, table)); err != nil {
		return err
	}

	return nil
}

func (r *DefaultDriver) Run() error {
	r.prepareDatabase()

	ran, err := r.repository.GetRan()
	if err != nil {
		return err
	}

	pendingMigrations := r.pendingMigrations(r.schema.Migrations(), ran)

	return r.runPending(pendingMigrations)
}

func (r *DefaultDriver) pendingMigrations(migrations []migration.Migration, ran []string) []migration.Migration {
	var pendingMigrations []migration.Migration
	for _, migration := range migrations {
		if !slices.Contains(ran, migration.Signature()) {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	return pendingMigrations
}

func (r *DefaultDriver) prepareDatabase() {
	if r.repository.RepositoryExists() {
		return
	}

	r.repository.CreateRepository()
}

func (r *DefaultDriver) runPending(migrations []migration.Migration) error {
	if len(migrations) == 0 {
		color.Infoln("Nothing to migrate")

		return nil
	}

	batch, err := r.repository.GetNextBatchNumber()
	if err != nil {
		return err
	}

	color.Infoln("Running migration")

	for _, migration := range migrations {
		color.Infoln("Running:", migration.Signature())

		if err := r.runUp(migration, batch); err != nil {
			return err
		}
	}

	return nil
}

func (r *DefaultDriver) runUp(file migration.Migration, batch int) error {
	if connectionMigration, ok := file.(migration.Connection); ok {
		previousConnection := r.schema.GetConnection()
		r.schema.SetConnection(connectionMigration.Connection())
		defer r.schema.SetConnection(previousConnection)
	}

	file.Up()

	return r.repository.Log(file.Signature(), batch)
}
