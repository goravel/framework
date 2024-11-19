package schema

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/grammars"
	"github.com/goravel/framework/database/schema/processors"
)

type SqliteSchema struct {
	schema.CommonSchema

	grammar   *grammars.Sqlite
	orm       orm.Orm
	prefix    string
	processor processors.Sqlite
}

func NewSqliteSchema(grammar *grammars.Sqlite, orm orm.Orm, prefix string) *SqliteSchema {
	return &SqliteSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewSqlite(),
	}
}

func (r *SqliteSchema) DropAllTables() error {
	if _, err := r.orm.Query().Exec(r.grammar.CompileEnableWriteableSchema()); err != nil {
		return err
	}
	if _, err := r.orm.Query().Exec(r.grammar.CompileDropAllTables(nil)); err != nil {
		return err
	}
	if _, err := r.orm.Query().Exec(r.grammar.CompileDisableWriteableSchema()); err != nil {
		return err
	}
	if _, err := r.orm.Query().Exec(r.grammar.CompileRebuild()); err != nil {
		return err
	}

	return nil
}

func (r *SqliteSchema) DropAllTypes() error {
	return nil
}

func (r *SqliteSchema) DropAllViews() error {
	return r.orm.Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(r.grammar.CompileEnableWriteableSchema()); err != nil {
			return err
		}
		if _, err := tx.Exec(r.grammar.CompileDropAllViews(nil)); err != nil {
			return err
		}
		if _, err := tx.Exec(r.grammar.CompileDisableWriteableSchema()); err != nil {
			return err
		}
		if _, err := tx.Exec(r.grammar.CompileRebuild()); err != nil {
			return err
		}

		return nil
	})
}

func (r *SqliteSchema) GetColumns(table string) ([]schema.Column, error) {
	table = r.prefix + table

	var dbColumns []processors.DBColumn
	if err := r.orm.Query().Raw(r.grammar.CompileColumns("", table)).Scan(&dbColumns); err != nil {
		return nil, err
	}

	return r.processor.ProcessColumns(dbColumns), nil
}

func (r *SqliteSchema) GetIndexes(table string) ([]schema.Index, error) {
	table = r.prefix + table

	var dbIndexes []processors.DBIndex
	if err := r.orm.Query().Raw(r.grammar.CompileIndexes("", table)).Scan(&dbIndexes); err != nil {
		return nil, err
	}

	return r.processor.ProcessIndexes(dbIndexes), nil
}

func (r *SqliteSchema) GetTypes() ([]schema.Type, error) {
	return nil, nil
}
