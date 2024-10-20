package schema

import (
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/convert"
)

const (
	commandAdd          = "add"
	commandChange       = "change"
	commandComment      = "comment"
	commandCreate       = "create"
	commandDropIfExists = "dropIfExists"
	defaultStringLength = 255
)

type Blueprint struct {
	columns  []*ColumnDefinition
	commands []*schema.Command
	prefix   string
	table    string
}

func NewBlueprint(prefix, table string) *Blueprint {
	return &Blueprint{
		prefix: prefix,
		table:  table,
	}
}

func (r *Blueprint) BigIncrements(column string) schema.ColumnDefinition {
	return r.UnsignedBigInteger(column).AutoIncrement()
}

func (r *Blueprint) BigInteger(column string) schema.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("bigInteger"),
	}

	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) Build(query ormcontract.Query, grammar schema.Grammar) error {
	for _, sql := range r.ToSql(query, grammar) {
		if _, err := query.Exec(sql); err != nil {
			return err
		}
	}

	return nil
}

func (r *Blueprint) Create() {
	r.addCommand(&schema.Command{
		Name: commandCreate,
	})
}

func (r *Blueprint) DropIfExists() {
	r.addCommand(&schema.Command{
		Name: commandDropIfExists,
	})
}

func (r *Blueprint) GetAddedColumns() []schema.ColumnDefinition {
	var columns []schema.ColumnDefinition
	for _, column := range r.columns {
		if column.change == nil || !*column.change {
			columns = append(columns, column)
		}
	}

	return columns
}

func (r *Blueprint) GetChangedColumns() []schema.ColumnDefinition {
	var columns []schema.ColumnDefinition
	for _, column := range r.columns {
		if column.change != nil && *column.change {
			columns = append(columns, column)
		}
	}

	return columns
}

func (r *Blueprint) GetTableName() string {
	// TODO Add schema for Postgres
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

func (r *Blueprint) ID(column ...string) schema.ColumnDefinition {
	if len(column) > 0 {
		return r.BigIncrements(column[0])
	}

	return r.BigIncrements("id")
}

func (r *Blueprint) Integer(column string) schema.ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &column,
		ttype: convert.Pointer("integer"),
	}

	r.addColumn(columnImpl)

	return columnImpl
}

func (r *Blueprint) SetTable(name string) {
	r.table = name
}

func (r *Blueprint) String(column string, length ...int) schema.ColumnDefinition {
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

func (r *Blueprint) ToSql(query ormcontract.Query, grammar schema.Grammar) []string {
	r.addImpliedCommands(grammar)

	var statements []string
	for _, command := range r.commands {
		switch command.Name {
		case commandAdd:
			statements = append(statements, grammar.CompileAdd(r))
		case commandChange:
			statements = append(statements, grammar.CompileChange(r))
		case commandCreate:
			statements = append(statements, grammar.CompileCreate(r, query))
		case commandDropIfExists:
			statements = append(statements, grammar.CompileDropIfExists(r))
		}
	}

	return statements
}

func (r *Blueprint) UnsignedBigInteger(column string) schema.ColumnDefinition {
	return r.BigInteger(column).Unsigned()
}

func (r *Blueprint) addAttributeCommands(grammar schema.Grammar) {
	attributeCommands := grammar.GetAttributeCommands()
	for _, column := range r.columns {
		for _, command := range attributeCommands {
			if command == "comment" && column.comment != nil {
				r.addCommand(&schema.Command{
					Column: column,
					Name:   commandComment,
				})
			}
		}
	}
}

func (r *Blueprint) addColumn(column *ColumnDefinition) {
	r.columns = append(r.columns, column)
}

func (r *Blueprint) addCommand(command *schema.Command) {
	r.commands = append(r.commands, command)
}

func (r *Blueprint) addImpliedCommands(grammar schema.Grammar) {
	var commands []*schema.Command
	if len(r.GetAddedColumns()) > 0 && !r.isCreate() {
		commands = append(commands, &schema.Command{
			Name: commandAdd,
		})
	}
	if len(r.GetChangedColumns()) > 0 && !r.isCreate() {
		commands = append(commands, &schema.Command{
			Name: commandChange,
		})
	}
	if len(commands) > 0 {
		r.commands = append(commands, r.commands...)
	}
	r.addAttributeCommands(grammar)
}

func (r *Blueprint) isCreate() bool {
	for _, command := range r.commands {
		if command.Name == commandCreate {
			return true
		}
	}

	return false
}
