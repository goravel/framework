package schema

import (
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/config"
	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/schema/grammars"
	"github.com/goravel/framework/errors"
)

const BindingSchema = "goravel.schema"

var _ contractsschema.Schema = (*Schema)(nil)

type Schema struct {
	contractsschema.CommonSchema
	contractsschema.DriverSchema

	config     config.Config
	grammar    contractsschema.Grammar
	log        log.Log
	migrations []contractsschema.Migration
	orm        contractsorm.Orm
	prefix     string
}

func NewSchema(config config.Config, log log.Log, orm contractsorm.Orm, migrations []contractsschema.Migration) *Schema {
	driver := contractsdatabase.Driver(config.GetString(fmt.Sprintf("database.connections.%s.driver", orm.Name())))
	prefix := config.GetString(fmt.Sprintf("database.connections.%s.prefix", orm.Name()))
	var (
		driverSchema contractsschema.DriverSchema
		grammar      contractsschema.Grammar
	)

	switch driver {
	case contractsdatabase.DriverPostgres:
		schema := config.GetString(fmt.Sprintf("database.connections.%s.search_path", orm.Name()), "public")

		postgresGrammar := grammars.NewPostgres(prefix)
		driverSchema = NewPostgresSchema(postgresGrammar, orm, schema, prefix)
		grammar = postgresGrammar
	case contractsdatabase.DriverMysql:
		mysqlGrammar := grammars.NewMysql(prefix)
		driverSchema = NewMysqlSchema(mysqlGrammar, orm, prefix)
		grammar = mysqlGrammar
	case contractsdatabase.DriverSqlserver:
		sqlserverGrammar := grammars.NewSqlserver(prefix)
		driverSchema = NewSqlserverSchema(sqlserverGrammar, orm, prefix)
		grammar = sqlserverGrammar
	case contractsdatabase.DriverSqlite:
		sqliteGrammar := grammars.NewSqlite(prefix)
		driverSchema = NewSqliteSchema(sqliteGrammar, orm, prefix)
		grammar = sqliteGrammar
	default:
		panic(errors.SchemaDriverNotSupported.Args(driver))
	}

	return &Schema{
		DriverSchema: driverSchema,
		CommonSchema: NewCommonSchema(grammar, orm),

		config:     config,
		grammar:    grammar,
		log:        log,
		migrations: migrations,
		orm:        orm,
		prefix:     prefix,
	}
}

func (r *Schema) Connection(name string) contractsschema.Schema {
	return NewSchema(r.config, r.log, r.orm.Connection(name), r.migrations)
}

func (r *Schema) Create(table string, callback func(table contractsschema.Blueprint)) error {
	blueprint := r.createBlueprint(table)
	blueprint.Create()
	callback(blueprint)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToCreateTable.Args(table, err)
	}

	return nil
}

func (r *Schema) DropIfExists(table string) error {
	blueprint := r.createBlueprint(table)
	blueprint.DropIfExists()

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToDropTable.Args(table, err)
	}

	return nil
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

func (r *Schema) GetConnection() string {
	return r.orm.Name()
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
	tableName := r.prefix + name

	tables, err := r.GetTables()
	if err != nil {
		r.log.Errorf(errors.SchemaFailedToGetTables.Args(r.orm.Name(), err).Error())
		return false
	}

	for _, table := range tables {
		if table.Name == tableName {
			return true
		}
	}

	return false
}

func (r *Schema) Migrations() []contractsschema.Migration {
	return r.migrations
}

func (r *Schema) Orm() contractsorm.Orm {
	return r.orm
}

func (r *Schema) Register(migrations []contractsschema.Migration) {
	r.migrations = migrations
}

func (r *Schema) SetConnection(name string) {
	r.orm = r.orm.Connection(name)
}

func (r *Schema) Sql(sql string) {
	if _, err := r.orm.Query().Exec(sql); err != nil {
		r.log.Fatalf("failed to execute sql: %v", err)
	}
}

func (r *Schema) Table(table string, callback func(table contractsschema.Blueprint)) {
	blueprint := r.createBlueprint(table)
	callback(blueprint)

	if err := r.build(blueprint); err != nil {
		r.log.Fatalf("failed to modify %s table: %v", table, err)
	}
}

func (r *Schema) build(blueprint contractsschema.Blueprint) error {
	if r.orm.Query().InTransaction() {
		return blueprint.Build(r.orm.Query(), r.grammar)
	}

	return r.orm.Transaction(func(tx contractsorm.Query) error {
		return blueprint.Build(tx, r.grammar)
	})
}

func (r *Schema) createBlueprint(table string) contractsschema.Blueprint {
	return NewBlueprint(r.prefix, table)
}
