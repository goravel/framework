package grammars

import (
	"fmt"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/constants"
)

type Sqlserver struct {
	attributeCommands []string
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
	wrap              *Wrap
}

func NewSqlserver(tablePrefix string) *Sqlserver {
	sqlserver := &Sqlserver{
		attributeCommands: []string{constants.CommandComment},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		wrap:              NewWrap(database.DriverSqlserver, tablePrefix),
	}
	sqlserver.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		sqlserver.ModifyDefault,
		sqlserver.ModifyIncrement,
		sqlserver.ModifyNullable,
	}

	return sqlserver
}

func (r *Sqlserver) CompileAdd(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s add %s", r.wrap.Table(blueprint.GetTableName()), r.getColumn(blueprint, command.Column))
}

func (r *Sqlserver) CompileColumns(schema, table string) string {
	newSchema := "schema_name()"
	if schema != "" {
		newSchema = r.wrap.Quote(schema)
	}

	return fmt.Sprintf(
		"select col.name, type.name as type_name, "+
			"col.max_length as length, col.precision as precision, col.scale as places, "+
			"col.is_nullable as nullable, def.definition as [default], "+
			"col.is_identity as autoincrement, col.collation_name as collation, "+
			"com.definition as [expression], is_persisted as [persisted], "+
			"cast(prop.value as nvarchar(max)) as comment "+
			"from sys.columns as col "+
			"join sys.types as type on col.user_type_id = type.user_type_id "+
			"join sys.objects as obj on col.object_id = obj.object_id "+
			"join sys.schemas as scm on obj.schema_id = scm.schema_id "+
			"left join sys.default_constraints def on col.default_object_id = def.object_id and col.object_id = def.parent_object_id "+
			"left join sys.extended_properties as prop on obj.object_id = prop.major_id and col.column_id = prop.minor_id and prop.name = 'MS_Description' "+
			"left join sys.computed_columns as com on col.column_id = com.column_id and col.object_id = com.object_id "+
			"where obj.type in ('U', 'V') and obj.name = %s and scm.name = %s "+
			"order by col.column_id", r.wrap.Quote(table), newSchema)
}

func (r *Sqlserver) CompileComment(blueprint schema.Blueprint, command *schema.Command) string {
	return ""
}

func (r *Sqlserver) CompileCreate(blueprint schema.Blueprint) string {
	return fmt.Sprintf("create table %s (%s)", r.wrap.Table(blueprint.GetTableName()), strings.Join(r.getColumns(blueprint), ", "))
}

func (r *Sqlserver) CompileDropAllDomains(domains []string) string {
	return ""
}

func (r *Sqlserver) CompileDropAllForeignKeys() string {
	return `DECLARE @sql NVARCHAR(MAX) = N'';
            SELECT @sql += 'ALTER TABLE '
                + QUOTENAME(OBJECT_SCHEMA_NAME(parent_object_id)) + '.' + + QUOTENAME(OBJECT_NAME(parent_object_id))
                + ' DROP CONSTRAINT ' + QUOTENAME(name) + ';'
            FROM sys.foreign_keys;

            EXEC sp_executesql @sql;`
}

func (r *Sqlserver) CompileDropAllTables(tables []string) string {
	return "EXEC sp_msforeachtable 'DROP TABLE ?'"
}

func (r *Sqlserver) CompileDropAllTypes(types []string) string {
	return ""
}

func (r *Sqlserver) CompileDropAllViews(views []string) string {
	return `DECLARE @sql NVARCHAR(MAX) = N'';
	SELECT @sql += 'DROP VIEW ' + QUOTENAME(OBJECT_SCHEMA_NAME(object_id)) + '.' + QUOTENAME(name) + ';'
	FROM sys.views;

	EXEC sp_executesql @sql;`
}

func (r *Sqlserver) CompileDropIfExists(blueprint schema.Blueprint) string {
	table := r.wrap.Table(blueprint.GetTableName())

	return fmt.Sprintf("if object_id(%s, 'U') is not null drop table %s", r.wrap.Quote(table), table)
}

func (r *Sqlserver) CompileForeign(blueprint schema.Blueprint, command *schema.Command) string {
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

func (r *Sqlserver) CompileIndex(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("create index %s on %s (%s)",
		r.wrap.Column(command.Index),
		r.wrap.Table(blueprint.GetTableName()),
		r.wrap.Columnize(command.Columns),
	)
}

func (r *Sqlserver) CompileIndexes(schema, table string) string {
	newSchema := "schema_name()"
	if schema != "" {
		newSchema = r.wrap.Quote(schema)
	}

	return fmt.Sprintf(
		"select idx.name as name, string_agg(col.name, ',') within group (order by idxcol.key_ordinal) as columns, "+
			"idx.type_desc as [type], idx.is_unique as [unique], idx.is_primary_key as [primary] "+
			"from sys.indexes as idx "+
			"join sys.tables as tbl on idx.object_id = tbl.object_id "+
			"join sys.schemas as scm on tbl.schema_id = scm.schema_id "+
			"join sys.index_columns as idxcol on idx.object_id = idxcol.object_id and idx.index_id = idxcol.index_id "+
			"join sys.columns as col on idxcol.object_id = col.object_id and idxcol.column_id = col.column_id "+
			"where tbl.name = %s and scm.name = %s "+
			"group by idx.name, idx.type_desc, idx.is_unique, idx.is_primary_key",
		r.wrap.Quote(table),
		newSchema,
	)
}

func (r *Sqlserver) CompilePrimary(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s add constraint %s primary key (%s)",
		r.wrap.Table(blueprint.GetTableName()),
		r.wrap.Column(command.Index),
		r.wrap.Columnize(command.Columns))
}

func (r *Sqlserver) CompileTables(database string) string {
	return "select t.name as name, schema_name(t.schema_id) as [schema], sum(u.total_pages) * 8 * 1024 as size " +
		"from sys.tables as t " +
		"join sys.partitions as p on p.object_id = t.object_id " +
		"join sys.allocation_units as u on u.container_id = p.hobt_id " +
		"group by t.name, t.schema_id " +
		"order by t.name"
}

func (r *Sqlserver) CompileTypes() string {
	return ""
}

func (r *Sqlserver) CompileViews(database string) string {
	return "select name, schema_name(v.schema_id) as [schema], definition from sys.views as v " +
		"inner join sys.sql_modules as m on v.object_id = m.object_id " +
		"order by name"
}

func (r *Sqlserver) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Sqlserver) ModifyDefault(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Sqlserver) ModifyNullable(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Sqlserver) ModifyIncrement(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		if blueprint.HasCommand("primary") {
			return " identity"
		}
		return " identity primary key"
	}

	return ""
}

func (r *Sqlserver) TypeBigInteger(column schema.ColumnDefinition) string {
	return "bigint"
}

func (r *Sqlserver) TypeChar(column schema.ColumnDefinition) string {
	return fmt.Sprintf("nchar(%d)", column.GetLength())
}

func (r *Sqlserver) TypeDecimal(column schema.ColumnDefinition) string {
	return fmt.Sprintf("decimal(%d, %d)", column.GetTotal(), column.GetPlaces())
}

func (r *Sqlserver) TypeDouble(column schema.ColumnDefinition) string {
	return "double precision"
}

func (r *Sqlserver) TypeEnum(column schema.ColumnDefinition) string {
	return fmt.Sprintf(`nvarchar(255) check ("%s" in (%s))`, column.GetName(), strings.Join(r.wrap.Quotes(column.GetAllowed()), ", "))
}

func (r *Sqlserver) TypeFloat(column schema.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("float(%d)", precision)
	}

	return "float"
}

func (r *Sqlserver) TypeInteger(column schema.ColumnDefinition) string {
	return "int"
}

func (r *Sqlserver) TypeJson(column schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeJsonb(column schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeLongText(column schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeMediumInteger(column schema.ColumnDefinition) string {
	return "int"
}

func (r *Sqlserver) TypeMediumText(column schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeText(column schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeSmallInteger(column schema.ColumnDefinition) string {
	return "smallint"
}

func (r *Sqlserver) TypeString(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("nvarchar(%d)", length)
	}

	return "nvarchar(255)"
}

func (r *Sqlserver) TypeTinyInteger(column schema.ColumnDefinition) string {
	return "tinyint"
}

func (r *Sqlserver) TypeTinyText(column schema.ColumnDefinition) string {
	return "nvarchar(255)"
}

func (r *Sqlserver) getColumns(blueprint schema.Blueprint) []string {
	var columns []string
	for _, column := range blueprint.GetAddedColumns() {
		columns = append(columns, r.getColumn(blueprint, column))
	}

	return columns
}

func (r *Sqlserver) getColumn(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	sql := fmt.Sprintf("%s %s", r.wrap.Column(column.GetName()), getType(r, column))

	for _, modifier := range r.modifiers {
		sql += modifier(blueprint, column)
	}

	return sql
}
