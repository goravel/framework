package grammars

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	schemacontract "github.com/goravel/framework/contracts/database/schema"
)

type Postgres struct {
	attributeCommands []string
	modifiers         []func(schemacontract.Blueprint, schemacontract.ColumnDefinition) string
}

func NewPostgres() *Postgres {
	postgres := &Postgres{
		attributeCommands: []string{"comment"},
	}
	postgres.modifiers = []func(schemacontract.Blueprint, schemacontract.ColumnDefinition) string{
		postgres.ModifyNullable,
	}

	return postgres
}

func (r *Postgres) CompileAdd(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileAutoIncrementStartingValues(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileChange(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileColumns(database, table, schema string) string {
	return fmt.Sprintf(
		"select a.attname as name, t.typname as type_name, format_type(a.atttypid, a.atttypmod) as type, "+
			"(select tc.collcollate from pg_catalog.pg_collation tc where tc.oid = a.attcollation) as collation, "+
			"not a.attnotnull as nullable, "+
			"(select pg_get_expr(adbin, adrelid) from pg_attrdef where c.oid = pg_attrdef.adrelid and pg_attrdef.adnum = a.attnum) as default, "+
			"col_description(c.oid, a.attnum) as comment "+
			"from pg_attribute a, pg_class c, pg_type t, pg_namespace n "+
			"where c.relname = '%s' and n.nspname = '%s' and a.attnum > 0 and a.attrelid = c.oid and a.atttypid = t.oid and n.oid = c.relnamespace "+
			"order by a.attnum", table, schema)
}

func (r *Postgres) CompileComment(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	return fmt.Sprintf("comment on column %s.%s is '%s'",
		blueprint.GetTableName(),
		command.Column.GetName(),
		strings.ReplaceAll(command.Column.GetComment(), "'", "''"))
}

func (r *Postgres) CompileCreate(blueprint schemacontract.Blueprint, query ormcontract.Query) string {
	return fmt.Sprintf("create table %s (%s)", blueprint.GetTableName(), strings.Join(r.getColumns(blueprint), ","))
}

func (r *Postgres) CompileCreateEncoding(sql, connection string, blueprint schemacontract.Blueprint) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileCreateEngine(sql, connection string, blueprint schemacontract.Blueprint) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileCreateTable(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileDrop(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileDropAllTables(tables []string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileDropAllViews(views []string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileDropColumn(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	columns := prefixArray("drop column", command.Columns)

	return fmt.Sprintf("alter table %s %s", blueprint.GetTableName(), strings.Join(columns, ","))
}

func (r *Postgres) CompileDropIfExists(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileDropIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileDropPrimary(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileDropUnique(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompilePrimary(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileIndexes(database, table string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileRename(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileRenameColumn(blueprint schemacontract.Blueprint, command, connection string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileRenameIndex(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileTableComment(blueprint schemacontract.Blueprint, comment string) string {
	return fmt.Sprintf("comment on table %s is '%s'",
		blueprint.GetTableName(),
		strings.ReplaceAll(comment, "'", "''"))
}

func (r *Postgres) CompileTables(database string) string {
	return "select c.relname as name, n.nspname as schema, pg_total_relation_size(c.oid) as size, " +
		"obj_description(c.oid, 'pg_class') as comment from pg_class c, pg_namespace n " +
		"where c.relkind in ('r', 'p') and n.oid = c.relnamespace and n.nspname not in ('pg_catalog', 'information_schema') " +
		"order by c.relname"
}

func (r *Postgres) CompileUnique(blueprint schemacontract.Blueprint, command string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) CompileViews(database string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Postgres) ModifyNullable(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Postgres) ModifyDefault(blueprint schemacontract.Blueprint, column string) string {
	//TODO implement me
	panic("implement me")
}

func (r *Postgres) TypeBigInteger(column schemacontract.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "bigserial"
	}

	return "bigint"
}

func (r *Postgres) TypeBinary(column schemacontract.ColumnDefinition) string {
	return "bytea"
}

func (r *Postgres) TypeBoolean(column schemacontract.ColumnDefinition) string {
	return "boolean"
}

func (r *Postgres) TypeChar(column schemacontract.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("char(%d)", length)
	}

	return "char"
}

func (r *Postgres) TypeDate(column schemacontract.ColumnDefinition) string {
	return "date"
}

func (r *Postgres) TypeDateTime(column schemacontract.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}

func (r *Postgres) TypeDateTimeTz(column schemacontract.ColumnDefinition) string {
	return r.TypeTimestampTz(column)
}

func (r *Postgres) TypeDecimal(column schemacontract.ColumnDefinition) string {
	return fmt.Sprintf("decimal(%d, %d)", column.GetTotal(), column.GetPlaces())
}

func (r *Postgres) TypeDouble(column schemacontract.ColumnDefinition) string {
	return "double precision"
}

func (r *Postgres) TypeEnum(column schemacontract.ColumnDefinition) string {
	return fmt.Sprintf(`varchar(255) check ("%s" in (%s))`, column.GetName(), strings.Join(quoteString(column.GetAllowed()), ","))
}

func (r *Postgres) TypeFloat(column schemacontract.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("float(%d)", precision)
	}

	return "float"
}

func (r *Postgres) TypeInteger(column schemacontract.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "serial"
	}

	return "integer"
}

func (r *Postgres) TypeJson(column schemacontract.ColumnDefinition) string {
	return "json"
}

func (r *Postgres) TypeJsonb(column schemacontract.ColumnDefinition) string {
	return "jsonb"
}

func (r *Postgres) TypeString(column schemacontract.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}

	return "varchar"
}

func (r *Postgres) TypeText(column schemacontract.ColumnDefinition) string {
	return "text"
}

func (r *Postgres) TypeTime(column schemacontract.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("time(%d) without time zone", precision)
	}

	return "time"
}

func (r *Postgres) TypeTimeTz(column schemacontract.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("time(%d) with time zone", precision)
	}

	return "time"
}

func (r *Postgres) TypeTimestamp(column schemacontract.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("timestamp(%d) without time zone", precision)
	}

	return "timestamp"
}

func (r *Postgres) TypeTimestampTz(column schemacontract.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("timestamp(%d) with time zone", precision)
	}

	return "timestamp"
}

func (r *Postgres) addModify(sql string, blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	for _, modifier := range r.modifiers {
		sql += modifier(blueprint, column)
	}

	return sql
}

func (r *Postgres) getColumns(blueprint schemacontract.Blueprint) []string {
	var columns []string
	for _, column := range blueprint.GetAddedColumns() {
		sql := fmt.Sprintf("%s %s", column.GetName(), r.getType(column))

		columns = append(columns, r.addModify(sql, blueprint, column))
	}

	return columns
}

func (r *Postgres) getType(column schemacontract.ColumnDefinition) string {
	t := []rune(column.GetType())
	t[0] = unicode.ToUpper(t[0])
	methodName := fmt.Sprintf("Type%s", string(t))
	methodValue := reflect.ValueOf(r).MethodByName(methodName)
	if methodValue.IsValid() {
		args := []reflect.Value{reflect.ValueOf(column)}
		callResult := methodValue.Call(args)

		return callResult[0].String()
	}

	return ""
}
