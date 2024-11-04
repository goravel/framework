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

func (r *Sqlite) CompileAdd(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s add column %s", blueprint.GetTableName(), getColumn(r, blueprint, command.Column))
}

func (r *Sqlite) CompileCreate(blueprint schema.Blueprint, query orm.Query) string {
	return fmt.Sprintf("create table %s (%s%s%s)",
		blueprint.GetTableName(),
		strings.Join(getColumns(r, blueprint), ","),
		r.addForeignKeys(getCommandsByName(blueprint.GetCommands(), "foreign")),
		r.addPrimaryKeys(getCommandByName(blueprint.GetCommands(), "primary")))
}

func (r *Sqlite) CompileDropAllDomains(domains []string) string {
	return ""
}

func (r *Sqlite) CompileDropAllTables(tables []string) string {
	return "delete from sqlite_master where type in ('table', 'index', 'trigger')"
}

func (r *Sqlite) CompileDropAllTypes(types []string) string {
	return ""
}

func (r *Sqlite) CompileDropAllViews(views []string) string {
	return "delete from sqlite_master where type in ('view')"
}

func (r *Sqlite) CompileDropIfExists(blueprint schema.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", blueprint.GetTableName())
}

func (r *Sqlite) CompileTables() string {
	return "select name from sqlite_master where type = 'table' and name not like 'sqlite_%' order by name"
}

func (r *Sqlite) CompileTypes() string {
	return ""
}

func (r *Sqlite) CompileViews() string {
	return "select name, sql as definition from sqlite_master where type = 'view' order by name"
}

func (r *Sqlite) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Sqlite) GetModifiers() []func(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	return r.modifiers
}

func (r *Sqlite) ModifyDefault(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Sqlite) ModifyNullable(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Sqlite) ModifyIncrement(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		return " primary key autoincrement"
	}

	return ""
}

func (r *Sqlite) TypeBigInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "serial"
	}

	return "integer"
}

func (r *Sqlite) TypeString(column schema.ColumnDefinition) string {
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
