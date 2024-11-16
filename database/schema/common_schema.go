package schema

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
)

type CommonSchema struct {
	grammar schema.Grammar
	orm     orm.Orm
}

func NewCommonSchema(grammar schema.Grammar, orm orm.Orm) *CommonSchema {
	return &CommonSchema{
		grammar: grammar,
		orm:     orm,
	}
}

func (r *CommonSchema) GetTables() ([]schema.Table, error) {
	var tables []schema.Table
	if err := r.orm.Query().Raw(r.grammar.CompileTables(r.orm.DatabaseName())).Scan(&tables); err != nil {
		return nil, err
	}

	return tables, nil
}

func (r *CommonSchema) GetViews() ([]schema.View, error) {
	var views []schema.View
	if err := r.orm.Query().Raw(r.grammar.CompileViews(r.orm.DatabaseName())).Scan(&views); err != nil {
		return nil, err
	}

	return views, nil
}
