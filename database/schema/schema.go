package schema

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/schema/grammars"
	"github.com/goravel/framework/errors"
)

var _ schema.Schema = (*Schema)(nil)

type Schema struct {
	schema.CommonSchema
	schema.DriverSchema

	config     config.Config
	grammar    schema.Grammar
	log        log.Log
	migrations []schema.Migration
	orm        contractsorm.Orm
	prefix     string
}

func NewSchema(config config.Config, log log.Log, orm contractsorm.Orm, migrations []schema.Migration) *Schema {
	driver := contractsdatabase.Driver(config.GetString(fmt.Sprintf("database.connections.%s.driver", orm.Name())))
	prefix := config.GetString(fmt.Sprintf("database.connections.%s.prefix", orm.Name()))
	var (
		driverSchema schema.DriverSchema
		grammar      schema.Grammar
	)

	switch driver {
	case contractsdatabase.DriverMysql:
		// TODO Optimize here when implementing Mysql driver
	case contractsdatabase.DriverPostgres:
		postgresGrammar := grammars.NewPostgres()
		driverSchema = NewPostgresSchema(config, postgresGrammar, orm)
		grammar = postgresGrammar
	case contractsdatabase.DriverSqlserver:
		// TODO Optimize here when implementing Mysql driver
	case contractsdatabase.DriverSqlite:
		// TODO Optimize here when implementing Mysql driver
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

func (r *Schema) Connection(name string) schema.Schema {
	return NewSchema(r.config, r.log, r.orm.Connection(name), r.migrations)
}

func (r *Schema) Create(table string, callback func(table schema.Blueprint)) error {
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

func (r *Schema) GetConnection() string {
	return r.orm.Name()
}

func (r *Schema) HasTable(name string) bool {
	blueprint := r.createBlueprint(name)
	tableName := blueprint.GetTableName()

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

func (r *Schema) Migrations() []schema.Migration {
	return r.migrations
}

func (r *Schema) Orm() contractsorm.Orm {
	return r.orm
}

func (r *Schema) Register(migrations []schema.Migration) {
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

func (r *Schema) Table(table string, callback func(table schema.Blueprint)) {
	blueprint := r.createBlueprint(table)
	callback(blueprint)

	if err := r.build(blueprint); err != nil {
		r.log.Fatalf("failed to modify %s table: %v", table, err)
	}
}

func (r *Schema) build(blueprint schema.Blueprint) error {
	return r.orm.Transaction(func(tx contractsorm.Query) error {
		return blueprint.Build(tx, r.grammar)
	})
}

func (r *Schema) createBlueprint(table string) schema.Blueprint {
	return NewBlueprint(r.prefix, table)
}
