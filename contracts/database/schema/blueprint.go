package schema

import (
	"github.com/goravel/framework/contracts/database/orm"
)

type Blueprint interface {
	// BigIncrements Create a new auto-incrementing big integer (8-byte) column on the table.
	BigIncrements(column string) ColumnDefinition
	// BigInteger Create a new big integer (8-byte) column on the table.
	BigInteger(column string) ColumnDefinition
	// Build Execute the blueprint to build / modify the table.
	Build(query orm.Query, grammar Grammar) error
	// Create Indicate that the table needs to be created.
	Create()
	// DropIfExists Indicate that the table should be dropped if it exists.
	DropIfExists()
	// Foreign Specify a foreign key for the table.
	Foreign(column ...string) ForeignKeyDefinition
	// GetAddedColumns Get the added columns.
	GetAddedColumns() []ColumnDefinition
	// GetCommands Get the commands.
	GetCommands() []*Command
	// GetTableName Get the table name with prefix.
	GetTableName() string
	// HasCommand Determine if the blueprint has a specific command.
	HasCommand(command string) bool
	// ID Create a new auto-incrementing big integer (8-byte) column on the table.
	ID(column ...string) ColumnDefinition
	// Index Specify an index for the table.
	Index(column ...string) IndexDefinition
	// Integer Create a new integer (4-byte) column on the table.
	Integer(column string) ColumnDefinition
	// Primary Specify the primary key(s) for the table.
	Primary(column ...string)
	// SetTable Set the table that the blueprint operates on.
	SetTable(name string)
	// String Create a new string column on the table.
	String(column string, length ...int) ColumnDefinition
	// ToSql Get the raw SQL statements for the blueprint.
	ToSql(grammar Grammar) []string
	// UnsignedBigInteger Create a new unsigned big integer (8-byte) column on the table.
	UnsignedBigInteger(column string) ColumnDefinition
}

type IndexConfig struct {
	Algorithm string
	Name      string
}
