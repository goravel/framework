package schema

import (
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/grammars"
	"github.com/goravel/framework/database/schema/processors"
)

type SqlserverSchema struct {
	contractsschema.CommonSchema

	grammar   *grammars.Sqlserver
	orm       orm.Orm
	prefix    string
	processor processors.Sqlserver
}

func NewSqlserverSchema(grammar *grammars.Sqlserver, orm orm.Orm, prefix string) *SqlserverSchema {
	return &SqlserverSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewSqlserver(),
	}
}

func (r *SqlserverSchema) DropAllTables() error {
	if _, err := r.orm.Query().Exec(r.grammar.CompileDropAllForeignKeys()); err != nil {
		return err
	}

	if _, err := r.orm.Query().Exec(r.grammar.CompileDropAllTables(nil)); err != nil {
		return err
	}

	return nil
}

func (r *SqlserverSchema) DropAllTypes() error {
	return nil
}

func (r *SqlserverSchema) DropAllViews() error {
	_, err := r.orm.Query().Exec(r.grammar.CompileDropAllViews(nil))

	return err
}

func (r *SqlserverSchema) GetIndexes(table string) ([]contractsschema.Index, error) {
	schema, table := r.parseSchemaAndTable(table)
	table = r.prefix + table

	var dbIndexes []processors.DBIndex
	if err := r.orm.Query().Raw(r.grammar.CompileIndexes(schema, table)).Scan(&dbIndexes); err != nil {
		return nil, err
	}

	return r.processor.ProcessIndexes(dbIndexes), nil
}

func (r *SqlserverSchema) GetTypes() ([]contractsschema.Type, error) {
	return nil, nil
}

func (r *SqlserverSchema) parseSchemaAndTable(reference string) (schema, table string) {
	parts := strings.Split(reference, ".")
	if len(parts) == 2 {
		schema = parts[0]
		parts = parts[1:]
	}

	table = parts[0]

	return
}
