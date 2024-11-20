package schema

import (
	"github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/grammars"
	"github.com/goravel/framework/database/schema/processors"
)

type MysqlSchema struct {
	contractsschema.CommonSchema

	grammar   *grammars.Mysql
	orm       orm.Orm
	prefix    string
	processor processors.Mysql
}

func NewMysqlSchema(grammar *grammars.Mysql, orm orm.Orm, prefix string) *MysqlSchema {
	return &MysqlSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewMysql(),
	}
}

func (r *MysqlSchema) DropAllTables() error {
	tables, err := r.GetTables()
	if err != nil {
		return err
	}

	if len(tables) == 0 {
		return nil
	}

	return r.orm.Transaction(func(tx orm.Query) error {
		if _, err = tx.Exec(r.grammar.CompileDisableForeignKeyConstraints()); err != nil {
			return err
		}

		var dropTables []string
		for _, table := range tables {
			dropTables = append(dropTables, table.Name)
		}
		if _, err = tx.Exec(r.grammar.CompileDropAllTables(dropTables)); err != nil {
			return err
		}

		if _, err = tx.Exec(r.grammar.CompileEnableForeignKeyConstraints()); err != nil {
			return err
		}

		return err
	})
}

func (r *MysqlSchema) DropAllTypes() error {
	return nil
}

func (r *MysqlSchema) DropAllViews() error {
	views, err := r.GetViews()
	if err != nil {
		return err
	}
	if len(views) == 0 {
		return nil
	}

	var dropViews []string
	for _, view := range views {
		dropViews = append(dropViews, view.Name)
	}

	_, err = r.orm.Query().Exec(r.grammar.CompileDropAllViews(dropViews))

	return err
}

func (r *MysqlSchema) GetColumns(table string) ([]contractsschema.Column, error) {
	table = r.prefix + table

	var dbColumns []processors.DBColumn
	if err := r.orm.Query().Raw(r.grammar.CompileColumns(r.orm.DatabaseName(), table)).Scan(&dbColumns); err != nil {
		return nil, err
	}

	return r.processor.ProcessColumns(dbColumns), nil
}

func (r *MysqlSchema) GetIndexes(table string) ([]contractsschema.Index, error) {
	table = r.prefix + table

	var dbIndexes []processors.DBIndex
	if err := r.orm.Query().Raw(r.grammar.CompileIndexes(r.orm.DatabaseName(), table)).Scan(&dbIndexes); err != nil {
		return nil, err
	}

	return r.processor.ProcessIndexes(dbIndexes), nil
}

func (r *MysqlSchema) GetTypes() ([]contractsschema.Type, error) {
	return nil, nil
}
