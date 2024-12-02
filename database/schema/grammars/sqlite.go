package grammars

import (
	"fmt"
	"slices"
	"strings"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/collect"
)

type Sqlite struct {
	attributeCommands []string
	log               log.Log
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
	tablePrefix       string
	wrap              *Wrap
}

func NewSqlite(log log.Log, tablePrefix string) *Sqlite {
	sqlite := &Sqlite{
		attributeCommands: []string{},
		log:               log,
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		tablePrefix:       tablePrefix,
		wrap:              NewWrap(contractsdatabase.DriverSqlite, tablePrefix),
	}
	sqlite.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		sqlite.ModifyDefault,
		sqlite.ModifyIncrement,
		sqlite.ModifyNullable,
	}

	return sqlite
}

func (r *Sqlite) CompileAdd(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("alter table %s add column %s", r.wrap.Table(blueprint.GetTableName()), r.getColumn(blueprint, command.Column))
}

func (r *Sqlite) CompileColumns(schema, table string) string {
	return fmt.Sprintf(
		`select name, type, not "notnull" as "nullable", dflt_value as "default", pk as "primary", hidden as "extra" `+
			"from pragma_table_xinfo(%s) order by cid asc", r.wrap.Quote(strings.ReplaceAll(table, ".", "__")))
}

func (r *Sqlite) CompileComment(blueprint schema.Blueprint, command *schema.Command) string {
	return ""
}

func (r *Sqlite) CompileCreate(blueprint schema.Blueprint) string {
	return fmt.Sprintf("create table %s (%s%s%s)",
		r.wrap.Table(blueprint.GetTableName()),
		strings.Join(r.getColumns(blueprint), ", "),
		r.addForeignKeys(getCommandsByName(blueprint.GetCommands(), "foreign")),
		r.addPrimaryKeys(getCommandByName(blueprint.GetCommands(), "primary")))
}

func (r *Sqlite) CompileDisableWriteableSchema() string {
	return r.pragma("writable_schema", "0")
}

func (r *Sqlite) CompileDrop(blueprint schema.Blueprint) string {
	return fmt.Sprintf("drop table %s", r.wrap.Table(blueprint.GetTableName()))
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

func (r *Sqlite) CompileDropColumn(blueprint schema.Blueprint, command *schema.Command) []string {
	// TODO check Sqlite 3.35
	table := r.wrap.Table(blueprint.GetTableName())
	columns := r.wrap.PrefixArray("drop column", r.wrap.Columns(command.Columns))

	return collect.Map(columns, func(column string, _ int) string {
		return fmt.Sprintf("alter table %s %s", table, column)
	})
}

func (r *Sqlite) CompileDropForeign(_ schema.Blueprint, _ *schema.Command) string {
	return ""
}

func (r *Sqlite) CompileDropFullText(_ schema.Blueprint, _ *schema.Command) string {
	return ""
}

func (r *Sqlite) CompileDropIfExists(blueprint schema.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", r.wrap.Table(blueprint.GetTableName()))
}

func (r *Sqlite) CompileDropIndex(_ schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("drop index %s", r.wrap.Column(command.Index))
}

func (r *Sqlite) CompileDropPrimary(_ schema.Blueprint, _ *schema.Command) string {
	return ""
}

func (r *Sqlite) CompileDropUnique(blueprint schema.Blueprint, command *schema.Command) string {
	return r.CompileDropIndex(blueprint, command)
}

func (r *Sqlite) CompileEnableWriteableSchema() string {
	return r.pragma("writable_schema", "1")
}

func (r *Sqlite) CompileForeign(_ schema.Blueprint, _ *schema.Command) string {
	return ""
}

func (r *Sqlite) CompileFullText(_ schema.Blueprint, _ *schema.Command) string {
	return ""
}

func (r *Sqlite) CompileIndex(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("create index %s on %s (%s)",
		r.wrap.Column(command.Index),
		r.wrap.Table(blueprint.GetTableName()),
		r.wrap.Columnize(command.Columns),
	)
}

func (r *Sqlite) CompileIndexes(_, table string) string {
	quotedTable := r.wrap.Quote(strings.ReplaceAll(table, ".", "__"))

	return fmt.Sprintf(
		`select 'primary' as name, group_concat(col) as columns, 1 as "unique", 1 as "primary" `+
			`from (select name as col from pragma_table_info(%s) where pk > 0 order by pk, cid) group by name `+
			`union select name, group_concat(col) as columns, "unique", origin = 'pk' as "primary" `+
			`from (select il.*, ii.name as col from pragma_index_list(%s) il, pragma_index_info(il.name) ii order by il.seq, ii.seqno) `+
			`group by name, "unique", "primary"`,
		quotedTable,
		r.wrap.Quote(table),
	)
}

func (r *Sqlite) CompilePrimary(_ schema.Blueprint, _ *schema.Command) string {
	return ""
}

func (r *Sqlite) CompileRebuild() string {
	return "vacuum"
}

func (r *Sqlite) CompileRenameIndex(s schema.Schema, blueprint schema.Blueprint, command *schema.Command) []string {
	indexes, err := s.GetIndexes(blueprint.GetTableName())
	if err != nil {
		r.log.Errorf("failed to get %s indexes: %v", blueprint.GetTableName(), err)
		return nil
	}

	collect.Filter(indexes, func(index schema.Index, _ int) bool {
		return index.Name == command.From
	})

	if len(indexes) == 0 {
		r.log.Warningf("index %s does not exist", command.From)
		return nil
	}
	if indexes[0].Primary {
		r.log.Warning("SQLite does not support altering primary keys")
		return nil
	}
	if indexes[0].Unique {
		return []string{
			r.CompileDropUnique(blueprint, &schema.Command{
				Index: indexes[0].Name,
			}),
			r.CompileUnique(blueprint, &schema.Command{
				Index:   command.To,
				Columns: indexes[0].Columns,
			}),
		}
	}

	return []string{
		r.CompileDropIndex(blueprint, &schema.Command{
			Index: indexes[0].Name,
		}),
		r.CompileIndex(blueprint, &schema.Command{
			Index:   command.To,
			Columns: indexes[0].Columns,
		}),
	}
}

func (r *Sqlite) CompileTables(database string) string {
	return "select name from sqlite_master where type = 'table' and name not like 'sqlite_%' order by name"
}

func (r *Sqlite) CompileTypes() string {
	return ""
}

func (r *Sqlite) CompileUnique(blueprint schema.Blueprint, command *schema.Command) string {
	return fmt.Sprintf("create unique index %s on %s (%s)",
		r.wrap.Column(command.Index),
		r.wrap.Table(blueprint.GetTableName()),
		r.wrap.Columnize(command.Columns))
}

func (r *Sqlite) CompileViews(database string) string {
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

func (r *Sqlite) TypeChar(column schema.ColumnDefinition) string {
	return "varchar"
}

func (r *Sqlite) TypeDate(column schema.ColumnDefinition) string {
	return "date"
}

func (r *Sqlite) TypeDateTime(column schema.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}

func (r *Sqlite) TypeDateTimeTz(column schema.ColumnDefinition) string {
	return r.TypeDateTime(column)
}

func (r *Sqlite) TypeDecimal(column schema.ColumnDefinition) string {
	return "numeric"
}

func (r *Sqlite) TypeDouble(column schema.ColumnDefinition) string {
	return "double"
}

func (r *Sqlite) TypeEnum(column schema.ColumnDefinition) string {
	return fmt.Sprintf(`varchar check ("%s" in (%s))`, column.GetName(), strings.Join(r.wrap.Quotes(column.GetAllowed()), ", "))
}

func (r *Sqlite) TypeFloat(column schema.ColumnDefinition) string {
	return "float"
}

func (r *Sqlite) TypeInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeJson(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeJsonb(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeLongText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeMediumInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeMediumText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeSmallInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeString(column schema.ColumnDefinition) string {
	return "varchar"
}

func (r *Sqlite) TypeText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeTime(column schema.ColumnDefinition) string {
	return "time"
}

func (r *Sqlite) TypeTimeTz(column schema.ColumnDefinition) string {
	return r.TypeTime(column)
}

func (r *Sqlite) TypeTimestamp(column schema.ColumnDefinition) string {
	if column.GetUseCurrent() {
		column.Default(Expression("CURRENT_TIMESTAMP"))
	}

	return "datetime"
}

func (r *Sqlite) TypeTimestampTz(column schema.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}

func (r *Sqlite) TypeTinyInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeTinyText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) addForeignKeys(commands []*schema.Command) string {
	var sql string

	for _, command := range commands {
		sql += r.getForeignKey(command)
	}

	return sql
}

func (r *Sqlite) addPrimaryKeys(command *schema.Command) string {
	if command == nil {
		return ""
	}

	return fmt.Sprintf(", primary key (%s)", r.wrap.Columnize(command.Columns))
}

func (r *Sqlite) getColumns(blueprint schema.Blueprint) []string {
	var columns []string
	for _, column := range blueprint.GetAddedColumns() {
		columns = append(columns, r.getColumn(blueprint, column))
	}

	return columns
}

func (r *Sqlite) getColumn(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	sql := fmt.Sprintf("%s %s", r.wrap.Column(column.GetName()), getType(r, column))

	for _, modifier := range r.modifiers {
		sql += modifier(blueprint, column)
	}

	return sql
}

func (r *Sqlite) getForeignKey(command *schema.Command) string {
	sql := fmt.Sprintf(", foreign key(%s) references %s(%s)",
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

func (r *Sqlite) pragma(name, value string) string {
	return fmt.Sprintf("pragma %s = %s", name, value)
}
