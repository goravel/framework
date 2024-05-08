package schema

import (
	schemacontract "github.com/goravel/framework/contracts/database/schema"
)

type ForeignKeyDefinition struct {
	command *schemacontract.Command
}

func NewForeignKeyDefinition(command *schemacontract.Command) schemacontract.ForeignKeyDefinition {
	return &ForeignKeyDefinition{
		command: command,
	}
}

func (f *ForeignKeyDefinition) CascadeOnDelete() schemacontract.ForeignKeyDefinition {
	f.command.OnDelete = "cascade"

	return f
}

func (f *ForeignKeyDefinition) CascadeOnUpdate() schemacontract.ForeignKeyDefinition {
	f.command.OnUpdate = "cascade"

	return f
}

func (f *ForeignKeyDefinition) On(table string) schemacontract.ForeignKeyDefinition {
	f.command.On = table

	return f
}

func (f *ForeignKeyDefinition) Name(name string) schemacontract.ForeignKeyDefinition {
	f.command.Index = name

	return f
}

func (f *ForeignKeyDefinition) NoActionOnDelete() schemacontract.ForeignKeyDefinition {
	f.command.OnDelete = "no action"

	return f
}

func (f *ForeignKeyDefinition) NoActionOnUpdate() schemacontract.ForeignKeyDefinition {
	f.command.OnUpdate = "no action"

	return f
}

func (f *ForeignKeyDefinition) NullOnDelete() schemacontract.ForeignKeyDefinition {
	f.command.OnDelete = "set null"

	return f
}

func (f *ForeignKeyDefinition) References(columns ...string) schemacontract.ForeignKeyDefinition {
	f.command.References = columns

	return f
}

func (f *ForeignKeyDefinition) RestrictOnDelete() schemacontract.ForeignKeyDefinition {
	f.command.OnDelete = "restrict"

	return f
}

func (f *ForeignKeyDefinition) RestrictOnUpdate() schemacontract.ForeignKeyDefinition {
	f.command.OnUpdate = "restrict"

	return f
}

type IndexDefinition struct {
	command *schemacontract.Command
}

func NewIndexDefinition(command *schemacontract.Command) schemacontract.IndexDefinition {
	return &IndexDefinition{
		command: command,
	}
}

func (f *IndexDefinition) Algorithm(algorithm string) schemacontract.IndexDefinition {
	f.command.Algorithm = algorithm

	return f
}

func (f *IndexDefinition) Name(name string) schemacontract.IndexDefinition {
	f.command.Index = name

	return f
}
