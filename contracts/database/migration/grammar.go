package migration

import (
	"github.com/goravel/framework/contracts/database/orm"
)

type Grammar interface {
	// CompileCreate Compile a create table command.
	CompileCreate(blueprint Blueprint, query orm.Query) string
	// GetAttributeCommands Get the commands for the schema build.
	GetAttributeCommands() []string
	// GetModifiers Get the column modifiers.
	GetModifiers() []func(Blueprint, ColumnDefinition) string
}
