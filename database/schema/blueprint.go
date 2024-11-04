package schema

import (
	"strings"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/constants"
	"github.com/goravel/framework/support/convert"
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
		Name: constants.CommandCreate,
	})
}

func (r *Blueprint) DropIfExists() {
	r.addCommand(&schema.Command{
		Name: constants.CommandDropIfExists,
	})
}

func (r *Blueprint) GetAddedColumns() []schema.ColumnDefinition {
	var columns []schema.ColumnDefinition
	for _, column := range r.columns {
		columns = append(columns, column)
	}

	return columns
}

func (r *Blueprint) GetCommands() []*schema.Command {
	return r.commands
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

func (r *Blueprint) Primary(column ...string) {
	r.indexCommand(constants.CommandPrimary, column)
}

func (r *Blueprint) SetTable(name string) {
	r.table = name
}

func (r *Blueprint) String(column string, length ...int) schema.ColumnDefinition {
	defaultLength := constants.DefaultStringLength
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
		case constants.CommandAdd:
			statements = append(statements, grammar.CompileAdd(r, command))
		case constants.CommandCreate:
			statements = append(statements, grammar.CompileCreate(r, query))
		case constants.CommandDropIfExists:
			statements = append(statements, grammar.CompileDropIfExists(r))
		case constants.CommandPrimary:
			statements = append(statements, grammar.CompilePrimary(r, command))
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
			if command == constants.CommandComment && column.comment != nil {
				r.addCommand(&schema.Command{
					Column: column,
					Name:   constants.CommandComment,
				})
			}
		}
	}
}

func (r *Blueprint) addColumn(column *ColumnDefinition) {
	r.columns = append(r.columns, column)

	if !r.isCreate() {
		r.addCommand(&schema.Command{
			Name:   constants.CommandAdd,
			Column: column,
		})
	}
}

func (r *Blueprint) addCommand(command *schema.Command) {
	r.commands = append(r.commands, command)
}

func (r *Blueprint) addImpliedCommands(grammar schema.Grammar) {
	r.addAttributeCommands(grammar)
}

func (r *Blueprint) createIndexName(ttype string, columns []string) string {
	table := r.GetTableName()
	index := strings.ToLower(table + "_" + strings.Join(columns, "_") + "_" + ttype)
	index = strings.ReplaceAll(index, "-", "_")

	return strings.ReplaceAll(index, ".", "_")
}

func (r *Blueprint) indexCommand(ttype string, columns []string, config ...schema.IndexConfig) *schema.Command {
	command := &schema.Command{
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
		if command.Name == constants.CommandCreate {
			return true
		}
	}

	return false
}
