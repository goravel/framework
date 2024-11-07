package schema

import (
	"fmt"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/grammars"
	"github.com/goravel/framework/database/schema/processors"
)

type PostgresSchema struct {
	contractsschema.CommonSchema

	grammar   *grammars.Postgres
	orm       orm.Orm
	prefix    string
	processor processors.Postgres
	schema    string
}

func NewPostgresSchema(grammar *grammars.Postgres, orm orm.Orm, schema, prefix string) *PostgresSchema {
	return &PostgresSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewPostgres(),
		schema:       schema,
	}
}

func (r *PostgresSchema) DropAllTables() error {
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

func (r *PostgresSchema) DropAllTypes() error {
	schema := r.grammar.EscapeNames([]string{r.schema})[0]
	types, err := r.GetTypes()
	if err != nil {
		return err
	}

	var dropTypes, dropDomains []string

	for _, t := range types {
		if !t.Implicit && schema == t.Schema {
			if t.Type == "domain" {
				dropDomains = append(dropDomains, fmt.Sprintf("%s.%s", t.Schema, t.Name))
			} else {
				dropTypes = append(dropTypes, fmt.Sprintf("%s.%s", t.Schema, t.Name))
			}
		}
	}

	if len(dropTypes) > 0 {
		if _, err := r.orm.Query().Exec(r.grammar.CompileDropAllTypes(dropTypes)); err != nil {
			return err
		}
	}

	if len(dropDomains) > 0 {
		if _, err := r.orm.Query().Exec(r.grammar.CompileDropAllDomains(dropDomains)); err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresSchema) DropAllViews() error {
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

func (r *PostgresSchema) GetIndexes(table string) ([]contractsschema.Index, error) {
	schema, table := r.parseSchemaAndTable(table)
	table = r.prefix + table

	type Index struct {
		Columns string
		Name    string
		Primary bool
		Type    string
		Unique  bool
	}

	var tempIndexes []Index
	if err := r.orm.Query().Raw(r.grammar.CompileIndexes(schema, table)).Scan(&tempIndexes); err != nil {
		return nil, err
	}

	var indexes []contractsschema.Index
	for _, tempIndex := range tempIndexes {
		indexes = append(indexes, contractsschema.Index{
			Columns: strings.Split(tempIndex.Columns, ","),
			Name:    tempIndex.Name,
			Primary: tempIndex.Primary,
			Type:    tempIndex.Type,
			Unique:  tempIndex.Unique,
		})
	}

	return r.processor.ProcessIndexes(indexes), nil
}

func (r *PostgresSchema) GetTypes() ([]contractsschema.Type, error) {
	var types []contractsschema.Type
	if err := r.orm.Query().Raw(r.grammar.CompileTypes()).Scan(&types); err != nil {
		return nil, err
	}

	return r.processor.ProcessTypes(types), nil
}

func (r *PostgresSchema) parseSchemaAndTable(reference string) (schema, table string) {
	parts := strings.Split(reference, ".")
	schema = r.schema
	if len(parts) == 2 {
		schema = parts[0]
		parts = parts[1:]
	}

	table = parts[0]

	return
}
