package schema

import (
	"fmt"
	"slices"

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
	schema    string
}

func NewSqlserverSchema(grammar *grammars.Sqlserver, orm orm.Orm, schema, prefix string) *SqlserverSchema {
	return &SqlserverSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewSqlserver(),
		schema:       schema,
	}
}

func (r *SqlserverSchema) DropAllTables() error {
	excludedTables := r.grammar.EscapeNames([]string{"spatial_ref_sys"})
	schema := r.grammar.EscapeNames([]string{r.schema})[0]

	tables, err := r.GetTables()
	if err != nil {
		return err
	}

	var dropTables []string
	for _, table := range tables {
		qualifiedName := fmt.Sprintf("%s.%s", table.Schema, table.Name)

		isExcludedTable := slices.Contains(excludedTables, qualifiedName) || slices.Contains(excludedTables, table.Name)
		isInCurrentSchema := schema == r.grammar.EscapeNames([]string{table.Schema})[0]

		if !isExcludedTable && isInCurrentSchema {
			dropTables = append(dropTables, qualifiedName)
		}
	}

	if len(dropTables) == 0 {
		return nil
	}

	_, err = r.orm.Query().Exec(r.grammar.CompileDropAllTables(dropTables))

	return err
}

func (r *SqlserverSchema) DropAllTypes() error {
	return nil
}

func (r *SqlserverSchema) DropAllViews() error {
	schema := r.grammar.EscapeNames([]string{r.schema})[0]

	views, err := r.GetViews()
	if err != nil {
		return err
	}

	var dropViews []string
	for _, view := range views {
		if schema == view.Schema {
			dropViews = append(dropViews, fmt.Sprintf("%s.%s", view.Schema, view.Name))
		}
	}

	if len(dropViews) == 0 {
		return nil
	}

	_, err = r.orm.Query().Exec(r.grammar.CompileDropAllViews(dropViews))

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
