package grammars

import (
	schemacontract "github.com/goravel/framework/contracts/database/schema"
)

type Mysql struct{}

func NewMysql() *Mysql {
	return &Mysql{}
}

func (r *Mysql) CompileAdd(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileAutoIncrementStartingValues(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileChange(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileColumns(database, table string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileCreate(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileCreateEncoding(sql, connection string, blueprint schemacontract.Blueprint) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileCreateEngine(sql, connection string, blueprint schemacontract.Blueprint) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileCreateTable(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDrop(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropAllTables(tables []string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropAllViews(views []string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropColumn(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropIfExists(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropPrimary(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropUnique(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompilePrimary(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileIndexes(database, table string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileRename(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileRenameColumn(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileRenameIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileTableComment(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileTables(database string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileUnique(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileViews(database string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) ModifyNullable(blueprint schemacontract.Blueprint, column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) ModifyDefault(blueprint schemacontract.Blueprint, column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeBigInteger(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeBinary(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeBoolean(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeChar(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeDate(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeDateTime(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeDateTimeTz(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeDecimal(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeDouble(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeEnum(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeFloat(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeInteger(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeJson(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeJsonb(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeString(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeText(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeTime(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeTimeTz(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeTimestamp(column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeTimestampTz(column string) string {
	//TODO implement me
	panic("implement me")
}
