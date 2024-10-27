package migration

import (
	"slices"

	"github.com/goravel/framework/contracts/console"
	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/color"
	supportfile "github.com/goravel/framework/support/file"
)

type DefaultMigrator struct {
	artisan    console.Artisan
	creator    *DefaultCreator
	migrations map[string]contractsschema.Migration
	repository contractsmigration.Repository
	schema     contractsschema.Schema
}

func NewDefaultMigrator(artisan console.Artisan, schema contractsschema.Schema, table string) *DefaultMigrator {
	migrations := make(map[string]contractsschema.Migration)
	for _, m := range schema.Migrations() {
		migrations[m.Signature()] = m
	}

	return &DefaultMigrator{
		artisan:    artisan,
		creator:    NewDefaultCreator(),
		migrations: migrations,
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
	if err := supportfile.Create(r.creator.GetPath(fileName), r.creator.PopulateStub(stub, fileName, table)); err != nil {
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

	pendingMigrations := r.pendingMigrations(ran)

	return r.runPending(pendingMigrations)
}

func (r *DefaultMigrator) getFilesForRollback(step, batch int) ([]contractsmigration.File, error) {
	if step > 0 {
		return r.repository.GetMigrations(step)
	}

	if batch > 0 {
		return r.repository.GetMigrationsByBatch(batch)
	}

	return r.repository.GetLast()
}

func (r *DefaultMigrator) getMigrationViaFile(file contractsmigration.File) contractsschema.Migration {
	if m, exists := r.migrations[file.Migration]; exists {
		return m
	}
	return nil
}

func (r *DefaultMigrator) pendingMigrations(ran []string) []contractsschema.Migration {
	var pendingMigrations []contractsschema.Migration
	for name, migration := range r.migrations {
		if !slices.Contains(ran, name) {
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

func (r *DefaultMigrator) runPending(migrations []contractsschema.Migration) error {
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

func (r *DefaultMigrator) runDown(migration contractsschema.Migration) error {
	return r.runMigration(migration, func() error {
		if err := migration.Down(); err != nil {
			return err
		}
		return r.repository.Delete(migration.Signature())
	})
}

func (r *DefaultMigrator) runUp(migration contractsschema.Migration, batch int) error {
	return r.runMigration(migration, func() error {
		if err := migration.Up(); err != nil {
			return err
		}
		return r.repository.Log(migration.Signature(), batch)
	})
}

func (r *DefaultMigrator) runMigration(migration contractsschema.Migration, operation func() error) error {
	defaultConnection := r.schema.GetConnection()
	defaultQuery := r.schema.Orm().Query()
	if connectionMigration, ok := migration.(contractsschema.Connection); ok {
		r.schema.SetConnection(connectionMigration.Connection())
	}

	defer func() {
		r.schema.Orm().SetQuery(defaultQuery)
		r.schema.SetConnection(defaultConnection)
	}()

	return r.schema.Orm().Transaction(func(tx orm.Query) error {
		r.schema.Orm().SetQuery(tx)

		if err := operation(); err != nil {
			return err
		}

		// Reset to default connection for repository operations
		r.schema.Orm().SetQuery(defaultQuery)
		r.schema.SetConnection(defaultConnection)

		return nil
	})
}
