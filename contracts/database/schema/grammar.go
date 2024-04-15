package schema

import (
	ormcontract "github.com/goravel/framework/contracts/database/orm"
)

type Grammar interface {
	// CompileAdd Compile an add column command.
	CompileAdd(blueprint Blueprint, command string) string
	// CompileChange Compile a change column command into a series of SQL statements.
	CompileChange(blueprint Blueprint, command, connection string) string
	// CompileColumns Compile the query to determine the columns.
	// TODO check if the database is required
	CompileColumns(schema, table string) string
	// CompileComment Compile a column comment command.
	CompileComment(blueprint Blueprint, command *Command) string
	// CompileCreate Compile a create table command.
	CompileCreate(blueprint Blueprint, query ormcontract.Query) string
	// CompileDrop Compile a drop table command.
	CompileDrop(blueprint Blueprint, command string) string
	// CompileDropAllTables Compile the SQL needed to drop all tables.
	CompileDropAllTables(tables []string) string
	// CompileDropColumn Compile a drop column command.
	CompileDropColumn(blueprint Blueprint, command *Command) string
	// CompileDropForeign Compile a drop foreign key command.
	CompileDropForeign(blueprint Blueprint, index string) string
	// CompileDropIfExists Compile a drop table (if exists) command.
	CompileDropIfExists(blueprint Blueprint) string
	// CompileDropIndex Compile a drop index command.
	CompileDropIndex(blueprint Blueprint, index string) string
	// CompileDropPrimary Compile a drop primary key command.
	CompileDropPrimary(blueprint Blueprint, command string) string
	// CompileDropUnique Compile a drop unique key command.
	CompileDropUnique(blueprint Blueprint, command string) string
	// CompileForeign Compile a foreign key command.
	CompileForeign(blueprint Blueprint, command *Command) string
	// CompileIndex Compile a plain index key command.
	CompileIndex(blueprint Blueprint, command *Command) string
	// CompileIndexes Compile the query to determine the indexes.
	CompileIndexes(database, table string) string
	// CompilePrimary Compile a primary key command.
	CompilePrimary(blueprint Blueprint, columns []string) string
	// CompileRename Compile a rename table command.
	CompileRename(blueprint Blueprint, to string) string
	// CompileRenameColumn Compile a rename column command.
	CompileRenameColumn(blueprint Blueprint, from, to string) string
	// CompileRenameIndex Compile a rename index command.
	CompileRenameIndex(blueprint Blueprint, from, to string) string
	// CompileTableComment Compile a table comment command.
	CompileTableComment(blueprint Blueprint, comment string) string
	// CompileTables Compile the query to determine the tables.
	CompileTables(database string) string
	// CompileUnique Compile a unique key command.
	CompileUnique(blueprint Blueprint, command *Command) string
	// CompileViews Compile the query to determine the views.
	CompileViews(database string) string
	// GetAttributeCommands Get the commands for the schema build.
	GetAttributeCommands() []string
	// ModifyDefault Get the SQL for a default column modifier.
	ModifyDefault(blueprint Blueprint, column ColumnDefinition) string
	// ModifyNullable Get the SQL for a nullable column modifier.
	ModifyNullable(blueprint Blueprint, column ColumnDefinition) string
	// ModifyIncrement Get the SQL for an auto-increment column modifier.
	ModifyIncrement(blueprint Blueprint, column ColumnDefinition) string
	// TypeBigInteger Create the column definition for a big integer type.
	TypeBigInteger(column ColumnDefinition) string
	// TypeBinary Create the column definition for a binary type.
	TypeBinary(column ColumnDefinition) string
	// TypeBoolean Create the column definition for a boolean type.
	TypeBoolean(column ColumnDefinition) string
	// TypeChar Create the column definition for a char type.
	TypeChar(column ColumnDefinition) string
	// TypeDate Create the column definition for a date type.
	// TODO check if the column is required
	TypeDate(column ColumnDefinition) string
	// TypeDateTime Create the column definition for a date-time type.
	TypeDateTime(column ColumnDefinition) string
	// TypeDateTimeTz Create the column definition for a date-time (with time zone) type.
	TypeDateTimeTz(column ColumnDefinition) string
	// TypeDecimal Create the column definition for a decimal type.
	TypeDecimal(column ColumnDefinition) string
	// TypeDouble Create the column definition for a double type.
	TypeDouble(column ColumnDefinition) string
	// TypeEnum Create the column definition for an enumeration type.
	TypeEnum(column ColumnDefinition) string
	// TypeFloat Create the column definition for a float type.
	TypeFloat(column ColumnDefinition) string
	// TypeInteger Create the column definition for an integer type.
	TypeInteger(column ColumnDefinition) string
	// TypeJson Create the column definition for a json type.
	TypeJson(column ColumnDefinition) string
	// TypeJsonb Create the column definition for a jsonb type.
	TypeJsonb(column ColumnDefinition) string
	// TypeString Create the column definition for a string type.
	TypeString(column ColumnDefinition) string
	// TypeText Create the column definition for a text type.
	TypeText(column ColumnDefinition) string
	// TypeTime Create the column definition for a time type.
	TypeTime(column ColumnDefinition) string
	// TypeTimeTz Create the column definition for a time (with time zone) type.
	TypeTimeTz(column ColumnDefinition) string
	// TypeTimestamp Create the column definition for a timestamp type.
	TypeTimestamp(column ColumnDefinition) string
	// TypeTimestampTz Create the column definition for a timestamp (with time zone) type.
	TypeTimestampTz(column ColumnDefinition) string
}
