package grammars

import (
	"fmt"
	"strings"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	schemacontract "github.com/goravel/framework/contracts/database/schema"
)

type Mysql struct {
	attributeCommands []string
	modifiers         []func(schemacontract.Blueprint, schemacontract.ColumnDefinition) string
	serials           []string
}

func NewMysql() *Mysql {
	mysql := &Mysql{
		attributeCommands: []string{},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
	}
	mysql.modifiers = []func(schemacontract.Blueprint, schemacontract.ColumnDefinition) string{
		mysql.ModifyCharset,
		mysql.ModifyComment,
		mysql.ModifyDefault,
		mysql.ModifyIncrement,
		mysql.ModifyNullable,
		mysql.ModifyUnsigned,
	}

	return mysql
}

func (r *Mysql) CompileAdd(blueprint schemacontract.Blueprint) string {
	return fmt.Sprintf("alter table %s %s", blueprint.GetTableName(), strings.Join(prefixArray("add column", getColumns(r, blueprint)), ","))
}

func (r *Mysql) CompileChange(blueprint schemacontract.Blueprint) string {
	var columns []string
	for _, column := range blueprint.GetChangedColumns() {
		changes := []string{
			fmt.Sprintf("modify %s %s", column.GetName(), getType(r, column)),
		}

		for _, modifier := range r.modifiers {
			if change := modifier(blueprint, column); change != "" {
				changes = append(changes, change)
			}
		}

		columns = append(columns, strings.Join(changes, ""))
	}

	if len(columns) == 0 {
		return ""
	}

	return fmt.Sprintf("alter table %s %s", blueprint.GetTableName(), strings.Join(columns, ", "))
}

func (r *Mysql) CompileColumns(database, _, table string) string {
	return fmt.Sprintf("select column_name as `name`, data_type as `type_name`, column_type as `type`, "+
		"collation_name as `collation`, is_nullable as `nullable`, "+
		"column_default as `default`, column_comment as `comment`, "+
		"generation_expression as `expression`, extra as `extra` "+
		"from information_schema.columns where table_schema = %s and table_name = %s "+
		"order by ordinal_position asc", database, table)
}

func (r *Mysql) CompileComment(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	return ""
}

func (r *Mysql) CompileCreate(blueprint schemacontract.Blueprint, query ormcontract.Query) string {
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

func (r *Mysql) CompileDropColumn(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropForeign(blueprint schemacontract.Blueprint, index string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropIfExists(blueprint schemacontract.Blueprint) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileDropIndex(blueprint schemacontract.Blueprint, index string) string {
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

func (r *Mysql) CompileForeign(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileIndex(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileIndexes(database, table string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompilePrimary(blueprint schemacontract.Blueprint, columns []string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileRename(blueprint schemacontract.Blueprint, to string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileRenameColumn(blueprint schemacontract.Blueprint, from, to string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileRenameIndex(blueprint schemacontract.Blueprint, from, to string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileTableComment(blueprint schemacontract.Blueprint, comment string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileTables(database string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) CompileUnique(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) GetAttributeCommands() []string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) GetModifiers() []func(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	return r.modifiers
}

func (r *Mysql) ModifyCharset(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	return ""
}

func (r *Mysql) ModifyComment(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	return ""
}

func (r *Mysql) ModifyDefault(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) ModifyNullable(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) ModifyIncrement(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) ModifyUnsigned(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	return ""
}

func (r *Mysql) TypeBigInteger(column schemacontract.ColumnDefinition) string {
	return "bigint"
}

func (r *Mysql) TypeBinary(column schemacontract.ColumnDefinition) string {
	return "blob"
}

func (r *Mysql) TypeBoolean(column schemacontract.ColumnDefinition) string {
	return "tinyint(1)"
}

func (r *Mysql) TypeChar(column schemacontract.ColumnDefinition) string {
	return fmt.Sprintf("char(%d)", column.GetLength())
}

func (r *Mysql) TypeDate(column schemacontract.ColumnDefinition) string {
	return "date"
}

func (r *Mysql) TypeDateTime(column schemacontract.ColumnDefinition) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeDateTimeTz(column schemacontract.ColumnDefinition) string {
	return r.TypeDateTime(column)
}

func (r *Mysql) TypeDecimal(column schemacontract.ColumnDefinition) string {
	return fmt.Sprintf("decimal(%d, %d)", column.GetTotal(), column.GetPlaces())
}

func (r *Mysql) TypeDouble(column schemacontract.ColumnDefinition) string {
	return "double"
}

func (r *Mysql) TypeEnum(column schemacontract.ColumnDefinition) string {
	return fmt.Sprintf(`enum(%s)`, strings.Join(quoteString(column.GetAllowed()), ","))
}

func (r *Mysql) TypeFloat(column schemacontract.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("float(%d)", precision)
	}

	return "float"
}

func (r *Mysql) TypeInteger(column schemacontract.ColumnDefinition) string {
	return "int"
}

func (r *Mysql) TypeJson(column schemacontract.ColumnDefinition) string {
	return "json"
}

func (r *Mysql) TypeJsonb(column schemacontract.ColumnDefinition) string {
	return "json"
}

func (r *Mysql) TypeString(column schemacontract.ColumnDefinition) string {
	return fmt.Sprintf("varchar(%d)", column.GetLength())
}

func (r *Mysql) TypeText(column schemacontract.ColumnDefinition) string {
	return "text"
}

func (r *Mysql) TypeTime(column schemacontract.ColumnDefinition) string {
	//TODO implement me
	if column.GetPrecision() > 0 {
		return fmt.Sprintf("time(%d)", column.GetPrecision())
	} else {
		return "time"
	}
}

func (r *Mysql) TypeTimeTz(column schemacontract.ColumnDefinition) string {
	return r.TypeTime(column)
}

func (r *Mysql) TypeTimestamp(column schemacontract.ColumnDefinition) string {
	//TODO implement me
	panic("implement me")
}

func (r *Mysql) TypeTimestampTz(column schemacontract.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}
