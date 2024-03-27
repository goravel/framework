package grammars

import (
	schemacontract "github.com/goravel/framework/contracts/database/schema"
)

type Sqlserver struct{}

func NewSqlserver() *Sqlserver {
	return &Sqlserver{}
}

func (r *Sqlserver) CompileAdd(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileAutoIncrementStartingValues(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileChange(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileColumns(database, table string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileCreate(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileCreateEncoding(sql, connection string, blueprint schemacontract.Blueprint) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileCreateEngine(sql, connection string, blueprint schemacontract.Blueprint) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileCreateTable(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileDrop(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileDropAllTables(tables []string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileDropAllViews(views []string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileDropColumn(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileDropIfExists(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileDropIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileDropPrimary(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileDropUnique(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompilePrimary(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileIndexes(database, table string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileRename(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileRenameColumn(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileRenameIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileTableComment(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileTables(database string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileUnique(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) CompileViews(database string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) ModifyNullable(blueprint schemacontract.Blueprint, column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) ModifyDefault(blueprint schemacontract.Blueprint, column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeBigInteger(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeBinary(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeBoolean(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeChar(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeDate(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeDateTime(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeDateTimeTz(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeDecimal(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeDouble(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeEnum(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeFloat(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeInteger(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeJson(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeJsonb(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeString(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeText(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeTime(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeTimeTz(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeTimestamp(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlserver) TypeTimestampTz(column string) string {
	//TODO implement me
	panic("implement me")
}
