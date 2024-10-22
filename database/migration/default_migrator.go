package migration

import (
	"slices"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/orm"
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
	if err := r.artisan.Call("db:wipe --force"); err != nil {
		return err
	}
	if err := r.artisan.Call("migrate"); err != nil {
		return err
	}

	return nil
}

func (r *DefaultMigrator) Rollback(step, batch int) error {
	files, err := r.getFilesForRollback(step, batch)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		color.Infoln("Nothing to rollback")

		return nil
	}

	color.Infoln("Rolling back migration")

	for _, file := range files {
		migration := r.getMigrationViaFile(file)
		if migration == nil {
			color.Warnf("Migration not found: %s\n", file.Migration)

			continue
		}

		if err := r.runDown(migration); err != nil {
			return err
		}

		color.Infoln("Rolled back:", migration.Signature())
	}

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

func (r *DefaultMigrator) getFilesForRollback(step, batch int) ([]migration.File, error) {
	if step > 0 {
		return r.repository.GetMigrations(step)
	}

	if batch > 0 {
		return r.repository.GetMigrationsByBatch(batch)
	}

	return r.repository.GetLast()
}

func (r *DefaultMigrator) getMigrationViaFile(file migration.File) schema.Migration {
	for _, migration := range r.schema.Migrations() {
		if migration.Signature() == file.Migration {
			return migration
		}
	}

	return nil
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

func (r *DefaultMigrator) runUp(migration schema.Migration, batch int) error {
	defaultConnection := r.schema.GetConnection()
	if connectionMigration, ok := migration.(schema.Connection); ok {
		r.schema.SetConnection(connectionMigration.Connection())
	}

	return r.schema.Orm().Transaction(func(tx orm.Query) error {
		defaultQuery := r.schema.Orm().Query()
		r.schema.Orm().SetQuery(tx)

		if err := migration.Up(); err != nil {
			// reset the connection and query to default.
			r.schema.SetConnection(defaultConnection)
			r.schema.Orm().SetQuery(defaultQuery)

			return err
		}

		// repository.Log should be called in the default connection.
		r.schema.SetConnection(defaultConnection)
		r.schema.Orm().SetQuery(defaultQuery)

		return r.repository.Log(migration.Signature(), batch)
	})
}

func (r *DefaultMigrator) runDown(migration schema.Migration) error {
	defaultConnection := r.schema.GetConnection()
	if connectionMigration, ok := migration.(schema.Connection); ok {
		r.schema.SetConnection(connectionMigration.Connection())
	}

	return r.schema.Orm().Transaction(func(tx orm.Query) error {
		defaultQuery := r.schema.Orm().Query()
		r.schema.Orm().SetQuery(tx)

		if err := migration.Down(); err != nil {
			// reset the connection and query to default.
			r.schema.SetConnection(defaultConnection)
			r.schema.Orm().SetQuery(defaultQuery)

			return err
		}

		// repository.Log should be called in the default connection.
		r.schema.SetConnection(defaultConnection)
		r.schema.Orm().SetQuery(defaultQuery)

		return r.repository.Delete(migration.Signature())
	})
}
