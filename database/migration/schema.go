package migration

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/migration"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/migration/grammars"
)

var _ migration.Schema = (*Schema)(nil)

type Schema struct {
	config     config.Config
	connection string
	grammar    migration.Grammar
	log        log.Log
	migrations []migration.Migration
	orm        contractsorm.Orm
	prefix     string
}

func NewSchema(config config.Config, connection string, log log.Log, orm contractsorm.Orm) *Schema {
	driver := config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))
	prefix := config.GetString(fmt.Sprintf("database.connections.%s.prefix", connection))
	grammar := getGrammar(driver)

	return &Schema{
		config:     config,
		connection: connection,
		grammar:    grammar,
		log:        log,
		orm:        orm,
		prefix:     prefix,
	}
}

func (r *Schema) Connection(name string) migration.Schema {
	return NewSchema(r.config, name, r.log, r.orm)
}

func (r *Schema) Create(table string, callback func(table migration.Blueprint)) error {
	blueprint := r.createBlueprint(table)
	blueprint.Create()
	callback(blueprint)

	// TODO catch error and rollback
	return r.build(blueprint)
}

func (r *Schema) DropIfExists(table string) error {
	blueprint := r.createBlueprint(table)
	blueprint.DropIfExists()

	// TODO catch error when run migrate command
	return r.build(blueprint)
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
		r.log.Errorf("failed to get %s tables: %v", r.connection, err)
		return false
	}

	for _, table := range tables {
		if table.Name == tableName {
			return true
		}
	}

	return false
}

func (r *Schema) Register(migrations []migration.Migration) {
	r.migrations = migrations
}

func (r *Schema) Sql(sql string) {
	// TODO catch error and rollback, optimize test
	_, _ = r.orm.Connection(r.connection).Query().Exec(sql)
}

func (r *Schema) Table(table string, callback func(table migration.Blueprint)) error {
	blueprint := r.createBlueprint(table)
	callback(blueprint)

	// TODO catch error and rollback
	return r.build(blueprint)
}

func (r *Schema) build(blueprint migration.Blueprint) error {
	return blueprint.Build(r.orm.Connection(r.connection).Query(), r.grammar)
}

func (r *Schema) createBlueprint(table string) migration.Blueprint {
	return NewBlueprint(r.prefix, table)
}

func getGrammar(driver string) migration.Grammar {
	switch driver {
	case contractsdatabase.DriverMysql.String():
		// TODO Optimize here when implementing Mysql driver
		return nil
	case contractsdatabase.DriverPostgres.String():
		return grammars.NewPostgres()
	case contractsdatabase.DriverSqlserver.String():
		// TODO Optimize here when implementing Mysql driver
		return nil
	case contractsdatabase.DriverSqlite.String():
		// TODO Optimize here when implementing Mysql driver
		return nil
	default:
		panic(fmt.Sprintf("unsupported database driver: %s", driver))
	}
}
