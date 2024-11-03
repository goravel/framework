package grammars

import (
	"fmt"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
)

type Sqlite struct {
	attributeCommands []string
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
}

func NewSqlite() *Sqlite {
	sqlite := &Sqlite{
		attributeCommands: []string{},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
	}
	sqlite.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		sqlite.ModifyDefault,
		sqlite.ModifyIncrement,
		sqlite.ModifyNullable,
	}

	return sqlite
}

func (r *Sqlite) CompileAdd(blueprint schema.Blueprint) string {
	//return fmt.Sprintf("alter table %s add column %s", blueprint.GetTableName(), getColumn(r, blueprint, command.Column))
	return ""
}

func (r *Sqlite) CompileCreate(blueprint schema.Blueprint, query orm.Query) string {
	return fmt.Sprintf("create table %s (%s%s%s)",
		blueprint.GetTableName(),
		strings.Join(getColumns(r, blueprint), ","),
		r.addForeignKeys(getCommandsByName(blueprint.GetCommands(), "foreign")),
		r.addPrimaryKeys(getCommandByName(blueprint.GetCommands(), "primary")))
}

func (r *Sqlite) CompileDropAllDomains(domains []string) string {
	return fmt.Sprintf("drop domain %s cascade", strings.Join(domains, ", "))
}

func (r *Sqlite) CompileDropAllTables(tables []string) string {
	return fmt.Sprintf("drop table %s cascade", strings.Join(tables, ", "))
}

func (r *Sqlite) CompileDropAllTypes(types []string) string {
	return fmt.Sprintf("drop type %s cascade", strings.Join(types, ", "))
}

func (r *Sqlite) CompileDropAllViews(views []string) string {
	return fmt.Sprintf("drop view %s cascade", strings.Join(views, ", "))
}

func (r *Sqlite) CompileDropIfExists(blueprint schema.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", blueprint.GetTableName())
}

func (r *Sqlite) CompileTables(database string) string {
	return "select c.relname as name, n.nspname as schema, pg_total_relation_size(c.oid) as size, " +
		"obj_description(c.oid, 'pg_class') as comment from pg_class c, pg_namespace n " +
		"where c.relkind in ('r', 'p') and n.oid = c.relnamespace and n.nspname not in ('pg_catalog', 'information_schema') " +
		"order by c.relname"
}

func (r *Sqlite) CompileTypes() string {
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

func (r *Sqlite) CompileViews() string {
	return "select viewname as name, schemaname as schema, definition from pg_views where schemaname not in ('pg_catalog', 'information_schema') order by viewname"
}

func (r *Sqlite) EscapeNames(names []string) []string {
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

func (r *Sqlite) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Sqlite) GetModifiers() []func(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	return r.modifiers
}

func (r *Sqlite) ModifyDefault(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
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

func (r *Sqlite) ModifyNullable(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
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

func (r *Sqlite) ModifyIncrement(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if !column.GetChange() && !blueprint.HasCommand("primary") && slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		return " primary key"
	}

	return ""
}

func (r *Sqlite) TypeBigInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "bigserial"
	}

	return "bigint"
}

func (r *Sqlite) TypeInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "serial"
	}

	return "integer"
}

func (r *Sqlite) TypeString(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}

	return "varchar"
}

// addForeignKeys Get the foreign key syntax for a table creation statement.
func (r *Sqlite) addForeignKeys([]*schema.Command) string {
	return ""
}

func (r *Sqlite) addPrimaryKeys(command *schema.Command) string {
	if command == nil {
		return ""
	}

	return fmt.Sprintf(", primary key (%s)", strings.Join(r.EscapeNames(command.Columns), ", "))
}

// getForeignKey Get the SQL for the foreign key.
func (r *Sqlite) getForeignKey(*schema.Command) string {
	return ""
}
