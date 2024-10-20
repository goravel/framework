package migration

import (
	"slices"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

type DefaultMigrator struct {
	artisan    console.Artisan
	creator    *DefaultCreator
	repository migration.Repository
	schema     schema.Schema
}

func NewDefaultMigrator(artisan console.Artisan, schema schema.Schema, table string) *DefaultMigrator {
	return &DefaultMigrator{
		artisan:    artisan,
		creator:    NewDefaultCreator(),
		repository: NewRepository(schema, table),
		schema:     schema,
	}
}

func (r *DefaultMigrator) Create(name string) error {
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

// TODO Remove this function and move the logic to the migrate:fresh command when the sql migrator is removed.
func (r *DefaultMigrator) Fresh() error {
	r.artisan.Call("db:wipe --force")
	r.artisan.Call("migrate")

	return nil
}

func (r *DefaultMigrator) Run() error {
	if err := r.prepareDatabase(); err != nil {
		return err
	}

	ran, err := r.repository.GetRan()
	if err != nil {
		return err
	}

	pendingMigrations := r.pendingMigrations(r.schema.Migrations(), ran)

	return r.runPending(pendingMigrations)
}

func (r *DefaultMigrator) pendingMigrations(migrations []schema.Migration, ran []string) []schema.Migration {
	var pendingMigrations []schema.Migration
	for _, migration := range migrations {
		if !slices.Contains(ran, migration.Signature()) {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	return pendingMigrations
}

func (r *DefaultMigrator) prepareDatabase() error {
	if r.repository.RepositoryExists() {
		return nil
	}

	return r.repository.CreateRepository()
}

func (r *DefaultMigrator) runPending(migrations []schema.Migration) error {
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

func (r *DefaultMigrator) runUp(file schema.Migration, batch int) error {
	if connectionMigration, ok := file.(schema.Connection); ok {
		previousConnection := r.schema.GetConnection()
		r.schema.SetConnection(connectionMigration.Connection())
		defer r.schema.SetConnection(previousConnection)
	}

	if err := file.Up(); err != nil {
		return err
	}

	return r.repository.Log(file.Signature(), batch)
}
