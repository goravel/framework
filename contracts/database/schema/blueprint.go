package schema

import (
	ormcontract "github.com/goravel/framework/contracts/database/orm"
)

type Blueprint interface {
	// Boolean Create a new boolean column on the table.
	Boolean(column string) ColumnDefinition
	// BigInteger Create a new big integer (8-byte) column on the table.
	BigInteger(column string) ColumnDefinition
	// Binary Create a new binary column on the table.
	Binary(column string) ColumnDefinition
	// Build Execute the blueprint to build / modify the table.
	Build(query ormcontract.Query, grammar Grammar) error
	// Char Create a new char column on the table.
	Char(column string, length ...int) ColumnDefinition
	// Comment Add a comment to the table.
	Comment(comment string)
	// Create Indicate that the table needs to be created.
	Create()
	// Date Create a new date column on the table.
	Date(column string) ColumnDefinition
	// DateTime Create a new date-time column on the table.
	DateTime(column string) ColumnDefinition
	// DateTimeTz Create a new date-time column (with time zone) on the table.
	DateTimeTz(column string) ColumnDefinition
	// Decimal Create a new decimal column on the table.
	Decimal(column string) ColumnDefinition
	// Double Create a new double column on the table.
	Double(column string) ColumnDefinition
	// DropColumn Indicate that the given columns should be dropped.
	DropColumn(column string) error
	// DropForeign Indicate that the given foreign key should be dropped.
	DropForeign(index string) error
	// DropIndex Indicate that the given index should be dropped.
	DropIndex(index string) error
	// DropSoftDeletes Indicate that the soft delete column should be dropped.
	DropSoftDeletes() error
	// DropTimestamps Indicate that the timestamp columns should be dropped.
	DropTimestamps() error
	// Enum Create a new enum column on the table.
	Enum(column string, array []any) ColumnDefinition
	// Float Create a new float column on the table.
	Float(column string) ColumnDefinition
	// Foreign Specify a foreign key for the table.
	Foreign(columns []string, name ...string) error
	// GetAddedColumns Get the added columns.
	GetAddedColumns() []ColumnDefinition
	// GetChangedColumns Get the changed columns.
	GetChangedColumns() []ColumnDefinition
	// GetTableName Get the table name with prefix.
	GetTableName() string
	// ID Create a new auto-incrementing big integer (8-byte) column on the table.
	ID() ColumnDefinition
	// Index Specify an index for the table.
	Index(columns []string, name string) error
	// Integer Create a new integer (4-byte) column on the table.
	Integer(column string) ColumnDefinition
	// Json Create a new json column on the table.
	Json(column string) ColumnDefinition
	// Jsonb Create a new jsonb column on the table.
	Jsonb(column string) ColumnDefinition
	// Primary Specify the primary key(s) for the table.
	Primary(columns []string, name string) error
	// RenameColumn Indicate that the given columns should be renamed.
	RenameColumn(from, to string) error
	// RenameIndex Indicate that the given indexes should be renamed.
	RenameIndex(from, to string) error
	// SoftDeletes Add a "deleted at" timestamp for the table.
	SoftDeletes(column ...string) ColumnDefinition
	// SoftDeletesTz Add a "deleted at" timestampTz for the table.
	SoftDeletesTz(column ...string) ColumnDefinition
	// String Create a new string column on the table.
	String(column string, length ...int) ColumnDefinition
	// Text Create a new text column on the table.
	Text(column string) ColumnDefinition
	// Time Create a new time column on the table.
	Time(column string) ColumnDefinition
	// TimeTz Create a new time column (with time zone) on the table.
	TimeTz(column string) ColumnDefinition
	// Timestamp Create a new time column on the table.
	Timestamp(column string) ColumnDefinition
	// Timestamps Add nullable creation and update timestamps to the table.
	Timestamps() ColumnDefinition
	// TimestampsTz Add creation and update timestampTz columns to the table.
	TimestampsTz() ColumnDefinition
	// TimestampTz Create a new time column (with time zone) on the table.
	TimestampTz(column string) ColumnDefinition
	// ToSql Get the raw SQL statements for the blueprint.
	ToSql(query ormcontract.Query, grammar Grammar) []string
	// Unique Specify a unique index for the table.
	Unique(columns []string, name string) error
	// UnsignedInteger Create a new unsigned integer (4-byte) column on the table.
	UnsignedInteger(column string) ColumnDefinition
	// UnsignedBigInteger Create a new unsigned big integer (8-byte) column on the table.
	UnsignedBigInteger(column string) ColumnDefinition
}
