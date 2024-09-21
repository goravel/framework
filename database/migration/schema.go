package migration

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/migration"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/migration/grammars"
)

var _ migration.Schema = (*Schema)(nil)

type Schema struct {
	blueprint  migration.Blueprint
	config     config.Config
	connection string
	grammar    migration.Grammar
	log        log.Log
	migrations []migration.Migration
	orm        contractsorm.Orm
}

func NewSchema(blueprint migration.Blueprint, config config.Config, connection string, log log.Log, orm contractsorm.Orm) (*Schema, error) {
	driver := config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))
	grammar := getGrammar(driver)

	schema := &Schema{
		blueprint:  blueprint,
		config:     config,
		connection: connection,
		grammar:    grammar,
		log:        log,
		orm:        orm,
	}

	return schema, nil
}

func (r *Schema) Connection(name string) migration.Schema {
	prefix := r.config.GetString(fmt.Sprintf("database.connections.%s.prefix", name))
	dbSchema := r.config.GetString(fmt.Sprintf("database.connections.%s.schema", name))
	blueprint := NewBlueprint(prefix, dbSchema)

	schema, err := NewSchema(blueprint, r.config, name, r.log, r.orm)
	if err != nil {
		r.log.Panic(err)
	}

	return schema
}

func (r *Schema) Create(table string, callback func(table migration.Blueprint)) {
	r.blueprint.SetTable(table)
	r.blueprint.Create()
	callback(r.blueprint)

	// TODO catch error and rollback
	_ = r.blueprint.Build(r.orm.Connection(r.connection).Query(), r.grammar)
}

func (r *Schema) DropIfExists(table string) {
	r.blueprint.SetTable(table)
	r.blueprint.DropIfExists()

	// TODO catch error
	_ = r.blueprint.Build(r.orm.Connection(r.connection).Query(), r.grammar)
}

func (r *Schema) Register(migrations []migration.Migration) {
	r.migrations = migrations
}

func (r *Schema) Sql(sql string) {
	// TODO catch error and rollback, optimize test
	_, _ = r.orm.Connection(r.connection).Query().Exec(sql)
}

func getGrammar(driver string) migration.Grammar {
	switch driver {
	case contractsorm.DriverMysql.String():
		// TODO Optimize here when implementing Mysql driver
		return nil
	case contractsorm.DriverPostgres.String():
		return grammars.NewPostgres()
	case contractsorm.DriverSqlserver.String():
		// TODO Optimize here when implementing Mysql driver
		return nil
	case contractsorm.DriverSqlite.String():
		// TODO Optimize here when implementing Mysql driver
		return nil
	default:
		panic(fmt.Sprintf("unsupported database driver: %s", driver))
	}
}
