package migration

import (
	"github.com/goravel/framework/contracts/database/orm"
)

type Grammar interface {
	// CompileCreate Compile a create table command.
	CompileCreate(blueprint Blueprint, query orm.Query) string
	// CompileDropIfExists Compile a drop table (if exists) command.
	CompileDropIfExists(blueprint Blueprint) string
	// CompileTables Compile the query to determine the tables.
	CompileTables(database string) string
	// GetAttributeCommands Get the commands for the schema build.
	GetAttributeCommands() []string
	// GetModifiers Get the column modifiers.
	GetModifiers() []func(Blueprint, ColumnDefinition) string
	// TypeBigInteger Create the column definition for a big integer type.
	TypeBigInteger(column ColumnDefinition) string
	// TypeInteger Create the column definition for an integer type.
	TypeInteger(column ColumnDefinition) string
	// TypeString Create the column definition for a string type.
	TypeString(column ColumnDefinition) string
}
