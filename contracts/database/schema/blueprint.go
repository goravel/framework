package schema

import (
	"github.com/goravel/framework/contracts/database/orm"
)

type Blueprint interface {
	// BigIncrements Create a new auto-incrementing big integer (8-byte) column on the table.
	BigIncrements(column string) ColumnDefinition
	// BigInteger Create a new big integer (8-byte) column on the table.
	BigInteger(column string, config ...IntegerConfig) ColumnDefinition
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
	// MediumIncrements Create a new auto-incrementing medium integer (3-byte) column on the table.
	MediumIncrements(column string) ColumnDefinition
	// MediumInteger Create a new medium integer (3-byte) column on the table.
	MediumInteger(column string, config ...IntegerConfig) ColumnDefinition
	// ID Create a new auto-incrementing big integer (8-byte) column on the table.
	ID(column ...string) ColumnDefinition
	// Increments Create a new auto-incrementing integer (4-byte) column on the table.
	Increments(column string) ColumnDefinition
	// Index Specify an index for the table.
	Index(column ...string) IndexDefinition
	// Integer Create a new integer (4-byte) column on the table.
	Integer(column string, config ...IntegerConfig) ColumnDefinition
	// IntegerIncrements Create a new auto-incrementing integer (4-byte) column on the table.
	IntegerIncrements(column string) ColumnDefinition
	// Primary Specify the primary key(s) for the table.
	Primary(column ...string)
	// SetTable Set the table that the blueprint operates on.
	SetTable(name string)
	// SmallIncrements Create a new auto-incrementing small integer (2-byte) column on the table.
	SmallIncrements(column string) ColumnDefinition
	// SmallInteger Create a new small integer (2-byte) column on the table.
	SmallInteger(column string, config ...IntegerConfig) ColumnDefinition
	// String Create a new string column on the table.
	String(column string, length ...int) ColumnDefinition
	// TinyIncrements Create a new auto-incrementing tiny integer (1-byte) column on the table.
	TinyIncrements(column string) ColumnDefinition
	// TinyInteger Create a new tiny integer (1-byte) column on the table.
	TinyInteger(column string, config ...IntegerConfig) ColumnDefinition
	// ToSql Get the raw SQL statements for the blueprint.
	ToSql(grammar Grammar) []string
	// UnsignedBigInteger Create a new unsigned big integer (8-byte) column on the table.
	UnsignedBigInteger(column string, autoIncrement ...bool) ColumnDefinition
	// UnsignedInteger Create a new unsigned integer (4-byte) column on the table.
	UnsignedInteger(column string, autoIncrement ...bool) ColumnDefinition
	// UnsignedMediumInteger Create a new unsigned medium integer (3-byte) column on the table.
	UnsignedMediumInteger(column string, autoIncrement ...bool) ColumnDefinition
	// UnsignedSmallInteger Create a new unsigned small integer (2-byte) column on the table.
	UnsignedSmallInteger(column string, autoIncrement ...bool) ColumnDefinition
	// UnsignedTinyInteger Create a new unsigned tiny integer (1-byte) column on the table.
	UnsignedTinyInteger(column string, autoIncrement ...bool) ColumnDefinition
}

type IndexConfig struct {
	Algorithm string
	Name      string
}
