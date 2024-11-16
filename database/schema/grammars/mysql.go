package grammars

import (
	"fmt"
	"slices"
	"strings"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/constants"
)

type Mysql struct {
	attributeCommands []string
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
	wrap              *Wrap
}

func NewMysql(tablePrefix string) *Mysql {
	postgres := &Mysql{
		attributeCommands: []string{constants.CommandComment},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		wrap:              NewWrap(contractsdatabase.DriverMysql, tablePrefix),
	}
	postgres.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		postgres.ModifyDefault,
		postgres.ModifyIncrement,
		postgres.ModifyNullable,
	}

	return postgres
}

func (r *Mysql) CompileAdd(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s add %s", r.wrap.Table(blueprint.GetTableName()), r.getColumn(blueprint, command.Column))
}

func (r *Mysql) CompileCreate(blueprint schema.Blueprint) string {
	columns := r.getColumns(blueprint)
	primaryCommand := getCommandByName(blueprint.GetCommands(), "primary")
	if primaryCommand != nil {
		var algorithm string
		if primaryCommand.Algorithm != "" {
			algorithm = "using " + primaryCommand.Algorithm
		}
		columns = append(columns, fmt.Sprintf("primary key %s(%s)", algorithm, r.wrap.Columnize(primaryCommand.Columns)))

		primaryCommand.ShouldBeSkipped = true
	}

	return fmt.Sprintf("create table %s (%s)", r.wrap.Table(blueprint.GetTableName()), strings.Join(columns, ", "))
}

func (r *Mysql) CompileDisableForeignKeyConstraints() string {
	return "SET FOREIGN_KEY_CHECKS=0;"
}

func (r *Mysql) CompileDropAllDomains(domains []string) string {
	return ""
}

func (r *Mysql) CompileDropAllTables(tables []string) string {
	return fmt.Sprintf("drop table %s", strings.Join(r.wrap.Columns(tables), ", "))
}

func (r *Mysql) CompileDropAllTypes(types []string) string {
	return ""
}

func (r *Mysql) CompileDropAllViews(views []string) string {
	return fmt.Sprintf("drop view %s", strings.Join(r.wrap.Columns(views), ", "))
}

func (r *Mysql) CompileDropIfExists(blueprint schema.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", r.wrap.Table(blueprint.GetTableName()))
}

func (r *Mysql) CompileEnableForeignKeyConstraints() string {
	return "SET FOREIGN_KEY_CHECKS=1;"
}

func (r *Mysql) CompileForeign(blueprint schema.Blueprint, command *schema.Command) string {
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

func (r *Mysql) CompileIndex(blueprint schema.Blueprint, command *schema.Command) string {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = " using " + command.Algorithm
	}

	return fmt.Sprintf("alter table %s add %s %s%s(%s)",
		r.wrap.Table(blueprint.GetTableName()),
		"index",
		r.wrap.Column(command.Index),
		algorithm,
		r.wrap.Columnize(command.Columns),
	)
}

func (r *Mysql) CompileIndexes(schema, table string) string {
	return fmt.Sprintf(
		"select index_name as `name`, group_concat(column_name order by seq_in_index) as `columns`, "+
			"index_type as `type`, not non_unique as `unique` "+
			"from information_schema.statistics where table_schema = %s and table_name = %s "+
			"group by index_name, index_type, non_unique",
		r.wrap.Quote(schema),
		r.wrap.Quote(table),
	)
}

func (r *Mysql) CompilePrimary(blueprint schema.Blueprint, command *schema.Command) string {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = "using " + command.Algorithm
	}

	return fmt.Sprintf("alter table %s add primary key %s(%s)", r.wrap.Table(blueprint.GetTableName()), algorithm, r.wrap.Columnize(command.Columns))
}

func (r *Mysql) CompileTables(database string) string {
	return fmt.Sprintf("select table_name as `name`, (data_length + index_length) as `size`, "+
		"table_comment as `comment`, engine as `engine`, table_collation as `collation` "+
		"from information_schema.tables where table_schema = %s and table_type in ('BASE TABLE', 'SYSTEM VERSIONED') "+
		"order by table_name", r.wrap.Quote(database))
}

func (r *Mysql) CompileTypes() string {
	return ""
}

func (r *Mysql) CompileViews(database string) string {
	return fmt.Sprintf("select table_name as `name`, view_definition as `definition` "+
		"from information_schema.views where table_schema = %s "+
		"order by table_name", r.wrap.Quote(database))
}

func (r *Mysql) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Mysql) ModifyDefault(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Mysql) ModifyNullable(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Mysql) ModifyIncrement(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		if blueprint.HasCommand("primary") {
			return "auto_increment"
		}
		return " auto_increment primary key"
	}

	return ""
}

func (r *Mysql) TypeBigInteger(column schema.ColumnDefinition) string {
	return "bigint"
}

func (r *Mysql) TypeInteger(column schema.ColumnDefinition) string {
	return "int"
}

func (r *Mysql) TypeString(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}

	return "varchar"
}

func (r *Mysql) getColumns(blueprint schema.Blueprint) []string {
	var columns []string
	for _, column := range blueprint.GetAddedColumns() {
		columns = append(columns, r.getColumn(blueprint, column))
	}

	return columns
}

func (r *Mysql) getColumn(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	sql := fmt.Sprintf("%s %s", r.wrap.Column(column.GetName()), getType(r, column))

	for _, modifier := range r.modifiers {
		sql += modifier(blueprint, column)
	}

	return sql
}
