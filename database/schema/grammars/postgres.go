package grammars

import (
	"fmt"
	"slices"
	"strings"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	schemacontract "github.com/goravel/framework/contracts/database/schema"
)

type Postgres struct {
	attributeCommands []string
	modifiers         []func(schemacontract.Blueprint, schemacontract.ColumnDefinition) string
	serials           []string
}

func NewPostgres() *Postgres {
	postgres := &Postgres{
		attributeCommands: []string{"comment"},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
	}
	postgres.modifiers = []func(schemacontract.Blueprint, schemacontract.ColumnDefinition) string{
		postgres.ModifyDefault,
		postgres.ModifyIncrement,
		postgres.ModifyNullable,
	}

	return postgres
}

func (r *Postgres) CompileAdd(blueprint schemacontract.Blueprint) string {
	return fmt.Sprintf("alter table %s %s", blueprint.GetTableName(), strings.Join(prefixArray("add column", getColumns(r, blueprint)), ","))
}

func (r *Postgres) CompileChange(blueprint schemacontract.Blueprint) string {
	var columns []string
	for _, column := range blueprint.GetChangedColumns() {
		var changes []string

		for _, modifier := range r.modifiers {
			if change := modifier(blueprint, column); change != "" {
				changes = append(changes, change)
			}
		}

		columns = append(columns, strings.Join(prefixArray("alter column "+column.GetName(), changes), ", "))
	}

	if len(columns) == 0 {
		return ""
	}

	return fmt.Sprintf("alter table %s %s", blueprint.GetTableName(), strings.Join(columns, ", "))
}

func (r *Postgres) CompileColumns(database, schema, table string) string {
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
	return fmt.Sprintf("create table %s (%s)", blueprint.GetTableName(), strings.Join(getColumns(r, blueprint), ","))
}

func (r *Postgres) CompileDrop(blueprint schemacontract.Blueprint, command string) string {
	return fmt.Sprintf("drop table %s", blueprint.GetTableName())
}

func (r *Postgres) CompileDropAllTables(tables []string) string {
	return fmt.Sprintf("drop table %s cascade", strings.Join(tables, ","))
}

func (r *Postgres) CompileDropColumn(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	columns := prefixArray("drop column", command.Columns)

	return fmt.Sprintf("alter table %s %s", blueprint.GetTableName(), strings.Join(columns, ","))
}

func (r *Postgres) CompileDropForeign(blueprint schemacontract.Blueprint, index string) string {
	return fmt.Sprintf("alter table %s drop constraint %s", blueprint.GetTableName(), index)
}

func (r *Postgres) CompileDropPrimary(blueprint schemacontract.Blueprint, index string) string {
	tableName := blueprint.GetTableName()

	return fmt.Sprintf("alter table %s drop constraint %s", tableName, tableName+"_pkey")
}

func (r *Postgres) CompileDropIfExists(blueprint schemacontract.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", blueprint.GetTableName())
}

func (r *Postgres) CompileDropIndex(blueprint schemacontract.Blueprint, index string) string {
	return "drop index " + index
}

func (r *Postgres) CompileDropUnique(blueprint schemacontract.Blueprint, command string) string {
	return fmt.Sprintf("alter table %s drop constraint %s", blueprint.GetTableName(), command)
}

func (r *Postgres) CompileForeign(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	sql := fmt.Sprintf("alter table %s add constraint %s foreign key (%s) references %s (%s)",
		blueprint.GetTableName(), command.Index, strings.Join(command.Columns, ", "), command.On, strings.Join(command.References, ", "))
	if command.OnDelete != "" {
		sql += " on delete " + command.OnDelete
	}
	if command.OnUpdate != "" {
		sql += " on update " + command.OnUpdate
	}

	return sql
}

func (r *Postgres) CompileIndex(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = " using " + command.Algorithm
	}

	return fmt.Sprintf("create index %s on %s%s (%s)",
		command.Index,
		blueprint.GetTableName(),
		algorithm,
		strings.Join(command.Columns, ", "),
	)
}

func (r *Postgres) CompileIndexes(schema, table string) string {
	return fmt.Sprintf(
		"select ic.relname as name, string_agg(a.attname, ',' order by indseq.ord) as columns, "+
			`am.amname as "type", i.indisunique as "unique", i.indisprimary as "primary" `+
			"from pg_index i "+
			"join pg_class tc on tc.oid = i.indrelid "+
			"join pg_namespace tn on tn.oid = tc.relnamespace "+
			"join pg_class ic on ic.oid = i.indexrelid "+
			"join pg_am am on am.oid = ic.relam "+
			"join lateral unnest(i.indkey) with ordinality as indseq(num, ord) on true "+
			"left join pg_attribute a on a.attrelid = i.indrelid and a.attnum = indseq.num "+
			"where tc.relname = '%s' and tn.nspname = '%s' "+
			"group by ic.relname, am.amname, i.indisunique, i.indisprimary",
		table,
		schema,
	)
}

func (r *Postgres) CompilePrimary(blueprint schemacontract.Blueprint, columns []string) string {
	return fmt.Sprintf("alter table %s add primary key (%s)",
		blueprint.GetTableName(),
		strings.Join(columns, ", "),
	)
}

func (r *Postgres) CompileRename(blueprint schemacontract.Blueprint, to string) string {
	return fmt.Sprintf("alter table %s rename to %s", blueprint.GetTableName(), blueprint.GetPrefix()+to)
}

func (r *Postgres) CompileRenameColumn(blueprint schemacontract.Blueprint, from, to string) string {
	return fmt.Sprintf("alter table %s rename column %s to %s", blueprint.GetTableName(), from, to)
}

func (r *Postgres) CompileRenameIndex(blueprint schemacontract.Blueprint, from, to string) string {
	return fmt.Sprintf("alter index %s rename to %s", from, to)
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

func (r *Postgres) CompileUnique(blueprint schemacontract.Blueprint, command *schemacontract.Command) string {
	return fmt.Sprintf("alter table %s add constraint %s unique (%s)",
		blueprint.GetTableName(),
		command.Index,
		strings.Join(command.Columns, ", "),
	)
}

func (r *Postgres) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Postgres) GetModifiers() []func(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	return r.modifiers
}

func (r *Postgres) ModifyNullable(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	if column.GetChange() {
		if column.GetNullable() {
			return "drop not null"
		} else {
			return "set not null"
		}
	}

	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Postgres) ModifyDefault(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	if column.GetChange() {
		if !column.GetAutoIncrement() {
			if column.GetDefault() == nil {
				return "drop default"
			} else {
				return fmt.Sprintf("set default %s", getDefaultValue(column.GetDefault()))
			}
		}

		return ""
	}

	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Postgres) ModifyIncrement(blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	if !column.GetChange() && !blueprint.HasCommand("primary") && slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		return " primary key"
	}

	return ""
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
	return fmt.Sprintf("time(%d) without time zone", column.GetPrecision())
}

func (r *Postgres) TypeTimeTz(column schemacontract.ColumnDefinition) string {
	return fmt.Sprintf("time(%d) with time zone", column.GetPrecision())
}

func (r *Postgres) TypeTimestamp(column schemacontract.ColumnDefinition) string {
	return fmt.Sprintf("timestamp(%d) without time zone", column.GetPrecision())
}

func (r *Postgres) TypeTimestampTz(column schemacontract.ColumnDefinition) string {
	return fmt.Sprintf("timestamp(%d) with time zone", column.GetPrecision())
}
