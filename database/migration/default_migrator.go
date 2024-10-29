package migration

import (
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/console"
	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/collect"
	"github.com/goravel/framework/support/color"
	supportfile "github.com/goravel/framework/support/file"
)

type status struct {
	Name  string
	Batch int
	Ran   bool
}

type DefaultMigrator struct {
	artisan    console.Artisan
	creator    *DefaultCreator
	repository contractsmigration.Repository
	schema     contractsschema.Schema
}

func NewDefaultMigrator(artisan console.Artisan, schema contractsschema.Schema, table string) *DefaultMigrator {
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

func (r *DefaultMigrator) Reset() error {
	ran, err := r.repository.GetRan()
	if err != nil {
		return err
	}

	return r.Rollback(len(ran), 0)
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
			color.Warningf("Migration not found: %s\n", file.Migration)

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

func (r *DefaultMigrator) Status() error {
	if !r.repository.RepositoryExists() {
		color.Warningln("Migration table not found")

		return nil
	}

	batches, err := r.repository.GetMigrations()
	if err != nil {
		return err
	}

	migrationStatus := r.getStatusForMigrations(batches)
	if len(migrationStatus) == 0 {
		color.Warningln("No migrations found")

		return nil
	}

	maxNameLength := r.getMaxNameLength(migrationStatus)
	r.printTitle(maxNameLength)

	for _, s := range migrationStatus {
		color.Default().Print(fmt.Sprintf("%-*s", maxNameLength, s.Name))
		if s.Ran {
			color.Default().Printf(" | [%d] ", s.Batch)
			color.Green().Println("Ran")
		} else {
			color.Yellow().Println(" | Pending")
		}
	}

	return nil
}

func (r *DefaultMigrator) getFilesForRollback(step, batch int) ([]contractsmigration.File, error) {
	if step > 0 {
		return r.repository.GetMigrationsByStep(step)
	}

	if batch > 0 {
		return r.repository.GetMigrationsByBatch(batch)
	}

	return r.repository.GetLast()
}

func (r *DefaultMigrator) getMaxNameLength(migrationStatus []status) int {
	var length int
	for _, s := range migrationStatus {
		if len(s.Name) > length {
			length = len(s.Name)
		}
	}

	return length
}

func (r *DefaultMigrator) getMigrationViaFile(file contractsmigration.File) contractsschema.Migration {
	for _, migration := range r.schema.Migrations() {
		if migration.Signature() == file.Migration {
			return migration
		}
	}

	return nil
}

func (r *DefaultMigrator) getStatusForMigrations(batches []contractsmigration.File) []status {
	var migrationStatus []status

	for _, migration := range r.schema.Migrations() {
		var file contractsmigration.File
		collect.Each(batches, func(item contractsmigration.File, index int) {
			if item.Migration == migration.Signature() {
				file = item
				return
			}
		})

		if file.ID > 0 {
			migrationStatus = append(migrationStatus, status{
				Name:  migration.Signature(),
				Batch: file.Batch,
				Ran:   true,
			})
		} else {
			migrationStatus = append(migrationStatus, status{
				Name: migration.Signature(),
				Ran:  false,
			})
		}
	}

	return migrationStatus
}

func (r *DefaultMigrator) pendingMigrations(ran []string) []contractsschema.Migration {
	var pendingMigrations []contractsschema.Migration
	for _, migration := range r.schema.Migrations() {
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

func (r *DefaultMigrator) printTitle(maxNameLength int) {
	color.Default().Print(fmt.Sprintf("%-*s", maxNameLength, "Migration name"))
	color.Default().Println(" | Batch / Status")
	for i := 0; i < maxNameLength+17; i++ {
		color.Default().Print("-")
	}
	color.Default().Println()
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
	defaultConnection := r.schema.GetConnection()
	defaultQuery := r.schema.Orm().Query()
	if connectionMigration, ok := migration.(contractsschema.Connection); ok {
		r.schema.SetConnection(connectionMigration.Connection())
	}

	defer func() {
		// reset the connection and query to default, to avoid err and panic
		r.schema.Orm().SetQuery(defaultQuery)
		r.schema.SetConnection(defaultConnection)
	}()

	return r.schema.Orm().Transaction(func(tx orm.Query) error {
		r.schema.Orm().SetQuery(tx)

		if err := migration.Down(); err != nil {
			return err
		}

		// repository.Log should be called in the default connection.
		r.schema.Orm().SetQuery(defaultQuery)
		r.schema.SetConnection(defaultConnection)

		return r.repository.Delete(migration.Signature())
	})
}

func (r *DefaultMigrator) runUp(migration contractsschema.Migration, batch int) error {
	defaultConnection := r.schema.GetConnection()
	defaultQuery := r.schema.Orm().Query()
	if connectionMigration, ok := migration.(contractsschema.Connection); ok {
		r.schema.SetConnection(connectionMigration.Connection())
	}

	defer func() {
		// reset the connection and query to default, to avoid err and panic
		r.schema.Orm().SetQuery(defaultQuery)
		r.schema.SetConnection(defaultConnection)
	}()

	return r.schema.Orm().Transaction(func(tx orm.Query) error {
		r.schema.Orm().SetQuery(tx)

		if err := migration.Up(); err != nil {
			return err
		}

		// repository.Log should be called in the default connection.
		r.schema.Orm().SetQuery(defaultQuery)
		r.schema.SetConnection(defaultConnection)

		return r.repository.Log(migration.Signature(), batch)
	})
}
