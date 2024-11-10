package schema

import (
	"github.com/goravel/framework/contracts/database/schema"
)

type ForeignKeyDefinition struct {
	command *schema.Command
}

func NewForeignKeyDefinition(command *schema.Command) schema.ForeignKeyDefinition {
	return &ForeignKeyDefinition{
		command: command,
	}
}

func (f *ForeignKeyDefinition) CascadeOnDelete() schema.ForeignKeyDefinition {
	f.command.OnDelete = "cascade"

	return f
}

func (f *ForeignKeyDefinition) CascadeOnUpdate() schema.ForeignKeyDefinition {
	f.command.OnUpdate = "cascade"

	return f
}

func (f *ForeignKeyDefinition) On(table string) schema.ForeignKeyDefinition {
	f.command.On = table

	return f
}

func (f *ForeignKeyDefinition) Name(name string) schema.ForeignKeyDefinition {
	f.command.Index = name

	return f
}

func (f *ForeignKeyDefinition) NoActionOnDelete() schema.ForeignKeyDefinition {
	f.command.OnDelete = "no action"

	return f
}

func (f *ForeignKeyDefinition) NoActionOnUpdate() schema.ForeignKeyDefinition {
	f.command.OnUpdate = "no action"

	return f
}

func (f *ForeignKeyDefinition) NullOnDelete() schema.ForeignKeyDefinition {
	f.command.OnDelete = "set null"

	return f
}

func (f *ForeignKeyDefinition) References(columns ...string) schema.ForeignKeyDefinition {
	f.command.References = columns

	return f
}

func (f *ForeignKeyDefinition) RestrictOnDelete() schema.ForeignKeyDefinition {
	f.command.OnDelete = "restrict"

	return f
}

func (f *ForeignKeyDefinition) RestrictOnUpdate() schema.ForeignKeyDefinition {
	f.command.OnUpdate = "restrict"

	return f
}

type IndexDefinition struct {
	command *schema.Command
}

func NewIndexDefinition(command *schema.Command) schema.IndexDefinition {
	return &IndexDefinition{
		command: command,
	}
}

func (f *IndexDefinition) Algorithm(algorithm string) schema.IndexDefinition {
	f.command.Algorithm = algorithm

	return f
}

func (f *IndexDefinition) Name(name string) schema.IndexDefinition {
	f.command.Index = name

	return f
}
