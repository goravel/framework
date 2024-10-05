package migration

import (
	"github.com/goravel/framework/contracts/database/orm"
)

type Blueprint interface {
	// Build Execute the blueprint to build / modify the table.
	Build(query orm.Query, grammar Grammar) error
	// Create Indicate that the table needs to be created.
	Create()
	// DropIfExists Indicate that the table should be dropped if it exists.
	DropIfExists()
	// GetAddedColumns Get the added columns.
	GetAddedColumns() []ColumnDefinition
	// GetChangedColumns Get the changed columns.
	GetChangedColumns() []ColumnDefinition
	// GetTableName Get the table name with prefix.
	GetTableName() string
	// HasCommand Determine if the blueprint has a specific command.
	HasCommand(command string) bool
	// ID Create a new auto-incrementing big integer (8-byte) column on the table.
	ID(column ...string) ColumnDefinition
	// Integer Create a new integer (4-byte) column on the table.
	Integer(column string) ColumnDefinition
	// SetTable Set the table that the blueprint operates on.
	SetTable(name string)
	// String Create a new string column on the table.
	String(column string, length ...int) ColumnDefinition
	// ToSql Get the raw SQL statements for the blueprint.
	ToSql(query orm.Query, grammar Grammar) []string
}
