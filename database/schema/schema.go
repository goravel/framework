package schema

import (
	"fmt"
	"slices"
	"strings"

	configcontract "github.com/goravel/framework/contracts/config"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	schemacontract "github.com/goravel/framework/contracts/database/schema"
	logcontract "github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/schema/grammars"
	"github.com/goravel/framework/database/schema/processors"
)

type DB struct {
	Connection string
	Database   string
	Prefix     string
	Schema     string
}

type Schema struct {
	config     configcontract.Config
	connection string
	db         DB
	grammar    schemacontract.Grammar
	log        logcontract.Log
	migrations []schemacontract.Migration
	orm        ormcontract.Orm
	processor  schemacontract.Processor
	query      ormcontract.Query
}

func NewSchema(connection string, config configcontract.Config, orm ormcontract.Orm, log logcontract.Log) (*Schema, error) {
	if connection == "" {
		connection = config.GetString("database.default")
	}

	schema := &Schema{
		config:     config,
		connection: connection,
		db: DB{
			Connection: connection,
			Database:   config.GetString(fmt.Sprintf("database.connections.%s.database", connection)),
			Prefix:     config.GetString(fmt.Sprintf("database.connections.%s.prefix", connection)),
			Schema:     config.GetString(fmt.Sprintf("database.connections.%s.schema", connection)),
		},
		log:   log,
		orm:   orm,
		query: orm.Connection(connection).Query(),
	}

	if err := schema.initGrammarAndProcess(); err != nil {
		return nil, err
	}

	return schema, nil
}

func (r *Schema) Connection(name string) schemacontract.Schema {
	schema, err := NewSchema(name, r.config, r.orm, r.log)
	if err != nil {
		panic(err)
	}

	return schema
}

func (r *Schema) Create(table string, callback func(table schemacontract.Blueprint)) error {
	blueprint := NewBlueprint(r.db.Prefix, table)
	blueprint.Create()
	callback(blueprint)

	return blueprint.Build(r.query, r.grammar)
}

func (r *Schema) Drop(table string) error {
	blueprint := NewBlueprint(r.db.Prefix, table)
	blueprint.Drop()

	return blueprint.Build(r.query, r.grammar)
}

func (r *Schema) DropAllTables() error {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) DropAllViews() error {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) DropColumns(table string, columns []string) error {
	return r.Table(table, func(table schemacontract.Blueprint) {
		table.DropColumn(columns...)
	})
}

func (r *Schema) DropIfExists(table string) error {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) GetColumns(table string) ([]schemacontract.Column, error) {
	_, schema, table := r.parseDatabaseAndSchemaAndTable(table)
	table = r.db.Prefix + table

	var columns []schemacontract.Column
	if err := r.query.Raw(r.grammar.CompileColumns(schema, table)).Scan(&columns); err != nil {
		return nil, err
	}

	return r.processor.ProcessColumns(columns), nil
}

func (r *Schema) GetColumnListing(table string) []string {
	columns, err := r.GetColumns(table)
	if err != nil {
		r.log.Errorf("failed to get %s columns: %v", table, err)
		return nil
	}

	var names []string
	for _, column := range columns {
		names = append(names, column.Name)
	}

	return names
}

func (r *Schema) GetIndexes(table string) ([]schemacontract.Index, error) {
	_, schema, table := r.parseDatabaseAndSchemaAndTable(table)
	table = r.db.Prefix + table

	type Index struct {
		Columns string
		Name    string
		Primary bool
		Type    string
		Unique  bool
	}

	var tempIndexes []Index
	if err := r.query.Raw(r.grammar.CompileIndexes(schema, table)).Scan(&tempIndexes); err != nil {
		return nil, err
	}

	var indexes []schemacontract.Index
	for _, tempIndex := range tempIndexes {
		indexes = append(indexes, schemacontract.Index{
			Columns: strings.Split(tempIndex.Columns, ","),
			Name:    tempIndex.Name,
			Primary: tempIndex.Primary,
			Type:    tempIndex.Type,
			Unique:  tempIndex.Unique,
		})
	}

	return r.processor.ProcessIndexes(indexes), nil
}

func (r *Schema) GetIndexListing(table string) []string {
	indexes, err := r.GetIndexes(table)
	if err != nil {
		r.log.Errorf("failed to get %s indexes: %v", table, err)
		return nil
	}

	var names []string
	for _, index := range indexes {
		names = append(names, index.Name)
	}

	return names
}

func (r *Schema) GetTableListing() []string {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) GetTables() ([]schemacontract.Table, error) {
	var tables []schemacontract.Table
	if err := r.query.Raw(r.grammar.CompileTables("")).Scan(&tables); err != nil {
		return nil, err
	}

	return tables, nil
}

func (r *Schema) GetViews() []schemacontract.View {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) HasColumn(table, column string) bool {
	return slices.Contains(r.GetColumnListing(table), column)
}

func (r *Schema) HasColumns(table string, columns []string) bool {
	columnListing := r.GetColumnListing(table)
	for _, column := range columns {
		if !slices.Contains(columnListing, column) {
			return false
		}
	}

	return true
}

func (r *Schema) HasIndex(table, index string) bool {
	indexListing := r.GetIndexListing(table)

	return slices.Contains(indexListing, index)
}

func (r *Schema) HasTable(name string) bool {
	tables, err := r.GetTables()
	if err != nil {
		r.log.Errorf("failed to get %s tables: %v", r.connection, err)
		return false
	}

	for _, table := range tables {
		if table.Name == name {
			return true
		}
	}

	return false
}

func (r *Schema) HasView(view string) bool {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) Register(migrations []schemacontract.Migration) {
	r.migrations = append(r.migrations, migrations...)
}

func (r *Schema) Rename(from, to string) {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) Table(table string, callback func(table schemacontract.Blueprint)) error {
	blueprint := NewBlueprint(r.db.Prefix, table)
	callback(blueprint)

	return blueprint.Build(r.query, r.grammar)
}

func (r *Schema) initGrammarAndProcess() error {
	switch r.query.Driver() {
	//case ormcontract.DriverMysql:
	//	grammar = grammars.NewMysql()
	case ormcontract.DriverPostgres:
		r.grammar = grammars.NewPostgres()
		r.processor = processors.NewPostgres()
		return nil
	//case ormcontract.DriverSqlserver:
	//	grammar = grammars.NewSqlserver()
	//case ormcontract.DriverSqlite:
	//	grammar = grammars.NewSqlite()
	default:
		return fmt.Errorf("unsupported database driver: %s", r.query.Driver())
	}
}

// parseSchemaAndTable Parse the database object reference and extract the database, schema, and table.
func (r *Schema) parseDatabaseAndSchemaAndTable(reference string) (database, schema, table string) {
	parts := strings.Split(reference, ".")
	database = r.db.Database
	schema = r.db.Schema
	if schema == "" {
		schema = "public"
	}

	if len(parts) == 2 {
		schema = parts[0]
		parts = parts[1:]
	}

	table = parts[0]

	return
}
