package schema

import (
	"github.com/goravel/framework/contracts/database/orm"
)

type Schema interface {
	DriverSchema
	// Connection Get the connection for the schema.
	Connection(name string) Schema
	// Create a new table on the schema.
	Create(table string, callback func(table Blueprint)) error
	// Drop a table from the schema.
	Drop(table string) error
	// DropColumns Drop columns from a table on the schema.
	DropColumns(table string, columns []string) error
	// DropIfExists Drop a table from the schema if exists.
	DropIfExists(table string) error
	// GetColumnListing Get the column listing for a given table.
	GetColumnListing(table string) []string
	// GetConnection Get the connection of the schema.
	GetConnection() string
	// GetForeignKeys Get the foreign keys for a given table.
	GetForeignKeys(table string) ([]ForeignKey, error)
	// GetIndexListing Get the names of the indexes for a given table.
	GetIndexListing(table string) []string
	// GetTableListing Get the table listing for the database.
	GetTableListing() []string
	// HasColumn Determine if the given table has a given column.
	HasColumn(table, column string) bool
	// HasColumns Determine if the given table has given columns.
	HasColumns(table string, columns []string) bool
	// HasIndex Determine if the given table has a given index.
	HasIndex(table, index string) bool
	// HasTable Determine if the given table exists.
	HasTable(name string) bool
	// HasType Determine if the given type exists.
	HasType(name string) bool
	// HasView Determine if the given view exists.
	HasView(name string) bool
	// Migrations Get the migrations.
	Migrations() []Migration
	// Orm Get the orm instance.
	Orm() orm.Orm
	// Register migrations.
	Register([]Migration)
	// Rename a table on the schema.
	Rename(from, to string) error
	// SetConnection Set the connection of the schema.
	SetConnection(name string)
	// Sql Execute a sql directly.
	Sql(sql string) error
	// Table Modify a table on the schema.
	Table(table string, callback func(table Blueprint)) error
}

type CommonSchema interface {
	// GetTables Get the tables that belong to the database.
	GetTables() ([]Table, error)
	// GetViews Get the views that belong to the database.
	GetViews() ([]View, error)
}

// TODO To Check if this can be removed or reduced
type DriverSchema interface {
	// DropAllTables Drop all tables from the schema.
	DropAllTables() error
	// DropAllTypes Drop all types from the schema.
	DropAllTypes() error
	// DropAllViews Drop all views from the schema.
	DropAllViews() error
	// GetColumns Get the columns for a given table.
	GetColumns(table string) ([]Column, error)
	// GetIndexes Get the indexes for a given table.
	GetIndexes(table string) ([]Index, error)
	// GetTables Get the tables that belong to the database.
	GetTables() ([]Table, error)
	// GetTypes Get the types that belong to the database.
	GetTypes() ([]Type, error)
	// GetViews Get the views that belong to the database.
	GetViews() ([]View, error)
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
	Algorithm          string
	Column             ColumnDefinition
	Columns            []string
	Deferrable         *bool
	From               string
	Index              string
	InitiallyImmediate *bool
	Language           string
	Name               string
	On                 string
	OnDelete           string
	OnUpdate           string
	References         []string
	ShouldBeSkipped    bool
	To                 string
	Value              string
}

type ForeignKey struct {
	Name           string
	Columns        []string
	ForeignSchema  string
	ForeignTable   string
	ForeignColumns []string
	OnUpdate       string
	OnDelete       string
}

type Index struct {
	Columns []string
	Name    string
	Primary bool
	Type    string
	Unique  bool
}

type Table struct {
	Collation string
	Comment   string
	Engine    string
	Name      string
	Schema    string
	Size      int
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
