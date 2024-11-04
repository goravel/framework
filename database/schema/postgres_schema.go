package schema

import (
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/grammars"
	"github.com/goravel/framework/database/schema/processors"
)

type PostgresSchema struct {
	contractsschema.CommonSchema

	config    config.Config
	grammar   *grammars.Postgres
	orm       orm.Orm
	processor processors.Postgres
}

func NewPostgresSchema(config config.Config, grammar *grammars.Postgres, orm orm.Orm) *PostgresSchema {
	return &PostgresSchema{
		CommonSchema: NewCommonSchema(grammar, orm),

		config:    config,
		grammar:   grammar,
		orm:       orm,
		processor: processors.NewPostgres(),
	}
}

func (r *PostgresSchema) DropAllTables() error {
	excludedTables := r.grammar.EscapeNames([]string{"spatial_ref_sys"})
	schema := r.grammar.EscapeNames([]string{r.getSchema()})[0]

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
	schema := r.grammar.EscapeNames([]string{r.getSchema()})[0]
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
	schema := r.grammar.EscapeNames([]string{r.getSchema()})[0]

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

func (r *PostgresSchema) GetTypes() ([]contractsschema.Type, error) {
	var types []contractsschema.Type
	if err := r.orm.Query().Raw(r.grammar.CompileTypes()).Scan(&types); err != nil {
		return nil, err
	}

	return r.processor.ProcessTypes(types), nil
}

func (r *PostgresSchema) getSchema() string {
	schema := r.config.GetString(fmt.Sprintf("database.connections.%s.search_path", r.orm.Name()))
	if schema == "" {
		return "public"
	}

	return schema
}
