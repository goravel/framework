package schema

import (
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/convert"
)

type ForeignKeyDefinition struct {
	command *driver.Command
}

func NewForeignKeyDefinition(command *driver.Command) schema.ForeignKeyDefinition {
	return &ForeignKeyDefinition{
		command: command,
	}
}

func (r *ForeignKeyDefinition) CascadeOnDelete() schema.ForeignKeyDefinition {
	r.command.OnDelete = "cascade"

	return r
}

func (r *ForeignKeyDefinition) CascadeOnUpdate() schema.ForeignKeyDefinition {
	r.command.OnUpdate = "cascade"

	return r
}

func (r *ForeignKeyDefinition) On(table string) schema.ForeignKeyDefinition {
	r.command.On = table

	return r
}

func (r *ForeignKeyDefinition) Name(name string) schema.ForeignKeyDefinition {
	r.command.Index = name

	return r
}

func (r *ForeignKeyDefinition) NoActionOnDelete() schema.ForeignKeyDefinition {
	r.command.OnDelete = "no action"

	return r
}

func (r *ForeignKeyDefinition) NoActionOnUpdate() schema.ForeignKeyDefinition {
	r.command.OnUpdate = "no action"

	return r
}

func (r *ForeignKeyDefinition) NullOnDelete() schema.ForeignKeyDefinition {
	r.command.OnDelete = "set null"

	return r
}

func (r *ForeignKeyDefinition) References(columns ...string) schema.ForeignKeyDefinition {
	r.command.References = columns

	return r
}

func (r *ForeignKeyDefinition) RestrictOnDelete() schema.ForeignKeyDefinition {
	r.command.OnDelete = "restrict"

	return r
}

func (r *ForeignKeyDefinition) RestrictOnUpdate() schema.ForeignKeyDefinition {
	r.command.OnUpdate = "restrict"

	return r
}

type IndexDefinition struct {
	command *driver.Command
}

func NewIndexDefinition(command *driver.Command) schema.IndexDefinition {
	return &IndexDefinition{
		command: command,
	}
}

func (r *IndexDefinition) Algorithm(algorithm string) schema.IndexDefinition {
	r.command.Algorithm = algorithm

	return r
}

func (r *IndexDefinition) Deferrable() schema.IndexDefinition {
	r.command.Deferrable = convert.Pointer(true)

	return r
}

func (r *IndexDefinition) InitiallyImmediate() schema.IndexDefinition {
	r.command.InitiallyImmediate = convert.Pointer(true)

	return r
}

func (r *IndexDefinition) Language(name string) schema.IndexDefinition {
	r.command.Language = name

	return r
}

func (r *IndexDefinition) Name(name string) schema.IndexDefinition {
	r.command.Index = name

	return r
}

type ForeignIdColumnDefinition struct {
	*ColumnDefinition
	blueprint *Blueprint
}

func (r *ForeignIdColumnDefinition) Constrained(table string, column string) schema.ForeignKeyDefinition {
	if column == "" {
		column = "id"
	}

	return r.References(column).On(table)
}

func (r *ForeignIdColumnDefinition) References(column string) schema.ForeignKeyDefinition {
	return r.blueprint.Foreign(r.GetName()).References(column)
}
