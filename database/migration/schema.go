package migration

import (
	"fmt"
	"os"

	"github.com/goravel/framework/contracts/config"
	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/migration"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/migration/grammars"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

var _ migration.Schema = (*Schema)(nil)

type Schema struct {
	config     config.Config
	grammar    migration.Grammar
	log        log.Log
	migrations []migration.Migration
	orm        contractsorm.Orm
	prefix     string
}

func NewSchema(config config.Config, log log.Log, orm contractsorm.Orm, migrations []migration.Migration) *Schema {
	driver := contractsdatabase.Driver(config.GetString(fmt.Sprintf("database.connections.%s.driver", orm.Name())))
	prefix := config.GetString(fmt.Sprintf("database.connections.%s.prefix", orm.Name()))
	grammar := getGrammar(driver)

	return &Schema{
		config:     config,
		grammar:    grammar,
		log:        log,
		migrations: migrations,
		orm:        orm,
		prefix:     prefix,
	}
}

func (r *Schema) Connection(name string) migration.Schema {
	return NewSchema(r.config, r.log, r.orm.Connection(name), r.migrations)
}

func (r *Schema) Create(table string, callback func(table migration.Blueprint)) {
	blueprint := r.createBlueprint(table)
	blueprint.Create()
	callback(blueprint)

	if err := r.build(blueprint); err != nil {
		color.Red().Printf("failed to create %s table: %v\n", table, err)
		os.Exit(1)
	}
}

func (r *Schema) DropIfExists(table string) {
	blueprint := r.createBlueprint(table)
	blueprint.DropIfExists()

	if err := r.build(blueprint); err != nil {
		color.Red().Printf("failed to drop %s table: %v\n", table, err)
		os.Exit(1)
	}
}

func (r *Schema) GetConnection() string {
	return r.orm.Name()
}

func (r *Schema) GetTables() ([]migration.Table, error) {
	var tables []migration.Table
	if err := r.orm.Query().Raw(r.grammar.CompileTables("")).Scan(&tables); err != nil {
		return nil, err
	}

	return tables, nil
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

func (r *Schema) Migrations() []migration.Migration {
	return r.migrations
}

func (r *Schema) Orm() contractsorm.Orm {
	return r.orm
}

func (r *Schema) Register(migrations []migration.Migration) {
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

func (r *Schema) Table(table string, callback func(table migration.Blueprint)) {
	blueprint := r.createBlueprint(table)
	callback(blueprint)

	if err := r.build(blueprint); err != nil {
		r.log.Fatalf("failed to modify %s table: %v", table, err)
	}
}

func (r *Schema) build(blueprint migration.Blueprint) error {
	return r.orm.Transaction(func(tx contractsorm.Query) error {
		return blueprint.Build(tx, r.grammar)
	})
}

func (r *Schema) createBlueprint(table string) migration.Blueprint {
	return NewBlueprint(r.prefix, table)
}

func getGrammar(driver contractsdatabase.Driver) migration.Grammar {
	switch driver {
	case contractsdatabase.DriverMysql:
		// TODO Optimize here when implementing Mysql driver
		return nil
	case contractsdatabase.DriverPostgres:
		return grammars.NewPostgres()
	case contractsdatabase.DriverSqlserver:
		// TODO Optimize here when implementing Mysql driver
		return nil
	case contractsdatabase.DriverSqlite:
		// TODO Optimize here when implementing Mysql driver
		return nil
	default:
		panic(errors.SchemaDriverNotSupported.Args(driver))
	}
}
