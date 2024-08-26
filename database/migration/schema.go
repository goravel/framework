package migration

import (
	"github.com/goravel/framework/contracts/database/migration"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

var _ migration.Schema = (*Schema)(nil)

type Schema struct {
	connection string
	orm        contractsorm.Orm
	migrations []migration.Migration
}

func NewSchema(orm contractsorm.Orm) *Schema {
	return &Schema{
		orm: orm,
	}
}

func (r *Schema) Connection(name string) migration.Schema {
	r.connection = name

	return r
}

func (r *Schema) Register(migrations []migration.Migration) {
	r.migrations = migrations
}

func (r *Schema) Sql(sql string) {
	// TODO catch error and rollback
	_, _ = r.orm.Connection(r.connection).Query().Exec(sql)
}
