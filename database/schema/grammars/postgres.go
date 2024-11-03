package grammars

import (
	"fmt"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/constants"
)

type Postgres struct {
	attributeCommands []string
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
}

func NewPostgres() *Postgres {
	postgres := &Postgres{
		attributeCommands: []string{constants.CommandComment},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
	}
	postgres.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		postgres.ModifyDefault,
		postgres.ModifyIncrement,
		postgres.ModifyNullable,
	}

	return postgres
}

func (r *Postgres) CompileAdd(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s add column %s", blueprint.GetTableName(), getColumn(r, blueprint, command.Column))
}

func (r *Postgres) CompileCreate(blueprint schema.Blueprint, query orm.Query) string {
	return fmt.Sprintf("create table %s (%s)", blueprint.GetTableName(), strings.Join(getColumns(r, blueprint), ","))
}

func (r *Postgres) CompileDropAllDomains(domains []string) string {
	return fmt.Sprintf("drop domain %s cascade", strings.Join(domains, ", "))
}

func (r *Postgres) CompileDropAllTables(tables []string) string {
	return fmt.Sprintf("drop table %s cascade", strings.Join(tables, ", "))
}

func (r *Postgres) CompileDropAllTypes(types []string) string {
	return fmt.Sprintf("drop type %s cascade", strings.Join(types, ", "))
}

func (r *Postgres) CompileDropAllViews(views []string) string {
	return fmt.Sprintf("drop view %s cascade", strings.Join(views, ", "))
}

func (r *Postgres) CompileDropIfExists(blueprint schema.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", blueprint.GetTableName())
}

func (r *Postgres) CompileTables() string {
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

func (r *Postgres) CompileViews() string {
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

func (r *Postgres) GetModifiers() []func(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	return r.modifiers
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

func (r *Postgres) TypeInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "serial"
	}

	return "integer"
}

func (r *Postgres) TypeString(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}

	return "varchar"
}
