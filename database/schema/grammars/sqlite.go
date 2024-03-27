package grammars

import (
	schemacontract "github.com/goravel/framework/contracts/database/schema"
)

type Sqlite struct{}

func NewSqlite() *Sqlite {
	return &Sqlite{}
}

func (r *Sqlite) CompileAdd(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileAutoIncrementStartingValues(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileChange(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileColumns(database, table string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileCreate(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileCreateEncoding(sql, connection string, blueprint schemacontract.Blueprint) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileCreateEngine(sql, connection string, blueprint schemacontract.Blueprint) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileCreateTable(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileDrop(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileDropAllTables(tables []string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileDropAllViews(views []string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileDropColumn(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileDropIfExists(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileDropIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileDropPrimary(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileDropUnique(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompilePrimary(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileIndexes(database, table string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileRename(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileRenameColumn(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileRenameIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileTableComment(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileTables(database string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileUnique(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) CompileViews(database string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) ModifyNullable(blueprint schemacontract.Blueprint, column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) ModifyDefault(blueprint schemacontract.Blueprint, column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeBigInteger(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeBinary(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeBoolean(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeChar(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeDate(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeDateTime(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeDateTimeTz(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeDecimal(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeDouble(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeEnum(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeFloat(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeInteger(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeJson(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeJsonb(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeString(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeText(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeTime(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeTimeTz(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeTimestamp(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Sqlite) TypeTimestampTz(column string) string {
	//TODO implement me
	panic("implement me")
}
