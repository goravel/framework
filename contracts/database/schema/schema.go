package schema

import (
	"github.com/goravel/framework/contracts/database/orm"
)

type Schema interface {
	CommonSchema
	DriverSchema
	// Connection Get the connection for the schema.
	Connection(name string) Schema
	// Create a new table on the schema.
	Create(table string, callback func(table Blueprint)) error
	// DropIfExists Drop a table from the schema if exists.
	DropIfExists(table string) error
	// GetConnection Get the connection of the schema.
	GetConnection() string
	// HasTable Determine if the given table exists.
	HasTable(table string) bool
	// Migrations Get the migrations.
	Migrations() []Migration
	// Orm Get the orm instance.
	Orm() orm.Orm
	// Register migrations.
	Register([]Migration)
	// SetConnection Set the connection of the schema.
	SetConnection(name string)
	// Sql Execute a sql directly.
	Sql(sql string)
	// Table Modify a table on the schema.
	Table(table string, callback func(table Blueprint))
}

type CommonSchema interface {
	// GetTables Get the tables that belong to the database.
	GetTables() ([]Table, error)
	// GetViews Get the views that belong to the database.
	GetViews() ([]View, error)
}

type DriverSchema interface {
	// DropAllTables Drop all tables from the schema.
	DropAllTables() error
	// DropAllTypes Drop all types from the schema.
	DropAllTypes() error
	// DropAllViews Drop all views from the schema.
	DropAllViews() error
	// GetTypes Get the types that belong to the database.
	GetTypes() ([]Type, error)
}

type Migration interface {
	// Signature Get the migration signature.
	Signature() string
	// Up Run the migrations.
	Up() error
	// Down Reverse the migrations.
	Down() error
}

type Connection interface {
	// Connection Get the connection for the migration.
	Connection() string
}

type Command struct {
	Algorithm  string
	Column     ColumnDefinition
	Columns    []string
	From       string
	Index      string
	On         string
	OnDelete   string
	OnUpdate   string
	Name       string
	To         string
	References []string
	Value      string
}

type Table struct {
	Comment string
	Name    string
	Schema  string
	Size    int
}

type Type struct {
	Category string
	Implicit bool
	Name     string
	Schema   string
	Type     string
}

type View struct {
	Name       string
	Schema     string
	Definition string
}
