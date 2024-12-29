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
	// Char Create a new char column on the table.
	Char(column string, length ...int) ColumnDefinition
	// Create Indicate that the table needs to be created.
	Create()
	// Date Create a new date column on the table.
	Date(column string) ColumnDefinition
	// DateTime Create a new date-time column on the table.
	DateTime(column string, precision ...int) ColumnDefinition
	// DateTimeTz Create a new date-time column (with time zone) on the table.
	DateTimeTz(column string, precision ...int) ColumnDefinition
	// Decimal Create a new decimal column on the table.
	Decimal(column string) ColumnDefinition
	// Double Create a new double column on the table.
	Double(column string) ColumnDefinition
	// Drop Indicate that the table should be dropped.
	Drop()
	// DropColumn Indicate that the given columns should be dropped.
	DropColumn(column ...string)
	// DropForeign Indicate that the given foreign key should be dropped.
	DropForeign(column ...string)
	// DropForeignByName Indicate that the given foreign key should be dropped.
	DropForeignByName(name string)
	// DropFullText Indicate that the given fulltext index should be dropped.
	DropFullText(column ...string)
	// DropFullTextByName Indicate that the given fulltext index should be dropped.
	DropFullTextByName(name string)
	// DropIfExists Indicate that the table should be dropped if it exists.
	DropIfExists()
	// DropIndex Indicate that the given index should be dropped.
	DropIndex(column ...string)
	// DropIndexByName Indicate that the given index should be dropped.
	DropIndexByName(name string)
	// DropPrimary Indicate that the given primary key should be dropped.
	DropPrimary(column ...string)
	// DropSoftDeletes Indicate that the soft delete column should be dropped.
	DropSoftDeletes(column ...string)
	// DropSoftDeletesTz Indicate that the soft delete column should be dropped.
	DropSoftDeletesTz(column ...string)
	// DropTimestamps Indicate that the timestamp columns should be dropped.
	DropTimestamps()
	// DropTimestampsTz Indicate that the timestamp columns should be dropped.
	DropTimestampsTz()
	// DropUnique Indicate that the given unique key should be dropped.
	DropUnique(column ...string)
	// DropUniqueByName Indicate that the given unique key should be dropped.
	DropUniqueByName(name string)
	// Enum Create a new enum column on the table.
	Enum(column string, array []any) ColumnDefinition
	// Float Create a new float column on the table.
	Float(column string, precision ...int) ColumnDefinition
	// Foreign Specify a foreign key for the table.
	Foreign(column ...string) ForeignKeyDefinition
	// FullText Specify a fulltext for the table.
	FullText(column ...string) IndexDefinition
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
	// Increments Create a new auto-incrementing integer (4-byte) column on the table.
	Increments(column string) ColumnDefinition
	// Index Specify an index for the table.
	Index(column ...string) IndexDefinition
	// Integer Create a new integer (4-byte) column on the table.
	Integer(column string) ColumnDefinition
	// IntegerIncrements Create a new auto-incrementing integer (4-byte) column on the table.
	IntegerIncrements(column string) ColumnDefinition
	// Json Create a new json column on the table.
	Json(column string) ColumnDefinition
	// Jsonb Create a new jsonb column on the table.
	Jsonb(column string) ColumnDefinition
	// LongText Create a new long text column on the table.
	LongText(column string) ColumnDefinition
	// MediumIncrements Create a new auto-incrementing medium integer (3-byte) column on the table.
	MediumIncrements(column string) ColumnDefinition
	// MediumInteger Create a new medium integer (3-byte) column on the table.
	MediumInteger(column string) ColumnDefinition
	// MediumText Create a new medium text column on the table.
	MediumText(column string) ColumnDefinition
	// Primary Specify the primary key(s) for the table.
	Primary(column ...string)
	// Rename the table to a given name.
	Rename(to string)
	// RenameIndex Indicate that the given indexes should be renamed.
	RenameIndex(from, to string)
	// SetTable Set the table that the blueprint operates on.
	SetTable(name string)
	// SmallIncrements Create a new auto-incrementing small integer (2-byte) column on the table.
	SmallIncrements(column string) ColumnDefinition
	// SmallInteger Create a new small integer (2-byte) column on the table.
	SmallInteger(column string) ColumnDefinition
	// SoftDeletes Add a "deleted at" timestamp for the table.
	SoftDeletes(column ...string) ColumnDefinition
	// SoftDeletesTz Add a "deleted at" timestampTz for the table.
	SoftDeletesTz(column ...string) ColumnDefinition
	// String Create a new string column on the table.
	String(column string, length ...int) ColumnDefinition
	// Text Create a new text column on the table.
	Text(column string) ColumnDefinition
	// Time Create a new time column on the table.
	Time(column string, precision ...int) ColumnDefinition
	// TimeTz Create a new time column (with time zone) on the table.
	TimeTz(column string, precision ...int) ColumnDefinition
	// Timestamp Create a new time column on the table.
	Timestamp(column string, precision ...int) ColumnDefinition
	// Timestamps Add nullable creation and update timestamps to the table.
	Timestamps(precision ...int)
	// TimestampsTz Add creation and update timestampTz columns to the table.
	TimestampsTz(precision ...int)
	// TimestampTz Create a new time column (with time zone) on the table.
	TimestampTz(column string, precision ...int) ColumnDefinition
	// TinyIncrements Create a new auto-incrementing tiny integer (1-byte) column on the table.
	TinyIncrements(column string) ColumnDefinition
	// TinyInteger Create a new tiny integer (1-byte) column on the table.
	TinyInteger(column string) ColumnDefinition
	// TinyText Create a new tiny text column on the table.
	TinyText(column string) ColumnDefinition
	// ToSql Get the raw SQL statements for the blueprint.
	ToSql(grammar Grammar) []string
	// Unique Specify a unique index for the table.
	Unique(column ...string) IndexDefinition
	// UnsignedBigInteger Create a new unsigned big integer (8-byte) column on the table.
	UnsignedBigInteger(column string) ColumnDefinition
	// UnsignedInteger Create a new unsigned integer (4-byte) column on the table.
	UnsignedInteger(column string) ColumnDefinition
	// UnsignedMediumInteger Create a new unsigned medium integer (3-byte) column on the table.
	UnsignedMediumInteger(column string) ColumnDefinition
	// UnsignedSmallInteger Create a new unsigned small integer (2-byte) column on the table.
	UnsignedSmallInteger(column string) ColumnDefinition
	// UnsignedTinyInteger Create a new unsigned tiny integer (1-byte) column on the table.
	UnsignedTinyInteger(column string) ColumnDefinition
}
