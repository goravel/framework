package grammars

import (
	"fmt"
	"slices"
	"strings"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
)

type Mysql struct {
	attributeCommands []string
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
	wrap              *Wrap
}

func NewMysql(tablePrefix string) *Mysql {
	mysql := &Mysql{
		attributeCommands: []string{},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		wrap:              NewWrap(contractsdatabase.DriverMysql, tablePrefix),
	}
	mysql.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		mysql.ModifyComment,
		mysql.ModifyDefault,
		mysql.ModifyIncrement,
		mysql.ModifyNullable,
		mysql.ModifyOnUpdate,
	}

	return mysql
}

func (r *Mysql) CompileAdd(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s add %s", r.wrap.Table(blueprint.GetTableName()), r.getColumn(blueprint, command.Column))
}

func (r *Mysql) CompileColumns(schema, table string) string {
	return fmt.Sprintf(
		"select column_name as `name`, data_type as `type_name`, column_type as `type`, "+
			"collation_name as `collation`, is_nullable as `nullable`, "+
			"column_default as `default`, column_comment as `comment`, "+
			"generation_expression as `expression`, extra as `extra` "+
			"from information_schema.columns where table_schema = %s and table_name = %s "+
			"order by ordinal_position asc", r.wrap.Quote(schema), r.wrap.Quote(table))
}

func (r *Mysql) CompileComment(_ schema.Blueprint, _ *schema.Command) string {
	return ""
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

func (r *Mysql) CompileDrop(blueprint schema.Blueprint) string {
	return fmt.Sprintf("drop table %s", r.wrap.Table(blueprint.GetTableName()))
}

func (r *Mysql) CompileDropAllDomains(_ []string) string {
	return ""
}

func (r *Mysql) CompileDropAllTables(tables []string) string {
	return fmt.Sprintf("drop table %s", r.wrap.Columnize(tables))
}

func (r *Mysql) CompileDropAllTypes(_ []string) string {
	return ""
}

func (r *Mysql) CompileDropAllViews(views []string) string {
	return fmt.Sprintf("drop view %s", r.wrap.Columnize(views))
}

func (r *Mysql) CompileDropColumn(blueprint schema.Blueprint, command *schema.Command) []string {
	columns := r.wrap.PrefixArray("drop", r.wrap.Columns(command.Columns))

	return []string{
		fmt.Sprintf("alter table %s %s", r.wrap.Table(blueprint.GetTableName()), strings.Join(columns, ", ")),
	}
}

func (r *Mysql) CompileDropForeign(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s drop foreign key %s", r.wrap.Table(blueprint.GetTableName()), r.wrap.Column(command.Index))
}

func (r *Mysql) CompileDropFullText(blueprint schema.Blueprint, command *schema.Command) string {
	return r.CompileDropIndex(blueprint, command)
}

func (r *Mysql) CompileDropIfExists(blueprint schema.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", r.wrap.Table(blueprint.GetTableName()))
}

func (r *Mysql) CompileDropIndex(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s drop index %s", r.wrap.Table(blueprint.GetTableName()), r.wrap.Column(command.Index))
}

func (r *Mysql) CompileDropPrimary(blueprint schema.Blueprint, _ *schema.Command) string {
	return fmt.Sprintf("alter table %s drop primary key", r.wrap.Table(blueprint.GetTableName()))
}

func (r *Mysql) CompileDropUnique(blueprint schema.Blueprint, command *schema.Command) string {
	return r.CompileDropIndex(blueprint, command)
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

func (r *Mysql) CompileFullText(blueprint schema.Blueprint, command *schema.Command) string {
	return r.compileKey(blueprint, command, "fulltext")
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

func (r *Mysql) CompileRenameIndex(_ schema.Schema, blueprint schema.Blueprint, command *schema.Command) []string {
	return []string{
		fmt.Sprintf("alter table %s rename index %s to %s", r.wrap.Table(blueprint.GetTableName()), r.wrap.Column(command.From), r.wrap.Column(command.To)),
	}
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

func (r *Mysql) CompileUnique(blueprint schema.Blueprint, command *schema.Command) string {
	return r.compileKey(blueprint, command, "unique")
}

func (r *Mysql) CompileViews(database string) string {
	return fmt.Sprintf("select table_name as `name`, view_definition as `definition` "+
		"from information_schema.views where table_schema = %s "+
		"order by table_name", r.wrap.Quote(database))
}

func (r *Mysql) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Mysql) ModifyComment(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if comment := column.GetComment(); comment != "" {
		// Escape special characters to prevent SQL injection
		comment = strings.ReplaceAll(comment, "'", "''")
		comment = strings.ReplaceAll(comment, "\\", "\\\\")

		return fmt.Sprintf(" comment '%s'", comment)
	}

	return ""
}

func (r *Mysql) ModifyDefault(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Mysql) ModifyNullable(_ schema.Blueprint, column schema.ColumnDefinition) string {
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

func (r *Mysql) ModifyOnUpdate(_ schema.Blueprint, column schema.ColumnDefinition) string {
	onUpdate := column.GetOnUpdate()
	if onUpdate != nil {
		switch value := onUpdate.(type) {
		case Expression:
			return " on update " + string(value)
		case string:
			if onUpdate.(string) != "" {
				return " on update " + value
			}
		}
	}

	return ""
}

func (r *Mysql) TypeBigInteger(_ schema.ColumnDefinition) string {
	return "bigint"
}

func (r *Mysql) TypeChar(column schema.ColumnDefinition) string {
	return fmt.Sprintf("char(%d)", column.GetLength())
}

func (r *Mysql) TypeDate(_ schema.ColumnDefinition) string {
	return "date"
}

func (r *Mysql) TypeDateTime(column schema.ColumnDefinition) string {
	current := "CURRENT_TIMESTAMP"
	precision := column.GetPrecision()
	if precision > 0 {
		current = fmt.Sprintf("CURRENT_TIMESTAMP(%d)", precision)
	}
	if column.GetUseCurrent() {
		column.Default(Expression(current))
	}
	if column.GetUseCurrentOnUpdate() {
		column.OnUpdate(Expression(current))
	}

	if precision > 0 {
		return fmt.Sprintf("datetime(%d)", precision)
	} else {
		return "datetime"
	}
}

func (r *Mysql) TypeDateTimeTz(column schema.ColumnDefinition) string {
	return r.TypeDateTime(column)
}

func (r *Mysql) TypeDecimal(column schema.ColumnDefinition) string {
	return fmt.Sprintf("decimal(%d, %d)", column.GetTotal(), column.GetPlaces())
}

func (r *Mysql) TypeDouble(_ schema.ColumnDefinition) string {
	return "double"
}

func (r *Mysql) TypeEnum(column schema.ColumnDefinition) string {
	return fmt.Sprintf(`enum(%s)`, strings.Join(r.wrap.Quotes(column.GetAllowed()), ", "))
}

func (r *Mysql) TypeFloat(column schema.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("float(%d)", precision)
	}

	return "float"
}

func (r *Mysql) TypeInteger(_ schema.ColumnDefinition) string {
	return "int"
}

func (r *Mysql) TypeJson(_ schema.ColumnDefinition) string {
	return "json"
}

func (r *Mysql) TypeJsonb(_ schema.ColumnDefinition) string {
	return "json"
}

func (r *Mysql) TypeLongText(_ schema.ColumnDefinition) string {
	return "longtext"
}

func (r *Mysql) TypeMediumInteger(_ schema.ColumnDefinition) string {
	return "mediumint"
}

func (r *Mysql) TypeMediumText(_ schema.ColumnDefinition) string {
	return "mediumtext"
}

func (r *Mysql) TypeSmallInteger(_ schema.ColumnDefinition) string {
	return "smallint"
}

func (r *Mysql) TypeString(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}

	return "varchar(255)"
}

func (r *Mysql) TypeText(_ schema.ColumnDefinition) string {
	return "text"
}

func (r *Mysql) TypeTime(column schema.ColumnDefinition) string {
	if column.GetPrecision() > 0 {
		return fmt.Sprintf("time(%d)", column.GetPrecision())
	} else {
		return "time"
	}
}

func (r *Mysql) TypeTimeTz(column schema.ColumnDefinition) string {
	return r.TypeTime(column)
}

func (r *Mysql) TypeTimestamp(column schema.ColumnDefinition) string {
	current := "CURRENT_TIMESTAMP"
	precision := column.GetPrecision()
	if precision > 0 {
		current = fmt.Sprintf("CURRENT_TIMESTAMP(%d)", precision)
	}
	if column.GetUseCurrent() {
		column.Default(Expression(current))
	}
	if column.GetUseCurrentOnUpdate() {
		column.OnUpdate(Expression(current))
	}

	if precision > 0 {
		return fmt.Sprintf("timestamp(%d)", precision)
	} else {
		return "timestamp"
	}
}

func (r *Mysql) TypeTimestampTz(column schema.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}

func (r *Mysql) TypeTinyInteger(_ schema.ColumnDefinition) string {
	return "tinyint"
}

func (r *Mysql) TypeTinyText(_ schema.ColumnDefinition) string {
	return "tinytext"
}

func (r *Mysql) compileKey(blueprint schema.Blueprint, command *schema.Command, ttype string) string {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = " using " + command.Algorithm
	}

	return fmt.Sprintf("alter table %s add %s %s%s(%s)",
		r.wrap.Table(blueprint.GetTableName()),
		ttype,
		r.wrap.Column(command.Index),
		algorithm,
		r.wrap.Columnize(command.Columns))
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
