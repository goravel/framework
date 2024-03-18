package schema

import (
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	schemacontract "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/convert"
)

const defaultStringLength = 255

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

func (r *Blueprint) BigInteger(column string) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("bigInteger"),
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
		Name:  "tableComment",
	})
}

func (r *Blueprint) Create() {
	r.addCommand(&schemacontract.Command{
		Name: "create",
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

func (r *Blueprint) Decimal(column string, length ...schemacontract.DecimalLength) schemacontract.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("decimal"),
	}
	if len(length) > 0 {
		columnImpl.total = &length[0].Total
		columnImpl.places = &length[0].Places
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

func (r *Blueprint) DropColumn(column string) {
	r.DropColumns([]string{column})
}

func (r *Blueprint) DropColumns(columns []string) {
	r.addCommand(&schemacontract.Command{
		Columns: columns,
		Name:    "dropColumn",
	})
}

func (r *Blueprint) DropForeign(index string) error {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) DropIndex(index string) error {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) DropSoftDeletes() error {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) DropTimestamps() error {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Enum(column string, array []any) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Float(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Foreign(columns []string, name ...string) error {
	//TODO implement me
	panic("implement me")
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

func (r *Blueprint) ID() schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Index(columns []string, name string) error {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Integer(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Json(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Jsonb(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Primary(columns []string, name string) error {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) RenameColumn(from, to string) error {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) RenameIndex(from, to string) error {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) SoftDeletes(column ...string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) SoftDeletesTz(column ...string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Time(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) TimeTz(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Timestamp(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Timestamps() schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) TimestampsTz() schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) TimestampTz(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) ToSql(query ormcontract.Query, grammar schemacontract.Grammar) []string {
	r.addImpliedCommands(grammar)

	var statements []string
	for _, command := range r.commands {
		switch command.Name {
		case "comment":
			statements = append(statements, grammar.CompileComment(r, command))
		case "create":
			statements = append(statements, grammar.CompileCreate(r, query))
		case "tableComment":
			statements = append(statements, grammar.CompileTableComment(r, command.Value))
		}
	}

	return statements
}

func (r *Blueprint) Unique(columns []string, name string) error {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) UnsignedInteger(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) UnsignedBigInteger(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
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
					Name:   "comment",
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

func (r *Blueprint) isCreate() bool {
	for _, command := range r.commands {
		if command.Name == "create" {
			return true
		}
	}

	return false
}

type Command struct {
	Comment string
	Name    string
}
