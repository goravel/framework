package schema

import (
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	schemacontract "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/convert"
)

const defaultStringLength = 255

type Blueprint struct {
	columns  []*ColumnDefinition
	commands []*Command
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
	r.addCommand(&Command{
		Comment: comment,
		Name:    "tableComment",
	})
}

func (r *Blueprint) Create() {
	r.addCommand(&Command{
		Name: "create",
	})
}

func (r *Blueprint) Date(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) DateTime(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) DateTimeTz(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Decimal(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) Double(column string) schemacontract.ColumnDefinition {
	//TODO implement me
	panic("implement me")
}

func (r *Blueprint) DropColumn(column string) error {
	//TODO implement me
	panic("implement me")
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
	var statements []string
	for _, command := range r.commands {
		switch command.Name {
		case "create":
			statements = append(statements, grammar.CompileCreate(r, query))
		case "tableComment":
			statements = append(statements, grammar.CompileTableComment(r, command.Comment))
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

func (r *Blueprint) addCommand(command *Command) {
	r.commands = append(r.commands, command)
}

type Command struct {
	Comment string
	Name    string
}
