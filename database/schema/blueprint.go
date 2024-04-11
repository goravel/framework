package schema

import (
	"strings"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	schemacontract "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/convert"
)

const (
	commandComment      = "comment"
	commandCreate       = "create"
	commandDropColumn   = "dropColumn"
	commandDropForeign  = "dropForeign"
	commandDropIndex    = "dropIndex"
	commandForeign      = "foreign"
	commandIndex        = "index"
	commandPrimary      = "primary"
	commandRenameColumn = "renameColumn"
	commandRenameIndex  = "renameIndex"
	commandTableComment = "tableComment"
	commandUnique       = "unique"
	defaultStringLength = 255
)

type Blueprint struct {
	columns  []*ColumnDefinition
	commands []*schemacontract.Command
	prefix   string
	table    string
}

func NewBlueprint(prefix, table string) *Blueprint {
	return &Blueprint{
		prefix: prefix,
		table:  table,
	}
}

func (r *Blueprint) Boolean(column string) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("boolean"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) BigIncrements(column string) schemacontract.ColumnDefinition {
	return r.UnsignedBigInteger(column, true)
}

func (r *Blueprint) BigInteger(column string, config ...schemacontract.IntegerConfig) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("bigInteger"),
	}

	if len(config) > 0 {
		columnImpl.autoIncrement = &config[0].AutoIncrement
		columnImpl.unsigned = &config[0].Unsigned
	}

	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Binary(column string) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("binary"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Build(query ormcontract.Query, grammar schemacontract.Grammar) error {
	for _, sql := range r.ToSql(query, grammar) {
		if _, err := query.Exec(sql); err != nil {
			return err
		}
	}

	return nil
}

func (r *Blueprint) Char(column string, length ...int) schemacontract.ColumnDefinition {
	defaultLength := defaultStringLength
	if len(length) > 0 {
		defaultLength = length[0]
	}

	columnImpl := &ColumnDefinition{
		length: &defaultLength,
		name:   &column,
		ttype:  convert.Pointer("char"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Comment(comment string) {
	r.addCommand(&schemacontract.Command{
		Value: comment,
		Name:  commandTableComment,
	})
}

func (r *Blueprint) Create() {
	r.addCommand(&schemacontract.Command{
		Name: commandCreate,
	})
}

func (r *Blueprint) Date(column string) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("date"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) DateTime(column string, precision ...int) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("dateTime"),
	}
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) DateTimeTz(column string, precision ...int) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("dateTimeTz"),
	}
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Decimal(column string, config ...schemacontract.DecimalConfig) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("decimal"),
	}
	if len(config) > 0 {
		columnImpl.total = &config[0].Total
		columnImpl.places = &config[0].Places
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Double(column string) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("double"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) DropColumn(column ...string) {
	if len(column) == 0 {
		panic("You must specify at least one column to drop.")
	}

	r.addCommand(&schemacontract.Command{
		Columns: column,
		Name:    commandDropColumn,
	})
}

func (r *Blueprint) DropForeign(columns []string) {
	r.indexCommand(commandDropForeign, columns, schemacontract.IndexConfig{
		Name: r.createIndexName("foreign", columns),
	})
}

func (r *Blueprint) DropForeignByName(name string) {
	r.indexCommand(commandDropForeign, nil, schemacontract.IndexConfig{
		Name: name,
	})
}

func (r *Blueprint) DropIndex(columns []string) {
	r.indexCommand(commandDropIndex, columns, schemacontract.IndexConfig{
		Name: r.createIndexName("index", columns),
	})
}

func (r *Blueprint) DropIndexByName(name string) {
	r.indexCommand(commandDropIndex, nil, schemacontract.IndexConfig{
		Name: name,
	})
}

func (r *Blueprint) DropSoftDeletes(column ...string) {
	c := "deleted_at"
	if len(column) > 0 {
		c = column[0]
	}

	r.DropColumn(c)
}

func (r *Blueprint) DropSoftDeletesTz(column ...string) {
	r.DropSoftDeletes(column...)
}

func (r *Blueprint) DropTimestamps() {
	r.DropColumn("created_at", "updated_at")
}

func (r *Blueprint) DropTimestampsTz() {
	r.DropTimestamps()
}

func (r *Blueprint) Enum(column string, allowed []string) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		allowed: allowed,
		name:    &column,
		ttype:   convert.Pointer("enum"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Float(column string, precision ...int) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:      &column,
		precision: convert.Pointer(53),
		ttype:     convert.Pointer("float"),
	}
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Foreign(columns []string, name ...string) schemacontract.ForeignKeyDefinition {
	var command *schemacontract.Command
	if len(name) == 0 {
		command = r.indexCommand(commandForeign, columns)
	} else {
		command = r.indexCommand(commandForeign, columns, schemacontract.IndexConfig{
			Name: name[0],
		})
	}

	return NewForeignKeyDefinition(command)
}

func (r *Blueprint) GetAddedColumns() []schemacontract.ColumnDefinition {
	var columns []schemacontract.ColumnDefinition
	for _, column := range r.columns {
		if column.change == nil || !*column.change {
			columns = append(columns, column)
		}
	}

	return columns
}

func (r *Blueprint) GetChangedColumns() []schemacontract.ColumnDefinition {
	var columns []schemacontract.ColumnDefinition
	for _, column := range r.columns {
		if column.change != nil && *column.change {
			columns = append(columns, column)
		}
	}

	return columns
}

func (r *Blueprint) GetTableName() string {
	return r.prefix + r.table
}

func (r *Blueprint) HasCommand(command string) bool {
	for _, c := range r.commands {
		if c.Name == command {
			return true
		}
	}

	return false
}

func (r *Blueprint) ID(column ...string) schemacontract.ColumnDefinition {
	if len(column) > 0 {
		return r.BigIncrements(column[0])
	}

	return r.BigIncrements("id")
}

func (r *Blueprint) Index(columns []string, config ...schemacontract.IndexConfig) {
	r.indexCommand(commandIndex, columns, config...)
}

func (r *Blueprint) Integer(column string, config ...schemacontract.IntegerConfig) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("integer"),
	}

	if len(config) > 0 {
		columnImpl.autoIncrement = &config[0].AutoIncrement
		columnImpl.unsigned = &config[0].Unsigned
	}

	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Json(column string) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("json"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Jsonb(column string) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("jsonb"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Primary(columns []string) {
	r.indexCommand(commandPrimary, columns)
}

func (r *Blueprint) RenameColumn(from, to string) {
	r.addCommand(&schemacontract.Command{
		From: from,
		Name: commandRenameColumn,
		To:   to,
	})
}

func (r *Blueprint) RenameIndex(from, to string) {
	r.addCommand(&schemacontract.Command{
		From: from,
		Name: commandRenameIndex,
		To:   to,
	})
}

func (r *Blueprint) SoftDeletes(column ...string) schemacontract.ColumnDefinition {
	c := "deleted_at"
	if len(column) > 0 {
		c = column[0]
	}

	return r.Timestamp(c).Nullable()
}

func (r *Blueprint) SoftDeletesTz(column ...string) schemacontract.ColumnDefinition {
	c := "deleted_at"
	if len(column) > 0 {
		c = column[0]
	}

	return r.TimestampTz(c).Nullable()
}

func (r *Blueprint) String(column string, length ...int) schemacontract.ColumnDefinition {
	defaultLength := defaultStringLength
	if len(length) > 0 {
		defaultLength = length[0]
	}

	columnImpl := &ColumnDefinition{
		length: &defaultLength,
		name:   &column,
		ttype:  convert.Pointer("string"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Text(column string) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("text"),
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Time(column string, precision ...int) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("time"),
	}
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) TimeTz(column string, precision ...int) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("timeTz"),
	}
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Timestamp(column string, precision ...int) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("timestamp"),
	}
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Timestamps(precision ...int) {
	r.Timestamp("created_at", precision...).Nullable()
	r.Timestamp("updated_at", precision...).Nullable()
}

func (r *Blueprint) TimestampsTz(precision ...int) {
	r.TimestampTz("created_at", precision...).Nullable()
	r.TimestampTz("updated_at", precision...).Nullable()
}

func (r *Blueprint) TimestampTz(column string, precision ...int) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("timestampTz"),
	}
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}
	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) ToSql(query ormcontract.Query, grammar schemacontract.Grammar) []string {
	r.addImpliedCommands(grammar)

	var statements []string
	for _, command := range r.commands {
		switch command.Name {
		case commandComment:
			statements = append(statements, grammar.CompileComment(r, command))
		case commandCreate:
			statements = append(statements, grammar.CompileCreate(r, query))
		case commandDropColumn:
			statements = append(statements, grammar.CompileDropColumn(r, command))
		case commandDropForeign:
			statements = append(statements, grammar.CompileDropForeign(r, command.Index))
		case commandDropIndex:
			statements = append(statements, grammar.CompileDropIndex(r, command.Index))
		case commandForeign:
			statements = append(statements, grammar.CompileForeign(r, command))
		case commandIndex:
			statements = append(statements, grammar.CompileIndex(r, command))
		case commandPrimary:
			statements = append(statements, grammar.CompilePrimary(r, command.Columns))
		case commandRenameColumn:
			statements = append(statements, grammar.CompileRenameColumn(r, command.From, command.To))
		case commandRenameIndex:
			statements = append(statements, grammar.CompileRenameIndex(r, command.From, command.To))
		case commandTableComment:
			statements = append(statements, grammar.CompileTableComment(r, command.Value))
		case commandUnique:
			statements = append(statements, grammar.CompileUnique(r, command))
		}
	}

	return statements
}

func (r *Blueprint) Unique(columns []string) {
	r.indexCommand(commandUnique, columns)
}

func (r *Blueprint) UnsignedInteger(column string, autoIncrement ...bool) schemacontract.ColumnDefinition {
	return r.Integer(column, schemacontract.IntegerConfig{
		AutoIncrement: len(autoIncrement) > 0 && autoIncrement[0],
		Unsigned:      true,
	})
}

func (r *Blueprint) UnsignedBigInteger(column string, autoIncrement ...bool) schemacontract.ColumnDefinition {
	return r.BigInteger(column, schemacontract.IntegerConfig{
		AutoIncrement: len(autoIncrement) > 0 && autoIncrement[0],
		Unsigned:      true,
	})
}

func (r *Blueprint) addColumn(column *ColumnDefinition) {
	r.columns = append(r.columns, column)
}

func (r *Blueprint) addCommand(command *schemacontract.Command) {
	r.commands = append(r.commands, command)
}

func (r *Blueprint) addAttributeCommands(grammar schemacontract.Grammar) {
	attributeCommands := grammar.GetAttributeCommands()
	for _, column := range r.columns {
		for _, command := range attributeCommands {
			if command == "comment" && column.comment != nil {
				r.addCommand(&schemacontract.Command{
					Column: column,
					Name:   commandComment,
				})
			}
		}
	}
}

func (r *Blueprint) addImpliedCommands(grammar schemacontract.Grammar) {
	var commands []*schemacontract.Command
	if len(r.GetAddedColumns()) > 0 && !r.isCreate() {
		commands = append(commands, &schemacontract.Command{
			Name: "add",
		})
	}
	if len(r.GetChangedColumns()) > 0 && !r.isCreate() {
		commands = append(commands, &schemacontract.Command{
			Name: "change",
		})
	}
	if len(commands) > 0 {
		r.commands = append(commands, r.commands...)
	}
	r.addAttributeCommands(grammar)
}

func (r *Blueprint) createIndexName(ttype string, columns []string) string {
	table := r.GetTableName()
	index := strings.ToLower(table + "_" + strings.Join(columns, "_") + "_" + ttype)
	index = strings.ReplaceAll(index, "-", "_")

	return strings.ReplaceAll(index, ".", "_")
}

func (r *Blueprint) indexCommand(ttype string, columns []string, config ...schemacontract.IndexConfig) *schemacontract.Command {
	command := &schemacontract.Command{
		Columns: columns,
		Name:    ttype,
	}

	if len(config) > 0 {
		command.Algorithm = config[0].Algorithm
		command.Index = config[0].Name
	} else {
		command.Index = r.createIndexName(ttype, columns)
	}

	r.addCommand(command)

	return command
}

func (r *Blueprint) isCreate() bool {
	for _, command := range r.commands {
		if command.Name == commandCreate {
			return true
		}
	}

	return false
}

type Command struct {
	Comment string
	Name    string
}
