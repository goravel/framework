package grammars

import (
	"fmt"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/orm"
)

type Postgres struct {
	attributeCommands []string
	modifiers         []func(migration.Blueprint, migration.ColumnDefinition) string
	serials           []string
}

func NewPostgres() *Postgres {
	postgres := &Postgres{
		attributeCommands: []string{"comment"},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
	}
	postgres.modifiers = []func(migration.Blueprint, migration.ColumnDefinition) string{
		postgres.ModifyDefault,
		postgres.ModifyIncrement,
		postgres.ModifyNullable,
	}

	return postgres
}

func (r *Postgres) CompileCreate(blueprint migration.Blueprint, query orm.Query) string {
	return fmt.Sprintf("create table %s (%s)", blueprint.GetTableName(), strings.Join(getColumns(r, blueprint), ","))
}

func (r *Postgres) CompileDropIfExists(blueprint migration.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", blueprint.GetTableName())
}

func (r *Postgres) CompileTables(database string) string {
	return "select c.relname as name, n.nspname as schema, pg_total_relation_size(c.oid) as size, " +
		"obj_description(c.oid, 'pg_class') as comment from pg_class c, pg_namespace n " +
		"where c.relkind in ('r', 'p') and n.oid = c.relnamespace and n.nspname not in ('pg_catalog', 'information_schema') " +
		"order by c.relname"
}

func (r *Postgres) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Postgres) GetModifiers() []func(blueprint migration.Blueprint, column migration.ColumnDefinition) string {
	return r.modifiers
}

func (r *Postgres) ModifyDefault(blueprint migration.Blueprint, column migration.ColumnDefinition) string {
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

func (r *Postgres) ModifyNullable(blueprint migration.Blueprint, column migration.ColumnDefinition) string {
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

func (r *Postgres) ModifyIncrement(blueprint migration.Blueprint, column migration.ColumnDefinition) string {
	if !column.GetChange() && !blueprint.HasCommand("primary") && slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		return " primary key"
	}

	return ""
}

func (r *Postgres) TypeBigInteger(column migration.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "bigserial"
	}

	return "bigint"
}

func (r *Postgres) TypeInteger(column migration.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "serial"
	}

	return "integer"
}

func (r *Postgres) TypeString(column migration.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}

	return "varchar"
}
