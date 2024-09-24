package migration

import (
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/orm"
)

type Repository struct {
	orm    orm.Orm
	schema migration.Schema
	table  string
}

func NewRepository(orm orm.Orm, schema migration.Schema, table string) *Repository {
	return &Repository{
		orm:    orm,
		schema: schema,
		table:  table,
	}
}

func (r *Repository) CreateRepository() {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) Delete(migration string) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) DeleteRepository() {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) GetLast() {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) GetMigrationBatches() {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) GetMigrations(steps int) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) GetMigrationsByBatch(batch int) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) GetNextBatchNumber() {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) GetRan() {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) Log(file, batch string) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) RepositoryExists() {
	//TODO implement me
	panic("implement me")
}
