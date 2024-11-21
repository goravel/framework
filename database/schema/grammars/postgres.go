package grammars

import (
	"fmt"
	"slices"
	"strings"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/constants"
)

type Postgres struct {
	attributeCommands []string
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
	wrap              *Wrap
}

func NewPostgres(tablePrefix string) *Postgres {
	postgres := &Postgres{
		attributeCommands: []string{constants.CommandComment},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		wrap:              NewWrap(contractsdatabase.DriverPostgres, tablePrefix),
	}
	postgres.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		postgres.ModifyDefault,
		postgres.ModifyIncrement,
		postgres.ModifyNullable,
	}

	return postgres
}

func (r *Postgres) CompileAdd(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s add column %s", r.wrap.Table(blueprint.GetTableName()), r.getColumn(blueprint, command.Column))
}

func (r *Postgres) CompileColumns(schema, table string) string {
	return fmt.Sprintf(
		"select a.attname as name, t.typname as type_name, format_type(a.atttypid, a.atttypmod) as type, "+
			"(select tc.collcollate from pg_catalog.pg_collation tc where tc.oid = a.attcollation) as collation, "+
			"not a.attnotnull as nullable, "+
			"(select pg_get_expr(adbin, adrelid) from pg_attrdef where c.oid = pg_attrdef.adrelid and pg_attrdef.adnum = a.attnum) as default, "+
			"col_description(c.oid, a.attnum) as comment "+
			"from pg_attribute a, pg_class c, pg_type t, pg_namespace n "+
			"where c.relname = %s and n.nspname = %s and a.attnum > 0 and a.attrelid = c.oid and a.atttypid = t.oid and n.oid = c.relnamespace "+
			"order by a.attnum", r.wrap.Quote(table), r.wrap.Quote(schema))
}

func (r *Postgres) CompileComment(blueprint schema.Blueprint, command *schema.Command) string {
	comment := "NULL"
	if command.Column.IsSetComment() {
		comment = r.wrap.Quote(strings.ReplaceAll(command.Column.GetComment(), "'", "''"))
	}

	return fmt.Sprintf("comment on column %s.%s is %s",
		r.wrap.Table(blueprint.GetTableName()),
		r.wrap.Column(command.Column.GetName()),
		comment)
}

func (r *Postgres) CompileCreate(blueprint schema.Blueprint) string {
	return fmt.Sprintf("create table %s (%s)", r.wrap.Table(blueprint.GetTableName()), strings.Join(r.getColumns(blueprint), ", "))
}

func (r *Postgres) CompileDropAllDomains(domains []string) string {
	return fmt.Sprintf("drop domain %s cascade", strings.Join(r.EscapeNames(domains), ", "))
}

func (r *Postgres) CompileDropAllTables(tables []string) string {
	return fmt.Sprintf("drop table %s cascade", strings.Join(r.EscapeNames(tables), ", "))
}

func (r *Postgres) CompileDropAllTypes(types []string) string {
	return fmt.Sprintf("drop type %s cascade", strings.Join(r.EscapeNames(types), ", "))
}

func (r *Postgres) CompileDropAllViews(views []string) string {
	return fmt.Sprintf("drop view %s cascade", strings.Join(r.EscapeNames(views), ", "))
}

func (r *Postgres) CompileDropIfExists(blueprint schema.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", r.wrap.Table(blueprint.GetTableName()))
}

func (r *Postgres) CompileForeign(blueprint schema.Blueprint, command *schema.Command) string {
	sql := fmt.Sprintf("alter table %s add constraint %s foreign key (%s) references %s (%s)",
		r.wrap.Table(blueprint.GetTableName()),
		r.wrap.Column(command.Index),
		r.wrap.Columnize(command.Columns),
		r.wrap.Table(command.On),
		r.wrap.Columnize(command.References))
	if command.OnDelete != "" {
		sql += " on delete " + command.OnDelete
	}
	if command.OnUpdate != "" {
		sql += " on update " + command.OnUpdate
	}

	return sql
}

func (r *Postgres) CompileIndex(blueprint schema.Blueprint, command *schema.Command) string {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = " using " + command.Algorithm
	}

	return fmt.Sprintf("create index %s on %s%s (%s)",
		r.wrap.Column(command.Index),
		r.wrap.Table(blueprint.GetTableName()),
		algorithm,
		r.wrap.Columnize(command.Columns),
	)
}

func (r *Postgres) CompileIndexes(schema, table string) string {
	return fmt.Sprintf(
		"select ic.relname as name, string_agg(a.attname, ',' order by indseq.ord) as columns, "+
			"am.amname as \"type\", i.indisunique as \"unique\", i.indisprimary as \"primary\" "+
			"from pg_index i "+
			"join pg_class tc on tc.oid = i.indrelid "+
			"join pg_namespace tn on tn.oid = tc.relnamespace "+
			"join pg_class ic on ic.oid = i.indexrelid "+
			"join pg_am am on am.oid = ic.relam "+
			"join lateral unnest(i.indkey) with ordinality as indseq(num, ord) on true "+
			"left join pg_attribute a on a.attrelid = i.indrelid and a.attnum = indseq.num "+
			"where tc.relname = %s and tn.nspname = %s "+
			"group by ic.relname, am.amname, i.indisunique, i.indisprimary",
		r.wrap.Quote(table),
		r.wrap.Quote(schema),
	)
}

func (r *Postgres) CompilePrimary(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s add primary key (%s)", r.wrap.Table(blueprint.GetTableName()), r.wrap.Columnize(command.Columns))
}

func (r *Postgres) CompileTables(database string) string {
	return "select c.relname as name, n.nspname as schema, pg_total_relation_size(c.oid) as size, " +
		"obj_description(c.oid, 'pg_class') as comment from pg_class c, pg_namespace n " +
		"where c.relkind in ('r', 'p') and n.oid = c.relnamespace and n.nspname not in ('pg_catalog', 'information_schema') " +
		"order by c.relname"
}

func (r *Postgres) CompileTypes() string {
	return `select t.typname as name, n.nspname as schema, t.typtype as type, t.typcategory as category, 
		((t.typinput = 'array_in'::regproc and t.typoutput = 'array_out'::regproc) or t.typtype = 'm') as implicit 
		from pg_type t 
		join pg_namespace n on n.oid = t.typnamespace 
		left join pg_class c on c.oid = t.typrelid 
		left join pg_type el on el.oid = t.typelem 
		left join pg_class ce on ce.oid = el.typrelid 
		where ((t.typrelid = 0 and (ce.relkind = 'c' or ce.relkind is null)) or c.relkind = 'c') 
		and not exists (select 1 from pg_depend d where d.objid in (t.oid, t.typelem) and d.deptype = 'e') 
		and n.nspname not in ('pg_catalog', 'information_schema')`
}

func (r *Postgres) CompileViews(database string) string {
	return "select viewname as name, schemaname as schema, definition from pg_views where schemaname not in ('pg_catalog', 'information_schema') order by viewname"
}

func (r *Postgres) EscapeNames(names []string) []string {
	escapedNames := make([]string, 0, len(names))

	for _, name := range names {
		segments := strings.Split(name, ".")
		for i, segment := range segments {
			segments[i] = strings.Trim(segment, `'"`)
		}
		escapedName := `"` + strings.Join(segments, `"."`) + `"`
		escapedNames = append(escapedNames, escapedName)
	}

	return escapedNames
}

func (r *Postgres) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Postgres) ModifyDefault(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Postgres) ModifyNullable(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Postgres) ModifyIncrement(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if !blueprint.HasCommand("primary") && slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		return " primary key"
	}

	return ""
}

func (r *Postgres) TypeBigInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "bigserial"
	}

	return "bigint"
}

func (r *Postgres) TypeChar(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("char(%d)", length)
	}

	return "char"
}

func (r *Postgres) TypeDecimal(column schema.ColumnDefinition) string {
	return fmt.Sprintf("decimal(%d, %d)", column.GetTotal(), column.GetPlaces())
}

func (r *Postgres) TypeDouble(column schema.ColumnDefinition) string {
	return "double precision"
}

func (r *Postgres) TypeEnum(column schema.ColumnDefinition) string {
	return fmt.Sprintf(`varchar(255) check ("%s" in (%s))`, column.GetName(), strings.Join(r.wrap.Quotes(column.GetAllowed()), ", "))
}

func (r *Postgres) TypeFloat(column schema.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("float(%d)", precision)
	}

	return "float"
}

func (r *Postgres) TypeInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "serial"
	}

	return "integer"
}

func (r *Postgres) TypeJson(column schema.ColumnDefinition) string {
	return "json"
}

func (r *Postgres) TypeJsonb(column schema.ColumnDefinition) string {
	return "jsonb"
}

func (r *Postgres) TypeLongText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Postgres) TypeMediumInteger(column schema.ColumnDefinition) string {
	return r.TypeInteger(column)
}

func (r *Postgres) TypeMediumText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Postgres) TypeText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Postgres) TypeSmallInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "smallserial"
	}

	return "smallint"
}

func (r *Postgres) TypeString(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}

	return "varchar"
}

func (r *Postgres) TypeTinyInteger(column schema.ColumnDefinition) string {
	return r.TypeSmallInteger(column)
}

func (r *Postgres) TypeTinyText(column schema.ColumnDefinition) string {
	return "varchar(255)"
}

func (r *Postgres) getColumns(blueprint schema.Blueprint) []string {
	var columns []string
	for _, column := range blueprint.GetAddedColumns() {
		columns = append(columns, r.getColumn(blueprint, column))
	}

	return columns
}

func (r *Postgres) getColumn(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	sql := fmt.Sprintf("%s %s", r.wrap.Column(column.GetName()), getType(r, column))

	for _, modifier := range r.modifiers {
		sql += modifier(blueprint, column)
	}

	return sql
}
