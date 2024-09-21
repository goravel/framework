package migration

import (
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/orm"
)

type Repository struct {
	query  orm.Query
	schema migration.Schema
	table  string
}

func NewRepository(query orm.Query, schema migration.Schema, table string) *Repository {
	return &Repository{
		query:  query,
		schema: schema,
		table:  table,
	}
}

func (r *Repository) CreateRepository() error {
	return r.schema.Create(r.table, func(table migration.Blueprint) {
		table.ID()
		table.String("migration")
		table.Integer("batch")
	})
}

func (r *Repository) Delete(migration string) error {
	_, err := r.query.Exec("DELETE FROM ? WHERE migration = ?", r.table, migration)

	return err
}

func (r *Repository) DeleteRepository() error {
	return r.schema.DropIfExists(r.table)
}

func (r *Repository) GetLast() ([]migration.File, error) {
	var files []migration.File
	if err := r.query.Table(r.table).Where("batch", r.getLastBatchNumber()).OrderByDesc("migration").Get(&files); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *Repository) GetMigrations(steps int) ([]migration.File, error) {
	var files []migration.File
	if err := r.query.Table(r.table).Where("batch >= 1").OrderByDesc("batch").OrderByDesc("migration").Limit(steps).Get(&files); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *Repository) GetMigrationsByBatch(batch int) ([]migration.File, error) {
	var files []migration.File
	if err := r.query.Table(r.table).Where("batch", batch).OrderByDesc("migration").Get(&files); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *Repository) GetNextBatchNumber() int {
	return r.getLastBatchNumber() + 1
}

func (r *Repository) GetRan() ([]string, error) {
	var migrations []string
	if err := r.query.Table(r.table).OrderBy("batch").OrderBy("migration").Pluck("migration", &migrations); err != nil {
		return nil, err
	}

	return migrations, nil
}

func (r *Repository) Log(file string, batch int) error {
	return r.query.Table(r.table).Create(map[string]any{
		"migration": file,
		"batch":     batch,
	})
}

func (r *Repository) RepositoryExists() {
	//TODO implement me when schema.HasTable is implemented
	panic("implement me")
}

func (r *Repository) getLastBatchNumber() int {
	var batch int
	if err := r.query.Table(r.table).OrderBy("batch", "desc").Pluck("batch", &batch); err != nil {
		return 0
	}

	return batch
}
